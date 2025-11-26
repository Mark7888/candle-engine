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

func (ti *TestIndicator2) HandleEntry(linePoints []indicators.PlotPoint, chartsData *sources.MultiSourceEntries, config indicators.ChartConfig) (indicators.PlotPoint, error) {
	lastPoint := indicators.PlotPoint{}
	if len(linePoints) > 0 {
		lastPoint = linePoints[len(linePoints)-1]
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

func (ti *TestIndicator2) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeStepLine
}

func (ti *TestIndicator2) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		StepLineIndicator: indicators.StepLineIndicatorConfig{
			StepAlignment: 1,
		},
	}
}

func CreateTestIndicator2() indicators.BaseIndicator {
	return &TestIndicator2{Name: "TestIndicator2"}
}
