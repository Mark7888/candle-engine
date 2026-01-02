package main

import (
	"testcharting/cmd/webserver/sample"
	"testcharting/internal/chartmanager"
	"testcharting/internal/webserver"
)

func main() {
	chartProject := sample.GetSampleChartProject()
	chartmanager.GetChartManager().AddChartProjects(chartProject)

	webserver.InitServer()
}
