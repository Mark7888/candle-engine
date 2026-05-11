package sample

import (
	"candle-engine/pkg/chartproject"
	"candle-engine/pkg/sources"
	"time"
)

func GetSampleChartProject() *chartproject.ChartProject {
	allPairsToAdd := []sources.SourcePairId{
		{SourceName: "Coinbase", PairId: "BTC-USDT"},
		{SourceName: "Coinbase", PairId: "ETH-USDT"},
		{SourceName: "Binance", PairId: "BTCUSDC"},
		{SourceName: "Binance", PairId: "ETHUSDC"},
	}

	startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")

	project := chartproject.NewChartProject("test-project-01", "Sample Project", chartproject.ChartProjectConfig{
		TimeFrame:    "1d",
		TimeRange:    sources.TimeRange{Start: startTime, End: endTime},
		CandleColors: sources.CandleColorConfig{BullishColor: "#ffffff", BearishColor: "#00c8ff"},
	})
	for _, pair := range allPairsToAdd {
		project.AddPair(pair)
	}
	project.AddIndicatorByName(GetSampleRSIIndicator().GetName())
	project.AddIndicatorByName(GetSampleKAMAIndicator().GetName())
	project.AddIndicatorByName(GetSampleVolumeIndicator().GetName())

	return project
}

func GetSampleChartProject2() *chartproject.ChartProject {
	allPairsToAdd := []sources.SourcePairId{
		{SourceName: "Binance", PairId: "BTCUSDC"},
		{SourceName: "Binance", PairId: "ETHUSDC"},
	}

	startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")

	project := chartproject.NewChartProject("test-project-02", "Second Project MA Only", chartproject.ChartProjectConfig{
		TimeFrame:    "1d",
		TimeRange:    sources.TimeRange{Start: startTime, End: endTime},
		CandleColors: sources.CandleColorConfig{BullishColor: "#ffff51", BearishColor: "#b872ff"},
	})
	for _, pair := range allPairsToAdd {
		project.AddPair(pair)
	}
	project.AddIndicatorByName(GetSampleMAIndicator().GetName())
	project.AddIndicatorByName(GetSampleVolumeIndicator().GetName())

	return project
}
