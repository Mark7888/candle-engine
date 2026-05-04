package sample

import (
	"candle-engine/pkg/charting/sources"
	"candle-engine/pkg/indicatorcolor"
	"candle-engine/pkg/indicators"
	"fmt"
	"math"
)

func GetSampleColoredKAMAIndicator() indicators.BaseIndicator {
	kama := &ColoredKAMAIndicator{
		Len:  90,
		Fast: 2,
		Slow: 30,
	}

	return indicatorcolor.GetNewColoredIndicator(kama, 3)
}

type ColoredKAMAIndicator struct {
	Len         int
	Fast        int
	Slow        int
	CurrentKAMA float64
}

func (r *ColoredKAMAIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (r *ColoredKAMAIndicator) GetName() string {
	return "Kaufman's Adaptive Moving Average (KAMA) - Colored"
}

func (r *ColoredKAMAIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "", // empty means main chart
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 1,
			LineTitles:    []string{"KAMA"},
		},
	}
}

func (r *ColoredKAMAIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	if chartsData.GetEntryCount(config.ActivePair) < r.Len+1 {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		// handler.LogDebug("Not enough data yet to calculate KAMA")
		return handler, nil // Not enough data yet
	}

	// Calculate change and volatility
	current := chartsData.GetEntry(config.ActivePair, 0)
	past := chartsData.GetEntry(config.ActivePair, r.Len)

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
		r.CurrentKAMA = prevKamaValue + sc*(current.Close-prevKamaValue)
	}

	msg := "KAMA calculated: " + fmt.Sprintf("%.4f", r.CurrentKAMA)
	handler.LogInfo(msg)

	color := "#6a59ff" // default color
	if r.CurrentKAMA > current.Close {
		color = "#ff5959" // red for bearish
	} else if r.CurrentKAMA < current.Close {
		color = "#59ff59" // green for bullish
	}

	longShrortPoint := indicators.PlotPoint{
		Value: r.CurrentKAMA,
		Color: color,
	}

	handler.SetSingleOutput(longShrortPoint)

	return handler, nil
}
