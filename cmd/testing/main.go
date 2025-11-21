package main

import (
	"fmt"
	"testcharting/pkg/charting"
)

func main() {
	// Test the caching functionality
	testCaching()

	fmt.Println()
	fmt.Println()

	// Original chart rendering code
	source := CreateMySource()
	indicator1 := CreateTestIndicator1()
	indicator2 := CreateTestIndicator2()

	chart := charting.NewCoolChart()
	chart.AddSource(source)
	chart.AddTradingPair("Coinbase:BTC-USD")
	chart.SetInterval("1h")
	chart.AddIndicator(indicator1)
	chart.AddIndicator(indicator2)

	chart.Render()
}
