package charting

import (
	"fmt"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
	"time"
)

type CoolChart struct {
	// TODO - enable multiple sources?
	source     sources.Source
	indicators []indicators.PlotIndicator

	// TODO - support selecting multiple pairs
	pair string // e.g., "Coinbase:BTC-USD"

	interval string
	data     []sources.ChartEntry
}

func NewCoolChart(source sources.Source) *CoolChart {
	return &CoolChart{
		source: source,
	}
}

func (c *CoolChart) AddIndicator(indicator indicators.PlotIndicator) {
	c.indicators = append(c.indicators, indicator)
}

func (c *CoolChart) FetchData() error {
	// TODO - implement logic to select pair and actual time range
	timeRange := sources.TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	data, err := c.source.GetData(c.pair, c.interval, timeRange)
	if err != nil {
		return err
	}
	c.data = data
	return nil
}

// TODO - implement rendering logic
func (c *CoolChart) Render() {
	if len(c.data) == 0 {
		fmt.Println("No data fetched yet, fetching now...")
		c.FetchData()
	}

	fmt.Println("Rendering CoolChart with", len(c.indicators), "indicators and", len(c.data), "data points.")
	// Implementation to render the chart

	chartConfig := indicators.ChartConfig{
		Interval: c.interval,
	}

	plotData := c.CalculateIndicators(chartConfig)

	// TODO - this is just temporary to show the calculated plot data
	for name, points := range plotData {
		fmt.Println("Indicator:", name)
		for i, point := range points {
			fmt.Println(" Point", i, "Value:", point.Value, "Color:", point.Color)
		}
	}

	// Final rendering logic
}

func (c *CoolChart) CalculateIndicators(chartConfig indicators.ChartConfig) map[string][]indicators.PlotPoint {
	plotData := make(map[string][]indicators.PlotPoint)

	for i := range c.data {
		// Loop through indicators and get their plot points
		for _, indicator := range c.indicators {
			point, err := indicator.HandleEntry(plotData[indicator.GetName()], c.data[:i+1], chartConfig)
			if err != nil {
				point = indicators.GenerateInvalidPlotPoint()
			}

			plotData[indicator.GetName()] = append(plotData[indicator.GetName()], point)
		}
	}

	return plotData
}
