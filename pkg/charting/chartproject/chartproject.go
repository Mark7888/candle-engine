package chartproject

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

type ChartProject struct {
	title         string
	sourceManager *ChartSourceManager
	indicators    []indicators.BaseIndicator

	timeFrame string
}

func NewChartProject(title string) *ChartProject {
	return &ChartProject{
		title:         title,
		sourceManager: NewChartSourceManager(),
		indicators:    []indicators.BaseIndicator{},
		timeFrame:     "1d",
	}
}

func (c *ChartProject) AddSources(sources_ ...sources.Source) {
	c.sourceManager.AddSources(sources_...)
}

func (c *ChartProject) AddTradingPairs(pairs ...sources.SourcePairId) {
	c.sourceManager.AddTradingPairs(pairs...)
}

func (c *ChartProject) SetTimeFrame(timeFrame string) {
	c.timeFrame = timeFrame
}

func (c *ChartProject) AddIndicator(indicator indicators.BaseIndicator) {
	c.indicators = append(c.indicators, indicator)
}

// TODO - implement rendering logic
// ! This will return an object with all the info needed to render the chart in the frontend
func (c *ChartProject) GetRender() RenderableChartProject {
	/*
		if len(c.data) == 0 {
			fmt.Println("No data fetched yet, fetching now...")
			c.FetchData()
		}

		fmt.Println("Rendering ChartProject with", len(c.indicators), "indicators and", len(c.data), "data points.")
		// Implementation to render the chart

		chartConfig := indicators.ChartConfig{
			TimeFrame: c.timeFrame,
		}

		plotData := c.CalculateIndicators(chartConfig)

		// TODO - this is just temporary to show the calculated plot data
		for name, points := range plotData {
			fmt.Println("Indicator:", name)
			for i, point := range points {
				fmt.Println(" Point", i, "Value:", point.Value, "Color:", point.Color)
			}
		}
	*/

	// Final rendering logic
	return RenderableChartProject{}
}

/*
// TODO - fix to support multiple pairs and sources
func (c *ChartProject) CalculateIndicators(chartConfig indicators.ChartConfig) map[string][]indicators.PlotPoint {
		plotData := make(map[string][]indicators.PlotPoint)

		for i := range c.sourceManager.fetchedData {
			multiSourceEntries := sources.NewMultiSourceEntries()
			multiSourceEntries.AddEntries(c.sourceManager.selectedPairs[0], c.sourceManager.fetchedData[:i+1])
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
*/
