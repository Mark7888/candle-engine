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

	/*
		// Example code to test data source directly
		source := coinbase.NewCoinbaseSource()
		pairs := source.GetTradingPairs()

		fmt.Println("Source Name:", source.GetName())
		fmt.Println("Trading Pairs:")
		for _, pair := range pairs {
			if pair.Status != "online" {
				continue
			}
			fmt.Printf("- %s (%s/%s) Status: %s\n", pair.Name, pair.Base, pair.Quote, pair.Status)
		}

		startTime, _ := time.Parse(time.RFC3339, "2024-06-01T00:00:00Z")
		endTime, _ := time.Parse(time.RFC3339, "2024-06-02T00:00:00Z")
		data, err := source.GetData("BTC-USD", "1h", startTime, endTime)
		if err != nil {
			panic(err)
		}

		for _, entry := range data {
			println(entry.Timestamp.String(), entry.Open, entry.High, entry.Low, entry.Close, entry.Volume)
		}
	*/
}
