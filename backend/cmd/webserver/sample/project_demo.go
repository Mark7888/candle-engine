package sample

import (
	"candle-engine/pkg/chartproject"
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"fmt"
	"math"
	"time"
)

func GetDemoChartProject() *chartproject.ChartProject {
	allPairs := []sources.SourcePairId{
		{SourceName: "Binance", PairId: "BTCUSDC"},
		{SourceName: "Binance", PairId: "ETHUSDC"},
	}

	startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")

	project := chartproject.NewChartProject("demo-project-01", "Demo Project", chartproject.ChartProjectConfig{
		TimeFrame:    "1d",
		TimeRange:    sources.TimeRange{Start: startTime, End: endTime},
		CandleColors: sources.CandleColorConfig{BullishColor: "#26a69a", BearishColor: "#ef5350"},
	})
	for _, pair := range allPairs {
		project.AddPair(pair)
	}
	project.AddIndicatorByName(GetDemoLineIndicator().GetName())
	project.AddIndicatorByName(GetDemoStepLineIndicator().GetName())
	project.AddIndicatorByName(GetDemoBarChartIndicator().GetName())
	project.AddIndicatorByName(GetDemoAreaChartIndicator().GetName())
	project.AddIndicatorByName(GetDemoCandleStickIndicator().GetName())
	project.AddIndicatorByName(GetDemoTableIndicator().GetName())
	project.AddIndicatorByName(GetDemoPieChartIndicator().GetName())
	project.AddIndicatorByName(GetDemoPieChart2Indicator().GetName())
	return project
}

// ---------------------------------------------------------------------------
// Line – SMA-20
// ---------------------------------------------------------------------------

func GetDemoLineIndicator() indicators.BaseIndicator {
	return &DemoLineIndicator{}
}

type DemoLineIndicator struct{}

func (d *DemoLineIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeLine
}

func (d *DemoLineIndicator) GetName() string {
	return "SMA-20 (Demo)"
}

func (d *DemoLineIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "",
		LineIndicator: indicators.LineIndicatorConfig{
			LineThickness: 2,
			LineTitles:    []string{"SMA-20"},
			IsSmooth:      true,
		},
	}
}

func (d *DemoLineIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	period := 20
	if chartsData.GetEntryCount(config.ActivePair) < period {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		return handler, nil
	}

	var sum float64
	for i := 0; i < period; i++ {
		sum += chartsData.GetEntry(config.ActivePair, i).Close
	}

	handler.SetSingleOutput(indicators.PlotPoint{
		Value: sum / float64(period),
		Color: "#f6c90e",
	})
	return handler, nil
}

// ---------------------------------------------------------------------------
// StepLine – Price Step
// ---------------------------------------------------------------------------

func GetDemoStepLineIndicator() indicators.BaseIndicator {
	return &DemoStepLineIndicator{}
}

type DemoStepLineIndicator struct{}

func (d *DemoStepLineIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeStepLine
}

func (d *DemoStepLineIndicator) GetName() string {
	return "Price Step (Demo)"
}

func (d *DemoStepLineIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "demo-step",
		StepLineIndicator: indicators.StepLineIndicatorConfig{
			StepAlignment: indicators.StepAlignMidpoint,
			LineTitles:    []string{"Close"},
			LineThickness: 1,
		},
	}
}

func (d *DemoStepLineIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	handler.SetSingleOutput(indicators.PlotPoint{
		Value: chartsData.GetEntry(config.ActivePair, 0).Close,
		Color: "#00e5ff",
	})
	return handler, nil
}

// ---------------------------------------------------------------------------
// BarChart – Volume (logs when volume >= 50k)
// ---------------------------------------------------------------------------

func GetDemoBarChartIndicator() indicators.BaseIndicator {
	return &DemoBarChartIndicator{}
}

type DemoBarChartIndicator struct{}

func (d *DemoBarChartIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeBarChart
}

func (d *DemoBarChartIndicator) GetName() string {
	return "Volume (Demo)"
}

