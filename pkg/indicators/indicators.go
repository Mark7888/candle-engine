package indicators

import "testcharting/pkg/charting/sources"

type ChartConfig struct {
	Interval string
	// Other configuration fields can be added here
}

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

type BasicIndicator interface {
	GetName() string
	HandleEntry(lastPoint PlotPoint, chartData []sources.ChartEntry, config ChartConfig) (PlotPoint, error)
}
