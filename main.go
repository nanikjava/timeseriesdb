package main

import (
	"fmt"
	"github.com/nakabonne/tstorage"
	"log"
	"net/http"
	"strings"
	"time"
	"tstorage_example/pkg"
)

var (
	sw *pkg.TimeSeriesWrapper
)

func getDataFromStorage(label tstorage.Label) []float64 {
	f := sw.GetFromStorage(label)

	data := make([]float64, len(f))
	for i := 0; i < len(f); i++ {
		data[i] = f[i]
	}

	return data
}

func indexHandler(label tstorage.Label, w http.ResponseWriter) {
	data := getDataFromStorage(label)
	var Data strings.Builder

	for i := 0; i < len(data); i++ {
		Data.WriteString(fmt.Sprintf("%f", data[i]))
		if i+1 < len(data) {
			Data.WriteString(",")
		}
	}

	tmplData := struct {
		Data []float64
	}{
		Data: data,
	}

	if err := pkg.IndexTemplate.Execute(w, tmplData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func collectAndStoreMemoryData(label tstorage.Label) {
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Fatalf("error loading location: %w", err)
	}
	var rows []tstorage.Row
	var ctr = 0
	for {
		for i := 1; i < 10; i++ {
			r := getMemoryValue(label, time.Now().In(location).Unix())
			log.Println(r)
			rows = append(rows, r...)
			time.Sleep(1000 * time.Millisecond)
		}
		if err := sw.InsertRows(rows); err != nil {
			log.Fatalf("error storing service metrics: %w", err)
		}
		rows = nil
		rows = []tstorage.Row{}
		ctr++
		if ctr >= 2 {
			time.Sleep(1000 * time.Hour)
		}

	}
}

func getMemoryValue(label tstorage.Label, timestamp int64) []tstorage.Row {
	return []tstorage.Row{
		{
			Metric:    pkg.Memory_Metric_Name,
			DataPoint: tstorage.DataPoint{Timestamp: timestamp, Value: pkg.BytesToUnit(pkg.GetMemoryStatistics())},
			Labels:    []tstorage.Label{label},
		},
	}
}

func getPartitionDuration() time.Duration {
	// function is to return duration, but because time unit (time.XXX) it's too long to wait
	// so use like a whole number that works like a counter check. Using this hack way is good
	// to demonstrate the creation of partition file. Internally the calculation will work like
	// counter. Once it reach same or bigger than value specified (for example in this case: 5) it will
	// automatically create a new partition file with the current data in the `wal` folder. The following
	// is an example of how the partition files will look like:
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:14 p-1728299657-1728299665/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:14 p-1728299666-1728299674/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299675-1728299683/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299684-1728299692/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299693-1728299701/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299702-1728299710/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299711-1728299719/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299720-1728299728/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:15 p-1728299729-1728299737/
	//	drwxrwxr-x  2 nanik nanik 4096 Oct  7 22:16 p-1728299738-1728299746/

	return time.Duration(5)
}

func main() {
	label := tstorage.Label{Name: "host", Value: "nanik"}

	basePath := pkg.GetBasePath("tmp")
	tstorageInstance, err := tstorage.NewStorage(
		tstorage.WithDataPath(basePath+"/data"),
		tstorage.WithWALBufferedSize(5),
		tstorage.WithPartitionDuration(getPartitionDuration()),
		tstorage.WithRetention(time.Duration(7)*24*time.Hour),
	)
	if err != nil {
		log.Panicf("Error initializing sw: %v\n", err)
	}
	sw = &pkg.TimeSeriesWrapper{Storage: tstorageInstance}

	go func() {
		for {
			collectAndStoreMemoryData(label)
		}
	}()

	// Serve the /index route
	http.HandleFunc("/index", func(rw http.ResponseWriter, rq *http.Request) {
		indexHandler(label, rw)
	})

	// Start the server on port 8080
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
