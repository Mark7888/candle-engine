package sample

import (
"candle-engine/pkg/charting/cachedsource"
"candle-engine/pkg/charting/chartproject"
"candle-engine/pkg/charting/sources"
"candle-engine/pkg/charting/sources/binance"
"candle-engine/pkg/indicators"
"math"
"time"
)

func GetDemoChartProject() chartproject.ChartProject {
sourceBinance := binance.NewBinanceSource()
cachedSourceBinance := cachedsource.NewCachedSource(sourceBinance, "./tmp/binance_cache.json")

activePair := sources.SourcePairId{SourceName: "Binance", PairId: "BTCUSDC"}
allPairs := []sources.SourcePairId{
{SourceName: "Binance", PairId: "BTCUSDC"},
{SourceName: "Binance", PairId: "ETHUSDC"},
}

startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")

project := chartproject.NewChartProject("Demo Project")
project.SetCandleColors(sources.CandleColorConfig{BullishColor: "#26a69a", BearishColor: "#ef5350"})
project.AddSources(cachedSourceBinance)
project.AddTradingPairs(allPairs...)
project.SetActivePair(activePair)
project.SetTimeFrame("1d")
project.SetTimeRange(sources.TimeRange{Start: startTime, End: endTime})
project.AddIndicators(
GetDemoLineIndicator(),
GetDemoStepLineIndicator(),
GetDemoBarChartIndicator(),
GetDemoAreaChartIndicator(),
GetDemoCandleStickIndicator(),
GetDemoTableIndicator(),
GetDemoPieChartIndicator(),
)
return *project
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
// BarChart – Volume
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
handler.SetSingleOutput(indicators.PlotPoint{
Value: chartsData.GetEntry(config.ActivePair, 0).Volume,
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
// Table – OHLCV Stats (2x3)
// ---------------------------------------------------------------------------

func GetDemoTableIndicator() indicators.BaseIndicator {
return &DemoTableIndicator{}
}

type DemoTableIndicator struct{}

func (d *DemoTableIndicator) GetType() indicators.IndicatorType {
return indicators.IndicatorTypeTable
}

func (d *DemoTableIndicator) GetName() string {
return "OHLCV Stats (Demo)"
}

func (d *DemoTableIndicator) GetConfig() indicators.IndicatorConfig {
return indicators.IndicatorConfig{
TargetChart: "",
TableIndicator: indicators.TableIndicatorConfig{
DisplayName:        "OHLCV Stats",
RowTitles:          []string{"Current", "Previous"},
ColumnTitles:       []string{"Open", "Hi-Lo Spread", "Close"},
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
indicators.PlotPoint{Value: current.High - current.Low, Color: "#26a69a"},
indicators.PlotPoint{Value: current.Close, Color: closeColor(current)},
// Row 1: previous
indicators.PlotPoint{Value: previous.Open, Color: "#ffffff"},
indicators.PlotPoint{Value: previous.High - previous.Low, Color: "#26a69a"},
indicators.PlotPoint{Value: previous.Close, Color: closeColor(previous)},
)
handler.SetOutput(coll)
return handler, nil
}

// ---------------------------------------------------------------------------
// PieChart – Market Snapshot (3 slices)
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
