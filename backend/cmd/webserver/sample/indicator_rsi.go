package sample

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
)

func GetSampleRSIIndicator() indicators.BaseIndicator {
	return &RSIIndicator{}
}

type RSIIndicator struct {
}

func (r *RSIIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (r *RSIIndicator) GetName() string {
	return "Relative Strength Index (RSI)"
}

func (r *RSIIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "rsi", // empty means main chart
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 1,
			LineTitles:    []string{"RSI"},
		},
	}
}

func (r *RSIIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	// RSI calculation requires 14 periods of data
	period := 14
	if chartsData.GetEntryCount(config.ActivePair) < period+1 {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		handler.LogDebug("Not enough data yet to calculate RSI")
		return handler, nil // Not enough data yet
	}

	// Calculate gains and losses over the period
	var gains, losses float64
	for i := 0; i < period; i++ {
		current := chartsData.GetEntry(config.ActivePair, i)
		previous := chartsData.GetEntry(config.ActivePair, i+1)

		change := current.Close - previous.Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	// Calculate RSI
	var rsi float64
	if avgLoss == 0 {
		rsi = 100
	} else {
		rs := avgGain / avgLoss
		rsi = 100 - (100 / (1 + rs))
	}

	handler.SetSingleOutput(indicators.PlotPoint{
		Value: rsi,
		Color: "#00ffff",
	})

	return handler, nil
}
