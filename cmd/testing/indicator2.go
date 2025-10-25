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

func (ti *TestIndicator2) HandleEntry(lastPoint indicators.PlotPoint, chartData []sources.ChartEntry, config indicators.ChartConfig) (indicators.PlotPoint, error) {

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
