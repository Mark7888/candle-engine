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

func (ti *TestIndicator1) HandleEntry(linePoints []indicators.PlotPoint, chartsData *sources.MultiSourceEntries, config indicators.ChartConfig) (indicators.PlotPoint, error) {
	lastPoint := indicators.PlotPoint{}
	if len(linePoints) > 0 {
		lastPoint = linePoints[len(linePoints)-1]
	}

	newPoint := indicators.PlotPoint{
		Value: lastPoint.Value + 1.0,
		Color: "green",
	}
	return newPoint, nil
}

func (ti *TestIndicator1) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (ti *TestIndicator1) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 2,
		},
	}
}

func CreateTestIndicator1() indicators.BaseIndicator {
	return &TestIndicator1{Name: "TestIndicator1"}
}
