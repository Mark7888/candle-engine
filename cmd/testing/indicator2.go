package main

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

type TestIndicator2 struct {
	Name string
}

func (ti *TestIndicator2) GetName() string {
	return ti.Name
}

func (ti *TestIndicator2) HandleEntry(plotPoints []indicators.PlotPoint, chartData []sources.ChartEntry, config indicators.ChartConfig) (indicators.PlotPoint, error) {
	lastPoint := indicators.PlotPoint{}
	if len(plotPoints) > 0 {
		lastPoint = plotPoints[len(plotPoints)-1]
	}

	newPoint := indicators.PlotPoint{
		Value: 0.00,
		Color: "green",
	}

	if lastPoint.Value == 0 {
		newPoint.Value = 10.00
	}

	return newPoint, nil
}

func CreateTestIndicator2() indicators.PlotIndicator {
	return &TestIndicator2{Name: "TestIndicator2"}
}
