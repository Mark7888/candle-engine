package indicators

// ChartConfig holds general chart configuration applicable to all indicators
type ChartConfig struct {
	TimeFrameSeconds int64 // in seconds
	ActivePair       string
	LoadedPairs      []string
}

// IndicatorConfig holds type-specific configuration for different indicator types
type IndicatorConfig struct {
	LineIndicator        LineIndicatorConfig
	StepLineIndicator    StepLineIndicatorConfig
	SingleValueIndicator SingleValueIndicatorConfig
}

type LineIndicatorConfig struct {
	LineThickness int
}

type StepAlignment int

const (
	StepAlignStart StepAlignment = iota
	StepAlignMidpoint
	StepAlignEnd
)

type StepLineIndicatorConfig struct {
	StepAlignment StepAlignment
}

type SingleValueIndicatorConfig struct {
	DisplayName   string
	DisplayFormat string // e.g., "%.2f" // TODO - implement proper formatting
}
