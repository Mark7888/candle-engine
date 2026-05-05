package indicators

import "candle-engine/pkg/sources"

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
	IndicatorTypePieChart
	IndicatorTypeCandleStick
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
	PieChartIndicator    PieChartIndicatorConfig
}

type LineIndicatorConfig struct {
	LineThickness int      `json:"lineThickness"`
	IsSmooth      bool     `json:"isSmooth"`
	LineTitles    []string `json:"lineTitles"`
	ShowSymbols   bool     `json:"showSymbols"` // Show circle markers at data points
	Opacity       float64  `json:"opacity"`     // Line opacity (0.0 to 1.0)
}

type StepAlignment int

const (
	StepAlignStart StepAlignment = iota
	StepAlignMidpoint
	StepAlignEnd
)

type StepLineIndicatorConfig struct {
	StepAlignment StepAlignment `json:"stepAlignment"`
	LineTitles    []string      `json:"lineTitles"`
	LineThickness int           `json:"lineThickness"`
	Opacity       float64       `json:"opacity"` // Line opacity (0.0 to 1.0)
}

type BarChartIndicatorConfig struct {
	BarTitles []string `json:"barTitles"`
	BarWidth  string   `json:"barWidth"` // Bar width (e.g., "60%", "20", "auto")
	Opacity   float64  `json:"opacity"`  // Bar opacity (0.0 to 1.0)
	Stacked   bool     `json:"stacked"`  // Whether bars should be stacked
}

type CandleStickIndicatorConfig struct {
	// Custom sub-chart candlestick (not the main price candles — a separate indicator sub-chart).
	// PlotPoints per timestamp: 4 points in order [Open, High, Low, Close].
	// Falls back to project CandleColors if these are empty strings.
	BullishColor string `json:"bullishColor"`
	BearishColor string `json:"bearishColor"`
}

type TableIndicatorConfig struct {
	DisplayName  string   `json:"displayName"`
	RowTitles    []string `json:"rowTitles"`    // This also defines the number of rows
	ColumnTitles []string `json:"columnTitles"` // This also defines the number of columns

	ValueDisplayFormat string `json:"valueDisplayFormat"` // e.g., "%.2f"
}

type AreaChartIndicatorConfig struct {
	LineTitles    []string `json:"lineTitles"`
	LineThickness int      `json:"lineThickness"`
	IsSmooth      bool     `json:"isSmooth"`
	ShowSymbols   bool     `json:"showSymbols"` // Show circle markers at data points
	Opacity       float64  `json:"opacity"`     // Area fill opacity (0.0 to 1.0)
}

type PieChartIndicatorConfig struct {
	DisplayName string   `json:"displayName"`
	SliceTitles []string `json:"sliceTitles"` // label for each slice; length defines number of slices
}
