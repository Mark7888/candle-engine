package chartproject

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/registry"
	"candle-engine/pkg/sources"
	"fmt"
	"slices"
	"sync"
	"time"
)

// ChartProject is a serialisable config object describing a project's settings,
// selected pairs, and indicator names. It holds a non-serialised registry reference
// that must be injected (via SetRegistry) before calling GetRender.
type ChartProject struct {
	ID             string                    `json:"id"`
	Title          string                    `json:"title"`
	SelectedPairs  []sources.SourcePairId    `json:"selectedPairs"`
	IndicatorNames []string                  `json:"indicatorNames"`
	TimeFrame      string                    `json:"timeFrame"`
	TimeRange      sources.TimeRange         `json:"timeRange"`
	Colors         sources.CandleColorConfig `json:"colors"`

	// Non-serialised – injected at startup or after load.
	mu  sync.RWMutex   `json:"-"`
	reg *registry.Registry `json:"-"`
}

var defaultChartConfig = ChartProjectConfig{
	TimeFrame: "1d",
	CandleColors: sources.CandleColorConfig{
		BullishColor: "#ff0000ff",
		BearishColor: "#00ff00ff",
	},
}

func NewChartProject(id, title string, cfg ChartProjectConfig) *ChartProject {
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
		ID:             id,
		Title:          title,
		SelectedPairs:  []sources.SourcePairId{},
		IndicatorNames: []string{},
		TimeFrame:      cfg.TimeFrame,
		TimeRange:      cfg.TimeRange,
		Colors:         cfg.CandleColors,
	}
}

func (c *ChartProject) SetRegistry(r *registry.Registry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reg = r
}

func (c *ChartProject) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ID
}

func (c *ChartProject) GetTitle() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Title
}

func (c *ChartProject) GetTimeFrame() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TimeFrame
}

func (c *ChartProject) GetTimeRange() sources.TimeRange {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TimeRange
}

func (c *ChartProject) GetLoadedPairs() []sources.SourcePairId {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.SelectedPairs
}

func (c *ChartProject) GetColors() sources.CandleColorConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Colors
}

func (c *ChartProject) GetAvailableTimeFrames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.reg == nil {
		return []string{}
	}
	return c.reg.GetCommonTimeFrames(c.uniqueSourceNames())
}

func (c *ChartProject) uniqueSourceNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, p := range c.SelectedPairs {
		if !seen[p.SourceName] {
			seen[p.SourceName] = true
			names = append(names, p.SourceName)
		}
	}
	return names
}

func (c *ChartProject) AddPair(pair sources.SourcePairId) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !slices.Contains(c.SelectedPairs, pair) {
		c.SelectedPairs = append(c.SelectedPairs, pair)
	}
}

func (c *ChartProject) RemovePair(pair sources.SourcePairId) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SelectedPairs = slices.DeleteFunc(c.SelectedPairs, func(p sources.SourcePairId) bool {
		return p == pair
	})
}

func (c *ChartProject) AddIndicatorByName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !slices.Contains(c.IndicatorNames, name) {
		c.IndicatorNames = append(c.IndicatorNames, name)
	}
}

func (c *ChartProject) RemoveIndicatorByName(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IndicatorNames = slices.DeleteFunc(c.IndicatorNames, func(n string) bool {
		return n == name
	})
}

func (c *ChartProject) UpdateConfig(cfg ChartProjectConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if cfg.TimeFrame != "" {
		c.TimeFrame = cfg.TimeFrame
	}
	if !cfg.TimeRange.Start.IsZero() {
		c.TimeRange.Start = cfg.TimeRange.Start
	}
	if !cfg.TimeRange.End.IsZero() {
		c.TimeRange.End = cfg.TimeRange.End
	}
	if cfg.CandleColors.BullishColor != "" {
		c.Colors.BullishColor = cfg.CandleColors.BullishColor
	}
	if cfg.CandleColors.BearishColor != "" {
		c.Colors.BearishColor = cfg.CandleColors.BearishColor
	}
}

