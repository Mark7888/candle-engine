package chartproject

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"time"
)

// TODO - move sourceManager and indicators into registry
// * Registry will handle all the sources and indicators available in the application, and provide them to the ChartProject when needed
// * We have to create, store, load, update or delete these ChartProject instances and their configurations, like added sources and indicators, default settings, etc.
// * ChartProject will only be responsible for assembling the data for rendering, and calculating indicators based on the data and the chart configuration.
// * It will not be responsible for storing the available sources and indicators in the application, or managing them in any way.
// * It will simply call the registry when needs anything.
// ! Have to rework how we handle project configurations and settings. Do we store everything on the backend?
// ? What do we store on the frontend?
// - Render settings (time frame, active pair, time range)
// ? What to store on backend in config?
// - Projects and their settings (title, added sources and indicators, default render settings like time frame, time range, candle colors, etc.)
// - Whole-app settings

type ChartProject struct {
	title         string
	sourceManager *ChartSourceManager
	indicators    []indicators.BaseIndicator

	colors    sources.CandleColorConfig
	timeRange sources.TimeRange
	timeFrame string // TODO - we should be storing time frame in seconds internally
}

var defaultChartConfig = ChartProjectConfig{
	TimeFrame: "1d",
	CandleColors: sources.CandleColorConfig{
		BullishColor: "#ff0000ff",
		BearishColor: "#00ff00ff",
	},
}

// NewChartProject creates a ChartProject with the given title and configuration.
// Any zero-value field in cfg is replaced by a sensible built-in default:
func NewChartProject(title string, cfg ChartProjectConfig) *ChartProject {
	if cfg.TimeFrame == "" {
		cfg.TimeFrame = defaultChartConfig.TimeFrame
	}

	if cfg.TimeRange.End.IsZero() {
		cfg.TimeRange.End = time.Now()
	}
	if cfg.TimeRange.Start.IsZero() {
		cfg.TimeRange.Start = cfg.TimeRange.End.AddDate(0, 0, -14)
	}

	if cfg.CandleColors.BullishColor == "" {
		cfg.CandleColors.BullishColor = defaultChartConfig.CandleColors.BullishColor
	}
	if cfg.CandleColors.BearishColor == "" {
		cfg.CandleColors.BearishColor = defaultChartConfig.CandleColors.BearishColor
	}

	return &ChartProject{
		title:         title,
		colors:        cfg.CandleColors,
		sourceManager: NewChartSourceManager(),
		indicators:    []indicators.BaseIndicator{},
		timeRange:     cfg.TimeRange,
		timeFrame:     cfg.TimeFrame,
	}
}

func (c *ChartProject) GetTitle() string {
	return c.title
}

func (c *ChartProject) GetTimeFrame() string {
	return c.timeFrame
}

func (c *ChartProject) GetAvailableTimeFrames() []string {
	return c.sourceManager.GetCommonTimeFrames()
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
// Fields in cfg that are at their zero value fall back to the project's
// configured defaults (TimeFrame, TimeRange, CandleColors).
func (c *ChartProject) GetRender(cfg RenderConfig) (RenderableChartProject, error) {
	timeFrame := c.timeFrame
	if cfg.TimeFrame != "" {
		timeFrame = cfg.TimeFrame
	}

	timeRange := c.timeRange
	if !cfg.RangeFrom.IsZero() {
		timeRange.Start = cfg.RangeFrom
	}
	if !cfg.RangeTo.IsZero() {
		timeRange.End = cfg.RangeTo
	}

	colors := c.colors
	if cfg.CandleColors != nil {
		colors = *cfg.CandleColors
	}

	entries, width, err := c.sourceManager.FetchSources(timeRange, c.sourceManager.selectedPairs, timeFrame)
	if err != nil {
		partial := NewRenderableChartProject()
		partial.ProjectErrors = []error{err}
		return partial, nil
	}

	chartConfig := c.assembleChartConfigWith(cfg.ActivePair, timeFrame)

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
	renderableChartProject.ActivePair = cfg.ActivePair

	renderableChartProject.CandleData = entries
	renderableChartProject.CandleColors = colors
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
