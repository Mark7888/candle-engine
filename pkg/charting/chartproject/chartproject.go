package chartproject

import (
	"fmt"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
	"time"
)

type ChartProject struct {
	title         string
	sourceManager *ChartSourceManager
	indicators    []indicators.BaseIndicator

	timeRange  sources.TimeRange
	timeFrame  string // TODO - we should be storing time frame in seconds internally
	activePair sources.SourcePairId
}

func NewChartProject(title string) *ChartProject {
	return &ChartProject{
		title:         title,
		sourceManager: NewChartSourceManager(),
		indicators:    []indicators.BaseIndicator{},
		timeRange: sources.TimeRange{
			Start: time.Now().AddDate(0, 0, -14),
			End:   time.Now(),
		},
		timeFrame: "1d",
	}
}

func (c *ChartProject) SetTimeFrame(timeFrame string) {
	c.timeFrame = timeFrame
}

func (c *ChartProject) GetTimeFrame() string {
	return c.timeFrame
}

func (c *ChartProject) GetAvailableTimeFrames() []string {
	return c.sourceManager.GetCommonTimeFrames()
}

// TODO - somehow handle available time frames from sources (only list the ones that are common across all sources)

func (c *ChartProject) SetTimeRange(timeRange sources.TimeRange) {
	c.timeRange = timeRange
}

func (c *ChartProject) GetTimeRange() sources.TimeRange {
	return c.timeRange
}

func (c *ChartProject) SetActivePair(pair sources.SourcePairId) {
	c.activePair = pair
}

func (c *ChartProject) GetActivePair() sources.SourcePairId {
	return c.activePair
}

func (c *ChartProject) AddSources(sources_ ...sources.Source) {
	c.sourceManager.AddSources(sources_...)
}

func (c *ChartProject) AddTradingPairs(pairs ...sources.SourcePairId) {
	c.sourceManager.AddTradingPairs(pairs...)
}

func (c *ChartProject) AddIndicators(indicator ...indicators.BaseIndicator) {
	c.indicators = append(c.indicators, indicator...)
}

// * This will return an object with all the info needed to render the chart in the frontend
func (c *ChartProject) GetRender() (RenderableChartProject, error) {
	entries, width, err := c.sourceManager.FetchSources(c.timeRange, c.sourceManager.selectedPairs, c.timeFrame)
	if err != nil {
		fmt.Println("Error fetching sources:", err)
		return RenderableChartProject{}, err
	}

	chartConfig := c.assembleChartConfig()

	var historyEntries map[sources.SourcePairId][]sources.ChartEntry = make(map[sources.SourcePairId][]sources.ChartEntry)
	var currentMultiSourceEntries *sources.MultiSourceEntries

	var indicatorPlotData map[string][]indicators.PlotPoint = make(map[string][]indicators.PlotPoint)
	var indicatorErrors map[string][]error

	for i := 0; i < width; i++ {
		for pairId, pairEntries := range entries {
			historyEntries[pairId] = append(historyEntries[pairId], pairEntries[i])
		}

		currentMultiSourceEntries = sources.NewMultiSourceEntries()
		currentMultiSourceEntries.ParseEntriesMap(historyEntries)

		indicatorPlotData, indicatorErrors = c.calculateIndicators(&indicatorPlotData, currentMultiSourceEntries, chartConfig, indicatorErrors)
	}

	renderableChartProject := NewRenderableChartProject()
	renderableChartProject.Title = c.title
	renderableChartProject.TimeRange = c.timeRange
	renderableChartProject.AvailableTimeFrames = c.sourceManager.GetCommonTimeFrames()
	renderableChartProject.ActiveTimeFrame = c.timeFrame
	renderableChartProject.LoadedPairs = c.sourceManager.GetLoadedPairs()
	renderableChartProject.ActivePair = c.activePair

	renderableChartProject.CandleData = entries

	for _, indicator := range c.indicators {
		renderableIndicator := RenderableIndicator{
			Name:            indicator.GetName(),
			Type:            indicator.GetType(),
			IndicatorConfig: indicator.GetConfig(),
			PlotPoints:      indicatorPlotData[indicator.GetName()],

			RenderingErrors: indicatorErrors[indicator.GetName()],
		}

		renderableChartProject.Indicators = append(renderableChartProject.Indicators, renderableIndicator)
	}

	return renderableChartProject, nil
}

func (c *ChartProject) assembleChartConfig() indicators.ChartConfig {
	return indicators.ChartConfig{
		TimeFrameSeconds: c.sourceManager.GetTimeFrameSeconds(c.timeFrame),
		ActivePair:       c.activePair,
		LoadedPairs:      c.sourceManager.GetLoadedPairs(),
	}
}

func (c *ChartProject) calculateIndicators(previousPlotData *map[string][]indicators.PlotPoint, entries *sources.MultiSourceEntries, chartConfig indicators.ChartConfig, previousErrors map[string][]error) (plotData map[string][]indicators.PlotPoint, indicatorErrors map[string][]error) {
	indicatorErrors = make(map[string][]error)

	// Loop through indicators and get their plot points
	for _, indicator := range c.indicators {
		point, err := indicator.HandleEntry((*previousPlotData)[indicator.GetName()], entries, chartConfig)
		if err != nil {
			point = indicators.GenerateInvalidPlotPoint()
			indicatorErrors[indicator.GetName()] = append(indicatorErrors[indicator.GetName()], err)
		}

		(*previousPlotData)[indicator.GetName()] = append((*previousPlotData)[indicator.GetName()], point)
	}

	return *previousPlotData, indicatorErrors
}

/*
func (c *ChartProject) Render() {
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
	}
*/

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
