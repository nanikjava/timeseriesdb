package pkg

import (
	"fmt"
	"github.com/shirou/gopsutil/mem"
	"html/template"
	"log"
	"math"
	"os"
)

// IndexTemplate is template to show simple graph to map the series data from the db
var IndexTemplate = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en" style="height: 100%">
<head>
  <meta charset="utf-8">
</head>
<body style="height: 100%; margin: 0">
  <div id="container" style="height: 100%"></div>

  <script type="text/javascript" src="https://fastly.jsdelivr.net/npm/echarts@5.5.1/dist/echarts.min.js"></script>

  <script type="text/javascript">
    var dom = document.getElementById('container');
    var myChart = echarts.init(dom, null, {
      renderer: 'canvas',
      useDirtyRect: false
    });
    var app = {};
    
    var option;

    option = {
      xAxis: {
        type: 'category',
        data: []
      },
      yAxis: {
        type: 'value'
      },
      series: [
        {
          data: {{ .Data }} ,
          type: 'line',
          smooth: true
        }
      ]
    };
    if (option && typeof option === 'object') {
      myChart.setOption(option);
    }

    window.addEventListener('resize', myChart.resize);
  </script>
</body>
</html>
`))

// BytesToUnit converts a float64 value (representing bytes) to a human-readable unit (KB, MB, GB, TB)
func BytesToUnit(value uint64) float64 {
	num := float64(value)
	switch {
	case num < 1024:
	case num < math.Pow(1024, 2): // Less than 1 MB
		num = num / 1024
	case num < math.Pow(1024, 3): // Less than 1 GB
		num = num / math.Pow(1024, 2)
	case num < math.Pow(1024, 4): // Less than 1 TB
		num = num / math.Pow(1024, 3)
	default:
		num = num / math.Pow(1024, 4)
	}

	return num
}

func GetBasePath(folder string) string {
	var path string
	appPath, _ := os.Getwd()
	if appPath == "/" {
		path = fmt.Sprintf("%s%s", appPath, folder)
	} else {
		path = fmt.Sprintf("%s/%s", appPath, folder)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	return path
}

func GetMemoryStatistics() uint64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Panicf("Error fetching memory info: %v", err)
	}

	return memInfo.Dirty
}