// GetRender fetches and assembles all data needed to render the chart for the
// specified source/pair. sourceName and pairId identify the active pair whose
// candle data and indicator output should be returned; if both are empty the
// first loaded pair is used as a fallback.
func (c *ChartProject) GetRender(sourceName, pairId string) (RenderableChartProject, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.reg == nil {
		partial := NewRenderableChartProject()
		partial.ProjectErrors = []error{fmt.Errorf("registry not set on project '%s'", c.Title)}
		return partial, nil
	}

	// Resolve the active pair: use the supplied source/pair, or fall back to the
	// first loaded pair when none is specified.
	activePair := sources.SourcePairId{SourceName: sourceName, PairId: pairId}
	if (activePair == sources.SourcePairId{}) && len(c.SelectedPairs) > 0 {
		activePair = c.SelectedPairs[0]
	}

	entries, width, err := c.reg.FetchSources(c.TimeRange, c.SelectedPairs, c.TimeFrame)
	if err != nil {
		partial := NewRenderableChartProject()
		partial.ProjectErrors = []error{err}
		return partial, nil
	}

	sourceNames := c.uniqueSourceNames()
	chartConfig := indicators.ChartConfig{
		TimeFrameSeconds: c.reg.GetTimeFrameSeconds(sourceNames, c.TimeFrame),
		ActivePair:       activePair,
		LoadedPairs:      c.SelectedPairs,
	}

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
	renderableChartProject.Title = c.Title
	renderableChartProject.TimeRange = c.TimeRange
	renderableChartProject.AvailableTimeFrames = c.reg.GetCommonTimeFrames(sourceNames)
	renderableChartProject.ActiveTimeFrame = c.TimeFrame
	renderableChartProject.LoadedPairs = c.SelectedPairs
	renderableChartProject.ActivePair = activePair
	renderableChartProject.CandleData = entries
	renderableChartProject.CandleColors = c.Colors
	renderableChartProject.IndicatorCharts = make(map[string][]RenderableIndicator)

	for _, name := range c.IndicatorNames {
		indicator, ok := c.reg.GetIndicator(name)
		if !ok {
			renderableChartProject.ProjectErrors = append(renderableChartProject.ProjectErrors, fmt.Errorf("indicator '%s' not found in registry", name))
			continue
		}

		isDataInconsistent := c.checkIndicatorDataConsistency(indicatorPlotData[name])
		renderableIndicator := RenderableIndicator{
			Name:            indicator.GetName(),
			Type:            indicator.GetType(),
			IndicatorConfig: indicator.GetConfig(),
			PlotPoints:      indicatorPlotData[name],
			DataError:       isDataInconsistent,
			RenderingErrors: indicatorErrors[name],
			Logs:            indicatorLogs[name],
		}

		renderableChartProject.IndicatorCharts[indicator.GetConfig().TargetChart] = append(
			renderableChartProject.IndicatorCharts[indicator.GetConfig().TargetChart],
			renderableIndicator,
		)
	}

	return renderableChartProject, nil
}

func (c *ChartProject) calculateIndicators(timestamp time.Time, previousPlotData *map[string][]indicators.PlotPointCollection, entries sources.MultiSourceEntries, chartConfig indicators.ChartConfig, previousLogs map[string][]indicators.IndicatorLog, previousErrors map[string][]error) (plotData map[string][]indicators.PlotPointCollection, indicatorLogs map[string][]indicators.IndicatorLog, indicatorErrors map[string][]error) {
	for _, name := range c.IndicatorNames {
		indicator, ok := c.reg.GetIndicator(name)
		if !ok {
			continue
		}

		outputHandler := indicators.NewIndicatorOutputHandler(timestamp)
		outputHandler, err := indicator.HandleEntry(indicators.NewPlotPointCollectionHistory((*previousPlotData)[name]), entries, chartConfig, outputHandler)
		if err != nil {
			outputHandler.SetOutput(indicators.NewInvalidPlotPointCollection())
			previousErrors[name] = append(previousErrors[name], err)
		}

		(*previousPlotData)[name] = append((*previousPlotData)[name], outputHandler.GetOutput())
		previousLogs[name] = append(previousLogs[name], outputHandler.GetLogs()...)
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
