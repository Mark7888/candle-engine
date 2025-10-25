package main

import (
	"testcharting/pkg/charting"
)

func main() {
	source := CreateMySource()
	indicator1 := CreateTestIndicator1()
	indicator2 := CreateTestIndicator2()

	chart := charting.NewCoolChart(source)
	chart.AddIndicator(indicator1)
	chart.AddIndicator(indicator2)

	chart.Render()
}
