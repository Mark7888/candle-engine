package chartproject

import (
	"candle-engine/pkg/charting/sources"
	"candle-engine/pkg/indicators"
	"fmt"
	"time"
)

type ChartProject struct {
	title         string
	colors        sources.CandleColorConfig
	sourceManager *ChartSourceManager
	indicators    []indicators.BaseIndicator

	timeRange sources.TimeRange
	timeFrame string // TODO - we should be storing time frame in seconds internally
}

func NewChartProject(title string) *ChartProject {
	return &ChartProject{
		title:         title,
		colors:        sources.CandleColorConfig{BullishColor: "#ff0000ff", BearishColor: "#00ff00ff"},
		sourceManager: NewChartSourceManager(),
		indicators:    []indicators.BaseIndicator{},
		timeRange: sources.TimeRange{
			Start: time.Now().AddDate(0, 0, -14),
			End:   time.Now(),
		},
		timeFrame: "1d",
	}
}

func (c *ChartProject) GetTitle() string {
	return c.title
}

func (c *ChartProject) SetTimeFrame(timeFrame string) {
	c.timeFrame = timeFrame
}

func (c *ChartProject) GetTimeFrame() string {
	return c.timeFrame
}

func (c *ChartProject) SetCandleColors(colors sources.CandleColorConfig) {
	c.colors = colors
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

func (c *ChartProject) GetLoadedPairs() []sources.SourcePairId {
	return c.sourceManager.GetLoadedPairs()
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

// GetRender returns an object with all the info needed to render the chart in
// the frontend.
//
// timeFrameOverride: if non-empty, it is used instead of the project's
// configured timeFrame.
// timeRangeOverride: if non-nil, it is used instead of the project's
// configured timeRange.
func (c *ChartProject) GetRender(activePair sources.SourcePairId, timeFrameOverride string, timeRangeOverride *sources.TimeRange) (RenderableChartProject, error) {
	timeFrame := c.timeFrame
	if timeFrameOverride != "" {
		timeFrame = timeFrameOverride
	}

	timeRange := c.timeRange
	if timeRangeOverride != nil {
		timeRange = *timeRangeOverride
	}

	entries, width, err := c.sourceManager.FetchSources(timeRange, c.sourceManager.selectedPairs, timeFrame)
	if err != nil {
		fmt.Println("Error fetching sources:", err)
		return RenderableChartProject{}, err
	}

	chartConfig := c.assembleChartConfigWith(activePair, timeFrame)

	var historyEntries map[sources.SourcePairId][]sources.ChartEntry = make(map[sources.SourcePairId][]sources.ChartEntry)
	var currentMultiSourceEntries sources.MultiSourceEntries
	var currentEntryTime time.Time

	var indicatorPlotData map[string][]indicators.PlotPointCollection = make(map[string][]indicators.PlotPointCollection)
	var indicatorLogs map[string][]indicators.IndicatorLog = make(map[string][]indicators.IndicatorLog)
	var indicatorErrors map[string][]error = make(map[string][]error)

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
	renderableChartProject.TimeRange = timeRange
	renderableChartProject.AvailableTimeFrames = c.sourceManager.GetCommonTimeFrames()
	renderableChartProject.ActiveTimeFrame = timeFrame
	renderableChartProject.LoadedPairs = c.sourceManager.GetLoadedPairs()
	renderableChartProject.ActivePair = activePair

	renderableChartProject.CandleData = entries
	renderableChartProject.CandleColors = c.colors
	renderableChartProject.IndicatorCharts = make(map[string][]RenderableIndicator)

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

func (c *ChartProject) assembleChartConfig(activePair sources.SourcePairId) indicators.ChartConfig {
	return c.assembleChartConfigWith(activePair, c.timeFrame)
}

func (c *ChartProject) assembleChartConfigWith(activePair sources.SourcePairId, timeFrame string) indicators.ChartConfig {
	return indicators.ChartConfig{
		TimeFrameSeconds: c.sourceManager.GetTimeFrameSeconds(timeFrame),
		ActivePair:       activePair,
		LoadedPairs:      c.sourceManager.GetLoadedPairs(),
	}
}

func (c *ChartProject) calculateIndicators(timestamp time.Time, previousPlotData *map[string][]indicators.PlotPointCollection, entries sources.MultiSourceEntries, chartConfig indicators.ChartConfig, previousLogs map[string][]indicators.IndicatorLog, previousErrors map[string][]error) (plotData map[string][]indicators.PlotPointCollection, indicatorLogs map[string][]indicators.IndicatorLog, indicatorErrors map[string][]error) {
	for _, indicator := range c.indicators {
		outputHandler := indicators.NewIndicatorOutputHandler(timestamp)
		outputHandler, err := indicator.HandleEntry(indicators.NewPlotPointCollectionHistory((*previousPlotData)[indicator.GetName()]), entries, chartConfig, outputHandler)
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
