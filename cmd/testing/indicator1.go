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

func (ti *TestIndicator1) HandleEntry(lastPoint indicators.PlotPoint, chartData []sources.ChartEntry, config indicators.ChartConfig) (indicators.PlotPoint, error) {
	newPoint := indicators.PlotPoint{
		Value: lastPoint.Value + 1.0,
		Color: "green",
	}
	return newPoint, nil
}

func CreateTestIndicator1() indicators.PlotIndicator {
	return &TestIndicator1{Name: "TestIndicator1"}
}
