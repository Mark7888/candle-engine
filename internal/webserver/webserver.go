package webserver

import (
	"fmt"
	"net/http"

	"testcharting/internal/webserver/routes"
	sse_handler "testcharting/internal/webserver/sse"
	"testcharting/pkg/charting/chartproject"
)

func InitServer(chartProjects ...*chartproject.ChartProject) {
	routes.LoadRoutes()

	// render static content from assets folder
	// http.Handle("/static/", http.StripPrefix(strings.TrimRight("/static/", "/"), http.FileServer(http.Dir("./assets/static"))))

	sse_handler.StartServer("/events/")

	fmt.Println("Webserver listening on localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}
