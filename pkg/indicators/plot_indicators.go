package indicators

import "testcharting/pkg/charting/sources"

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

type PlotIndicator interface {
	GetName() string
	HandleEntry(points []PlotPoint, chartData []sources.ChartEntry, config ChartConfig) (PlotPoint, error)
}
