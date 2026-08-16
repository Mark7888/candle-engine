package sample

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"fmt"
	"math"
)

func GetSampleKAMAIndicator() indicators.BaseIndicator {
	return &KAMAIndicator{
		Len:  90,
		Fast: 2,
		Slow: 30,
	}
}

type KAMAIndicator struct {
	Len         int
	Fast        int
	Slow        int
	CurrentKAMA float64
}

func (r *KAMAIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (r *KAMAIndicator) GetName() string {
	return "Kaufman's Adaptive Moving Average (KAMA)"
}

func (r *KAMAIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "", // empty means main chart
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 1,
			LineTitles:    []string{"KAMA"},
		},
	}
}

func (r *KAMAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	if chartsData.GetEntryCount(config.ActivePair) < r.Len+1 {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		// handler.LogDebug("Not enough data yet to calculate KAMA")
		return handler, nil // Not enough data yet
	}

	// Calculate change and volatility
	current := chartsData.GetEntry(config.ActivePair, 0)
	past := chartsData.GetEntry(config.ActivePair, r.Len)

	// Placeholder entries (pairs whose history starts later than the chart's
	// other pairs) carry no real price. Plotting a value derived from them would
	// stretch the main chart's y-axis and flatten every real candle.
	if current == nil || past == nil || !current.IsValid() || !past.IsValid() {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		return handler, nil
	}

	change := math.Abs(current.Close - past.Close)
	var volatility float64
	for i := 0; i < r.Len; i++ {
		entry1 := chartsData.GetEntry(config.ActivePair, i)
		entry2 := chartsData.GetEntry(config.ActivePair, i+1)
		volatility += math.Abs(entry1.Close - entry2.Close)
	}
	var er float64
	if volatility == 0 {
		er = 0
	} else {
		er = change / volatility
	}

	fastSC := 2.0 / float64(r.Fast+1)
	slowSC := 2.0 / float64(r.Slow+1)

	sc := math.Pow(er*(fastSC-slowSC)+slowSC, 2)

	// Calculate KAMA
	if points.Length() <= r.Len {
		r.CurrentKAMA = current.Close
	} else {
		prevKamaValue := r.CurrentKAMA
		handler.LogDebug(fmt.Sprintf("Previous KAMA point: %.4f", prevKamaValue))
		r.CurrentKAMA = prevKamaValue + sc*(current.Close-prevKamaValue)
	}

	msg := "KAMA calculated: " + fmt.Sprintf("%.4f", r.CurrentKAMA)
	handler.LogInfo(msg)

	longShrortPoint := indicators.PlotPoint{
		Value: r.CurrentKAMA,
		Color: "#6a59ff",
	}

	handler.SetSingleOutput(longShrortPoint)

	return handler, nil
}