func (d *DemoBarChartIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "demo-volume",
		BarChartIndicator: indicators.BarChartIndicatorConfig{
			BarTitles: []string{"Volume"},
			BarWidth:  "60%",
			Opacity:   0.8,
		},
	}
}

func (d *DemoBarChartIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	vol := chartsData.GetEntry(config.ActivePair, 0).Volume
	if vol >= 50000 {
		handler.LogInfo(fmt.Sprintf("High volume: %.2f", vol))
	}
	handler.SetSingleOutput(indicators.PlotPoint{
		Value: vol,
		Color: "#ab47bc",
	})
	return handler, nil
}

// ---------------------------------------------------------------------------
// AreaChart – RSI-14
// ---------------------------------------------------------------------------

func GetDemoAreaChartIndicator() indicators.BaseIndicator {
	return &DemoAreaChartIndicator{}
}

type DemoAreaChartIndicator struct{}

func (d *DemoAreaChartIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeAreaChart
}

func (d *DemoAreaChartIndicator) GetName() string {
	return "Area RSI (Demo)"
}

func (d *DemoAreaChartIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "demo-area",
		AreaChartIndicator: indicators.AreaChartIndicatorConfig{
			LineTitles:    []string{"RSI"},
			LineThickness: 1,
			IsSmooth:      true,
			Opacity:       0.4,
		},
	}
}

func (d *DemoAreaChartIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	period := 14
	if chartsData.GetEntryCount(config.ActivePair) < period+1 {
		handler.SetSingleOutput(indicators.NewInvalidPlotPoint())
		return handler, nil
	}

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

	var rsi float64
	if avgLoss == 0 {
		rsi = 100
	} else {
		rs := avgGain / avgLoss
		rsi = 100 - (100 / (1 + rs))
	}

	color := "#29b6f6"
	if rsi > 70 {
		color = "#ff7043"
	} else if rsi < 30 {
		color = "#66bb6a"
	}

	handler.SetSingleOutput(indicators.PlotPoint{
		Value: rsi,
		Color: color,
	})
	return handler, nil
}

// ---------------------------------------------------------------------------
// CandleStick – Custom OHLC sub-chart
// ---------------------------------------------------------------------------

func GetDemoCandleStickIndicator() indicators.BaseIndicator {
	return &DemoCandleStickIndicator{}
}

type DemoCandleStickIndicator struct{}

func (d *DemoCandleStickIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeCandleStick
}

func (d *DemoCandleStickIndicator) GetName() string {
	return "Custom Candles (Demo)"
}

func (d *DemoCandleStickIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "demo-candles",
		CandleStickIndicator: indicators.CandleStickIndicatorConfig{
			BullishColor: "#26a69a",
			BearishColor: "#ef5350",
		},
	}
}

func (d *DemoCandleStickIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	entry := chartsData.GetEntry(config.ActivePair, 0)
	if entry == nil || !entry.IsValid() {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	handler.SetOutput(indicators.NewPlotPointCollection(
		indicators.PlotPoint{Value: entry.Open, Color: ""},
		indicators.PlotPoint{Value: entry.High, Color: ""},
		indicators.PlotPoint{Value: entry.Low, Color: ""},
		indicators.PlotPoint{Value: entry.Close, Color: ""},
	))
	return handler, nil
}

// ---------------------------------------------------------------------------
// Table – OHLC Snapshot (2 rows x 2 columns)
// ---------------------------------------------------------------------------

func GetDemoTableIndicator() indicators.BaseIndicator {
	return &DemoTableIndicator{}
}

type DemoTableIndicator struct{}

func (d *DemoTableIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeTable
}

func (d *DemoTableIndicator) GetName() string {
	return "OHLC Snapshot (Demo)"
}

func (d *DemoTableIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "",
		TableIndicator: indicators.TableIndicatorConfig{
			DisplayName:        "OHLC Snapshot",
			RowTitles:          []string{"Current", "Previous"},
			ColumnTitles:       []string{"Open", "Close"},
			ValueDisplayFormat: "%.2f",
		},
	}
}

