package charting

import (
	"fmt"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

type CoolChart struct {
	source     sources.Source
	indicators []indicators.BasicIndicator

	interval string
	data     []sources.ChartEntry
}

func NewCoolChart(source sources.Source) *CoolChart {
	return &CoolChart{
		source: source,
	}
}

func (c *CoolChart) AddIndicator(indicator indicators.BasicIndicator) {
	c.indicators = append(c.indicators, indicator)
}

func (c *CoolChart) FetchData() error {
	data, err := c.source.GetData(c.interval)
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
			lastPoint := indicators.PlotPoint{}
			if len(plotData[indicator.GetName()]) > 0 {
				lastPoint = plotData[indicator.GetName()][i-1]
			}

			point, err := indicator.HandleEntry(lastPoint, c.data[:i+1], chartConfig)
			if err != nil {
				point = indicators.GenerateInvalidPlotPoint()
			}

			plotData[indicator.GetName()] = append(plotData[indicator.GetName()], point)
		}
	}

	return plotData
}
