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
	IndicatorTypeAreaChart
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
	AreaChartIndicator   AreaChartIndicatorConfig
}

type LineIndicatorConfig struct {
	LineThickness int
	IsSmooth      bool
	LineTitles    []string
	ShowSymbols   bool    // Show circle markers at data points
	Opacity       float64 // Line opacity (0.0 to 1.0)
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
	LineThickness int
	Opacity       float64 // Line opacity (0.0 to 1.0)
}

type BarChartIndicatorConfig struct {
	BarTitles []string
	BarWidth  string  // Bar width (e.g., "60%", "20", "auto")
	Opacity   float64 // Bar opacity (0.0 to 1.0)
	Stacked   bool    // Whether bars should be stacked
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

type AreaChartIndicatorConfig struct {
	LineTitles    []string
	LineThickness int
	IsSmooth      bool
	ShowSymbols   bool    // Show circle markers at data points
	Opacity       float64 // Area fill opacity (0.0 to 1.0)
}

// TODO - add indicator configs for type: PieChart, etc.
