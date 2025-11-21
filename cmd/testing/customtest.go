package main

import (
	"fmt"
	"testcharting/pkg/charting/chartproject"
)

func customTest() {
	fmt.Println()
	fmt.Println()

	// Original chart rendering code
	source := CreateMySource()
	indicator1 := CreateTestIndicator1()
	indicator2 := CreateTestIndicator2()

	chart := chartproject.NewChartProject("First Chart Project")
	chart.AddSources(source)
	chart.AddTradingPairs("Coinbase:BTC-USD")
	chart.SetInterval("1h")
	chart.AddIndicator(indicator1)
	chart.AddIndicator(indicator2)

	chart.Render()
}
