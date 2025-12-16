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
	var currentMultiSourceEntries sources.MultiSourceEntries
	var currentEntryTime time.Time

	var indicatorPlotData map[string][]indicators.PlotPointCollection = make(map[string][]indicators.PlotPointCollection)
	var indicatorLogs map[string][]indicators.IndicatorLog
	var indicatorErrors map[string][]error

	for i := 0; i < width; i++ {
		for pairId, pairEntries := range entries {
			historyEntries[pairId] = append(historyEntries[pairId], pairEntries[i])
			currentEntryTime = pairEntries[i].Timestamp
		}

		currentMultiSourceEntries = sources.NewMultiSourceEntries()
		currentMultiSourceEntries.ParseEntriesMap(historyEntries)

		indicatorPlotData, indicatorLogs, indicatorErrors = c.calculateIndicators(currentEntryTime, &indicatorPlotData, currentMultiSourceEntries, chartConfig, indicatorLogs, indicatorErrors)
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
		isDataInconsistent := c.checkIndicatorDataConsistency(indicatorPlotData[indicator.GetName()])
		// TODO - that would be also helpful to check if the number of points and the number of titles in the config match

		renderableIndicator := RenderableIndicator{
			Name:            indicator.GetName(),
			Type:            indicator.GetType(),
			IndicatorConfig: indicator.GetConfig(),
			PlotPoints:      indicatorPlotData[indicator.GetName()],

			DataError:       isDataInconsistent,
			RenderingErrors: indicatorErrors[indicator.GetName()],
			Logs:            indicatorLogs[indicator.GetName()],
		}

		renderableChartProject.IndicatorCharts[indicator.GetConfig().TargetChart] = append(renderableChartProject.IndicatorCharts[indicator.GetConfig().TargetChart], renderableIndicator)
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

func (c *ChartProject) calculateIndicators(timestamp time.Time, previousPlotData *map[string][]indicators.PlotPointCollection, entries sources.MultiSourceEntries, chartConfig indicators.ChartConfig, previousLogs map[string][]indicators.IndicatorLog, previousErrors map[string][]error) (plotData map[string][]indicators.PlotPointCollection, indicatorLogs map[string][]indicators.IndicatorLog, indicatorErrors map[string][]error) {
	for _, indicator := range c.indicators {
		outputHandler := indicators.NewIndicatorOutputHandler(timestamp)
		outputHandler, err := indicator.HandleEntry((*previousPlotData)[indicator.GetName()], entries, chartConfig, outputHandler)
		if err != nil {
			outputHandler.SetOutput(indicators.NewInvalidPlotPointCollection())
			previousErrors[indicator.GetName()] = append(previousErrors[indicator.GetName()], err)
		}

		(*previousPlotData)[indicator.GetName()] = append((*previousPlotData)[indicator.GetName()], outputHandler.GetOutput())
		previousLogs[indicator.GetName()] = append(previousLogs[indicator.GetName()], outputHandler.GetLogs()...)
	}

	return *previousPlotData, previousLogs, previousErrors
}

func (c *ChartProject) checkIndicatorDataConsistency(plotPoints []indicators.PlotPointCollection) bool {
	pointLength := -1

	for _, p := range plotPoints {
		if pointLength == -1 {
			pointLength = p.Length()
		}

		if p.Length() != pointLength {
			return true
		}
	}

	return false
}
