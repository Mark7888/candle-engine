package sample

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
)

func GetSampleMAIndicator() indicators.BaseIndicator {
	return &MAIndicator{}
}

type MAIndicator struct {
}

func (r *MAIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (r *MAIndicator) GetName() string {
	return "Moving Average (MA)"
}

func (r *MAIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "", // empty means main chart
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 1,
			LineTitles:    []string{"MA5", "MA10", "MA20"},
		},
	}
}

func (r *MAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	availableData := chartsData.GetEntryCount(config.ActivePair)

	// Calculate MA5
	var pointMA5 indicators.PlotPoint
	if availableData >= 5 {
		var sum5 float64
		valid := true
		for i := 0; i < 5; i++ {
			entry := chartsData.GetEntry(config.ActivePair, i)
			if entry == nil || !entry.IsValid() {
				valid = false
				break
			}
			sum5 += entry.Close
		}
		if valid {
			pointMA5 = indicators.PlotPoint{
				Value: sum5 / 5.0,
				Color: "#00ff00", // Green for fastest MA
			}
		} else {
			pointMA5 = indicators.NewInvalidPlotPoint()
		}
	} else {
		pointMA5 = indicators.NewInvalidPlotPoint()
	}

	// Calculate MA10
	var pointMA10 indicators.PlotPoint
	if availableData >= 10 {
		var sum10 float64
		valid := true
		for i := 0; i < 10; i++ {
			entry := chartsData.GetEntry(config.ActivePair, i)
			if entry == nil || !entry.IsValid() {
				valid = false
				break
			}
			sum10 += entry.Close
		}
		if valid {
			pointMA10 = indicators.PlotPoint{
				Value: sum10 / 10.0,
				Color: "#ffaa00", // Orange for medium MA
			}
		} else {
			pointMA10 = indicators.NewInvalidPlotPoint()
		}
	} else {
		pointMA10 = indicators.NewInvalidPlotPoint()
	}

	// Calculate MA20
	var pointMA20 indicators.PlotPoint
	if availableData >= 20 {
		var sum20 float64
		valid := true
		for i := 0; i < 20; i++ {
			entry := chartsData.GetEntry(config.ActivePair, i)
			if entry == nil || !entry.IsValid() {
				valid = false
				break
			}
			sum20 += entry.Close
		}
		if valid {
			pointMA20 = indicators.PlotPoint{
				Value: sum20 / 20.0,
				Color: "#ff0000", // Red for slowest MA
			}
		} else {
			pointMA20 = indicators.NewInvalidPlotPoint()
		}
	} else {
		pointMA20 = indicators.NewInvalidPlotPoint()
	}

	pointCollection := indicators.NewPlotPointCollection(pointMA5, pointMA10, pointMA20)

	handler.SetOutput(pointCollection)

	return handler, nil
}
