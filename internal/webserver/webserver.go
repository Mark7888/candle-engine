package webserver

import (
	"fmt"
	"net/http"
	"sync"

	"testcharting/internal/webserver/chartmanager"
	"testcharting/internal/webserver/routes"
	"testcharting/pkg/charting/chartproject"
)

var chartManager *chartmanager.ChartManager
var chartManagerMutex sync.RWMutex

func GetChartManager() *chartmanager.ChartManager {
	chartManagerMutex.Lock()
	defer chartManagerMutex.Unlock()

	if chartManager == nil {
		chartManager = chartmanager.NewChartManager()
	}
	return chartManager
}

func InitServer(chartProjects ...*chartproject.ChartProject) {
	GetChartManager().AddChartProjects(chartProjects...)

	routes.LoadRoutes()

	// render static content from assets folder
	// http.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), http.FileServer(http.Dir("./assets/static"))))

	// sse_handler.StartServer("/events/")

	fmt.Println("Webserver listening on localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}
