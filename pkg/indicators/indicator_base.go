package indicators

import "testcharting/pkg/charting/sources"

type IndicatorType int

const (
	IndicatorTypeLine IndicatorType = iota
	IndicatorTypeStepLine
)

// TODO - change PlotPoint to support multiple values (e.g., for pie charts, histograms, etc.)
type PlotPoint struct {
	Value float64
	Color string
}

func GenerateInvalidPlotPoint() PlotPoint {
	return PlotPoint{
		Value: -999,
		Color: "red",
	}
}

type BaseIndicator interface {
	GetName() string
	GetType() IndicatorType
	GetConfig() IndicatorConfig
	HandleEntry(points []PlotPoint, chartsData *sources.MultiSourceEntries, config ChartConfig) (PlotPoint, error)
}
