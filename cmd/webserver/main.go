package main

import (
	"testcharting/cmd/webserver/sample"
	"testcharting/internal/chartmanager"
	"testcharting/internal/webserver"
)

func main() {
	chartProject := sample.GetSampleChartProject()
	chartProject2 := sample.GetSampleChartProject2()
	chartmanager.GetChartManager().AddChartProjects(chartProject, chartProject2)

	webserver.InitServer()
}