func (d *DemoTableIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	if chartsData.GetEntryCount(config.ActivePair) < 2 {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	current := chartsData.GetEntry(config.ActivePair, 0)
	previous := chartsData.GetEntry(config.ActivePair, 1)

	closeColor := func(e *sources.ChartEntry) string {
		if e.Close >= e.Open {
			return "#26a69a"
		}
		return "#ef5350"
	}

	coll := indicators.NewPlotPointCollection(
		// Row 0: current
		indicators.PlotPoint{Value: current.Open, Color: "#ffffff"},
		indicators.PlotPoint{Value: current.Close, Color: closeColor(current)},
		// Row 1: previous
		indicators.PlotPoint{Value: previous.Open, Color: "#ffffff"},
		indicators.PlotPoint{Value: previous.Close, Color: closeColor(previous)},
	)
	handler.SetOutput(coll)
	return handler, nil
}

// ---------------------------------------------------------------------------
// PieChart – Market Snapshot (3 slices, main chart)
// ---------------------------------------------------------------------------

func GetDemoPieChartIndicator() indicators.BaseIndicator {
	return &DemoPieChartIndicator{}
}

type DemoPieChartIndicator struct{}

func (d *DemoPieChartIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypePieChart
}

func (d *DemoPieChartIndicator) GetName() string {
	return "Market Snapshot (Demo)"
}

func (d *DemoPieChartIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "",
		PieChartIndicator: indicators.PieChartIndicatorConfig{
			DisplayName: "Market Snapshot",
			SliceTitles: []string{"Open", "Hi-Close Range", "Volume (k)"},
		},
	}
}

func (d *DemoPieChartIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	entry := chartsData.GetEntry(config.ActivePair, 0)
	if entry == nil || !entry.IsValid() {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	v0 := math.Max(0, entry.Open)
	v1 := math.Max(0, math.Abs(entry.High-entry.Close))
	v2 := math.Max(0, entry.Volume/1000)

	if v0 == 0 && v1 == 0 && v2 == 0 {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	handler.SetOutput(indicators.NewPlotPointCollection(
		indicators.PlotPoint{Value: v0, Color: "#42a5f5"},
		indicators.PlotPoint{Value: v1, Color: "#66bb6a"},
		indicators.PlotPoint{Value: v2, Color: "#ffa726"},
	))
	return handler, nil
}

// ---------------------------------------------------------------------------
// PieChart 2 – Candle Structure (3 slices, separate chart group)
// ---------------------------------------------------------------------------

func GetDemoPieChart2Indicator() indicators.BaseIndicator {
	return &DemoPieChart2Indicator{}
}

type DemoPieChart2Indicator struct{}

func (d *DemoPieChart2Indicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypePieChart
}

func (d *DemoPieChart2Indicator) GetName() string {
	return "Candle Structure (Demo)"
}

func (d *DemoPieChart2Indicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "demo-pie2",
		PieChartIndicator: indicators.PieChartIndicatorConfig{
			DisplayName: "Candle Structure",
			SliceTitles: []string{"Upper Shadow", "Body", "Lower Shadow"},
		},
	}
}

func (d *DemoPieChart2Indicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	entry := chartsData.GetEntry(config.ActivePair, 0)
	if entry == nil || !entry.IsValid() {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	upperShadow := math.Max(0, entry.High-math.Max(entry.Open, entry.Close))
	body := math.Max(0, math.Abs(entry.Open-entry.Close))
	lowerShadow := math.Max(0, math.Min(entry.Open, entry.Close)-entry.Low)

	if upperShadow == 0 && body == 0 && lowerShadow == 0 {
		handler.SetOutput(indicators.NewInvalidPlotPointCollection())
		return handler, nil
	}

	handler.SetOutput(indicators.NewPlotPointCollection(
		indicators.PlotPoint{Value: upperShadow, Color: "#ef5350"},
		indicators.PlotPoint{Value: body, Color: "#29b6f6"},
		indicators.PlotPoint{Value: lowerShadow, Color: "#66bb6a"},
	))
	return handler, nil
}
