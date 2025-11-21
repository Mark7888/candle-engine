package chartproject

import (
	"fmt"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
	"time"
)

type ChartProject struct {
	title string
	// TODO - is map the best structure here?
	loadedSources map[string]sources.Source // key is source name
	selectedPairs []string                  // e.g., "Coinbase:BTC-USD"
	indicators    []indicators.PlotIndicator

	interval string

	data []sources.ChartEntry
}

func NewChartProject(title string) *ChartProject {
	return &ChartProject{
		title: title,
	}
}

func (c *ChartProject) AddSource(source sources.Source) {
	if c.loadedSources == nil {
		c.loadedSources = make(map[string]sources.Source)
	}
	c.loadedSources[source.GetName()] = source
}

func (c *ChartProject) AddTradingPair(pair string) {
	for _, p := range c.selectedPairs {
		if p == pair {
			return
		}
	}

	c.selectedPairs = append(c.selectedPairs, pair)
}

func (c *ChartProject) SetInterval(interval string) {
	c.interval = interval
}

func (c *ChartProject) AddIndicator(indicator indicators.PlotIndicator) {
	c.indicators = append(c.indicators, indicator)
}

func (c *ChartProject) FetchData() error {
	// TODO - implement logic to select pair and actual time range
	timeRange := sources.TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}
	data, err := c.loadedSources["Coinbase"].GetData(c.selectedPairs[0], c.interval, timeRange)
	if err != nil {
		return err
	}
	c.data = data
	return nil
}

// TODO - implement rendering logic
func (c *ChartProject) Render() {
	if len(c.data) == 0 {
		fmt.Println("No data fetched yet, fetching now...")
		c.FetchData()
	}

	fmt.Println("Rendering ChartProject with", len(c.indicators), "indicators and", len(c.data), "data points.")
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

func (c *ChartProject) CalculateIndicators(chartConfig indicators.ChartConfig) map[string][]indicators.PlotPoint {
	plotData := make(map[string][]indicators.PlotPoint)

	for i := range c.data {
		multiSourceEntries := sources.NewMultiSourceEntries()
		multiSourceEntries.AddEntries(c.selectedPairs[0], c.data[:i+1])
		// Loop through indicators and get their plot points
		for _, indicator := range c.indicators {
			point, err := indicator.HandleEntry(plotData[indicator.GetName()], multiSourceEntries, chartConfig)
			if err != nil {
				point = indicators.GenerateInvalidPlotPoint()
			}

			plotData[indicator.GetName()] = append(plotData[indicator.GetName()], point)
		}
	}

	return plotData
}
