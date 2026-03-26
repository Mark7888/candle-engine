package main

import (
	"candle-engine/cmd/webserver/sample"
	"candle-engine/pkg/chartmanager"
	"candle-engine/pkg/webserver"
)

func main() {
	chartProject := sample.GetSampleChartProject()
	chartProject2 := sample.GetSampleChartProject2()
	chartProject3 := sample.GetDemoChartProject()
	chartmanager.GetChartManager().AddChartProjects(chartProject, chartProject2, chartProject3)

	webserver.InitServer()
}
