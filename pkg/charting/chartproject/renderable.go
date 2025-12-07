package chartproject

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

// This will contain all the data needed to render the chart and indicators in the frontend
type RenderableChartProject struct {
	Title               string
	TimeRange           sources.TimeRange
	AvailableTimeFrames []string
	ActiveTimeFrame     string
	LoadedPairs         []sources.SourcePairId
	ActivePair          sources.SourcePairId

	CandleData map[sources.SourcePairId][]sources.ChartEntry

	Indicators []RenderableIndicator

	// Other metadata
}

func NewRenderableChartProject() RenderableChartProject {
	return RenderableChartProject{}
}

type RenderableIndicator struct {
	Name            string
	Type            indicators.IndicatorType
	IndicatorConfig indicators.IndicatorConfig
	PlotPoints      []indicators.PlotPoint

	RenderingErrors []error
}
