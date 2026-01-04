package indicators

import "testcharting/pkg/charting/sources"

// ChartConfig holds general chart configuration applicable to all indicators
type ChartConfig struct {
	TimeFrameSeconds int64 // in seconds
	ActivePair       sources.SourcePairId
	LoadedPairs      []sources.SourcePairId
}

// IndicatorType represents the type of indicator
type IndicatorType int

const (
	IndicatorTypeLine IndicatorType = iota
	IndicatorTypeStepLine
	IndicatorTypeBarChart
	IndicatorTypeTable
	IndicatorTypeHistogram
)

// IndicatorConfig holds type-specific configuration for different indicator types
type IndicatorConfig struct {
	// Empty string for main chart, any other string for custom sub-charts
	TargetChart string

	LineIndicator        LineIndicatorConfig
	StepLineIndicator    StepLineIndicatorConfig
	BarChartIndicator    BarChartIndicatorConfig
	CandleStickIndicator CandleStickIndicatorConfig
	TableIndicator       TableIndicatorConfig
	HistogramIndicator   HistogramIndicatorConfig
}

type LineIndicatorConfig struct {
	LineThickness int
	IsSmooth      bool
	LineTitles    []string
}

type StepAlignment int

const (
	StepAlignStart StepAlignment = iota
	StepAlignMidpoint
	StepAlignEnd
)

type StepLineIndicatorConfig struct {
	StepAlignment StepAlignment
	LineTitles    []string
}

type BarChartIndicatorConfig struct {
	BarTitles []string
	// TODO - add more bar chart specific configurations
}

type CandleStickIndicatorConfig struct {
	// TODO - add candlestick specific configurations
}

type TableIndicatorConfig struct {
	DisplayName  string
	RowTitles    []string // This also defines the number of rows
	ColumnTitles []string // This also defines the number of columns

	ValueDisplayFormat string // e.g., "%.2f" // TODO - implement proper formatting
}

type HistogramIndicatorConfig struct{}

// TODO - add indicator configs for type: BarChart, AreaChart, PieChart, etc.
