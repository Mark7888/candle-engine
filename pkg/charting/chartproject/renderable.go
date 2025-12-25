package chartproject

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

type RenderableIndicator struct {
	Name            string
	Type            indicators.IndicatorType
	IndicatorConfig indicators.IndicatorConfig
	PlotPoints      []indicators.PlotPointCollection

	DataError       bool // if true, indicates that there was an error in data rendering (e.g., inconsistent data lengths)
	RenderingErrors []error
	Logs            []indicators.IndicatorLog
}

// TODO - add methods to RenderableIndicator struct

// This will contain all the data needed to render the chart and indicators in the frontend
type RenderableChartProject struct {
	Title               string
	TimeRange           sources.TimeRange
	AvailableTimeFrames []string
	ActiveTimeFrame     string
	LoadedPairs         []sources.SourcePairId
	ActivePair          sources.SourcePairId

	CandleData map[sources.SourcePairId][]sources.ChartEntry

	IndicatorCharts map[string][]RenderableIndicator

	// Other metadata
}

func NewRenderableChartProject() RenderableChartProject {
	return RenderableChartProject{}
}

// TODO - add methods to RenderableChartProject struct
