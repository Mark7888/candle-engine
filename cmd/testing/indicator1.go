package main

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

type TestIndicator1 struct {
	Name string
}

func (ti *TestIndicator1) GetName() string {
	return ti.Name
}

func (ti *TestIndicator1) HandleEntry(plotPoints []indicators.PlotPoint, chartsData *sources.MultiSourceEntries, config indicators.ChartConfig) (indicators.PlotPoint, error) {
	lastPoint := indicators.PlotPoint{}
	if len(plotPoints) > 0 {
		lastPoint = plotPoints[len(plotPoints)-1]
	}

	newPoint := indicators.PlotPoint{
		Value: lastPoint.Value + 1.0,
		Color: "green",
	}
	return newPoint, nil
}

func CreateTestIndicator1() indicators.PlotIndicator {
	return &TestIndicator1{Name: "TestIndicator1"}
}
