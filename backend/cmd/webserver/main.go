package main

import (
	"candle-engine/cmd/webserver/sample"
	"candle-engine/internal/chartmanager"
	"candle-engine/internal/webserver"
)

func main() {
	chartProject := sample.GetSampleChartProject()
	chartProject2 := sample.GetSampleChartProject2()
	chartmanager.GetChartManager().AddChartProjects(chartProject, chartProject2)

	webserver.InitServer()
}
