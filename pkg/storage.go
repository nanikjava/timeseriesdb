package pkg

import (
	"github.com/nakabonne/tstorage"
	"log"
	"sync"
	"time"
)

const Memory_Metric_Name = "dirty_memory"

type TimeSeriesWrapper struct {
	Storage tstorage.Storage
	closed  bool
	mu      sync.Mutex
}

func (s *TimeSeriesWrapper) InsertRows(rows []tstorage.Row) error {
	return s.Storage.InsertRows(rows)
}

func (s *TimeSeriesWrapper) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil // or return a specific error indicating it's already closed
	}

	s.closed = true
	return s.Storage.Close()
}

func (s *TimeSeriesWrapper) GetFromStorage(label tstorage.Label) []float64 {
	currentTime := time.Now()
	timestamp := currentTime.Unix()

	now := time.Now()
	// Subtract 2 days from the current time
	twoDaysAgo := now.Add(-48 * time.Hour)
	// Convert to Unix timestamp (seconds since epoch)
	twoDaysAgounixTimestamp := twoDaysAgo.Unix()

	points, err := s.Storage.Select(Memory_Metric_Name, []tstorage.Label{label}, twoDaysAgounixTimestamp, timestamp)

	if err != nil {
		log.Fatalln("Error reading metrics")
	}

	var arrayData = make([]float64, len(points))

	for i, point := range points {
		arrayData[i] = point.Value
	}

	log.Println("Total read ", len(arrayData))
	return arrayData
}
