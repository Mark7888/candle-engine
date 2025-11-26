package indicators

// ChartConfig holds general chart configuration applicable to all indicators
type ChartConfig struct {
	IntervalSeconds int // in seconds
}

// IndicatorConfig holds type-specific configuration for different indicator types
type IndicatorConfig struct {
	LineIndicator     LineIndicatorConfig
	StepLineIndicator StepLineIndicatorConfig
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
