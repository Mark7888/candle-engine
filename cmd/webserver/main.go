package main

import (
	"fmt"
	"testcharting/cmd/webserver/sample"
)

func main() {
	chartProject := sample.GetSampleChartProject()

	render, err := chartProject.GetRender()
	if err != nil {
		fmt.Println("Error getting renderable chart project:", err)
		return
	}

	_ = render
	// fmt.Printf("Renderable Chart Project: %+v\n", render)

	// webserver.InitServer(chartProject)
}
