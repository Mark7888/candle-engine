package sample

import (
	"candle-engine/pkg/charting/cachedsource"
	"candle-engine/pkg/charting/chartproject"
	"candle-engine/pkg/charting/sources"
	"candle-engine/pkg/charting/sources/binance"
	"candle-engine/pkg/charting/sources/coinbase"
	"candle-engine/pkg/indicators"
	"time"
)

func GetSampleChartProject() chartproject.ChartProject {
	// Sources and pairs to add
	sourceCoinbase := coinbase.NewCoinbaseSource()
	sourceBinance := binance.NewBinanceSource()
	cachedSourceCoinbase := cachedsource.NewCachedSource(sourceCoinbase, "./tmp/coinbase_cache.json")
	cachedSourceBinance := cachedsource.NewCachedSource(sourceBinance, "./tmp/binance_cache.json")
	sourcesToAdd := []sources.Source{
		cachedSourceCoinbase,
		cachedSourceBinance,
	}
	activePair := sources.SourcePairId{
		SourceName: "Coinbase",
		PairId:     "BTC-USDT",
	}
	allPairsToAdd := []sources.SourcePairId{
		{SourceName: "Coinbase", PairId: "BTC-USDT"},
		{SourceName: "Coinbase", PairId: "ETH-USDT"},
		{SourceName: "Binance", PairId: "BTCUSDC"},
		{SourceName: "Binance", PairId: "ETHUSDC"},
	}

	// Time range to set
	startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")
	timeRange := sources.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	// Indicators to add
	indicatorsToAdd := []indicators.BaseIndicator{
		GetSampleRSIIndicator(),
		GetSampleKAMAIndicator(),
		GetSampleVolumeIndicator(),
		// GetSampleMACDIndicator(),
		// GetSampleSMAIndicator(),
		// GetSampleEMAIndicator(),
		// GetSampleBollingerBandsIndicator(),
		// GetSampleStochasticIndicator(),
		// GetSampleATRIndicator(),
	}

	chartProject1 := chartproject.NewChartProject("Sample Project")
	chartProject1.SetCandleColors(sources.CandleColorConfig{BullishColor: "#ffffff", BearishColor: "#00c8ff"})
	chartProject1.AddSources(sourcesToAdd...)
	chartProject1.AddTradingPairs(allPairsToAdd...)
	chartProject1.SetActivePair(activePair)
	chartProject1.SetTimeFrame("1d")
	chartProject1.SetTimeRange(timeRange)
	chartProject1.AddIndicators(indicatorsToAdd...)

	return *chartProject1
}

func GetSampleChartProject2() chartproject.ChartProject {
	// Sources and pairs to add
	sourceBinance := binance.NewBinanceSource()
	cachedSourceBinance := cachedsource.NewCachedSource(sourceBinance, "./tmp/binance_cache.json")
	sourcesToAdd := []sources.Source{
		cachedSourceBinance,
	}
	activePair := sources.SourcePairId{
		SourceName: "Binance",
		PairId:     "BTCUSDC",
	}
	allPairsToAdd := []sources.SourcePairId{
		{SourceName: "Binance", PairId: "BTCUSDC"},
		{SourceName: "Binance", PairId: "ETHUSDC"},
	}

	// Time range to set
	startTime, _ := time.Parse(time.RFC3339, "2024-12-01T00:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2025-12-01T00:00:00Z")
	timeRange := sources.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	// Indicators to add
	indicatorsToAdd := []indicators.BaseIndicator{
		GetSampleMAIndicator(),
		GetSampleVolumeIndicator(),
	}

	chartProject1 := chartproject.NewChartProject("Second Project MA Only")
	chartProject1.SetCandleColors(sources.CandleColorConfig{BullishColor: "#ffff51", BearishColor: "#b872ff"})
	chartProject1.AddSources(sourcesToAdd...)
	chartProject1.AddTradingPairs(allPairsToAdd...)
	chartProject1.SetActivePair(activePair)
	chartProject1.SetTimeFrame("1d")
	chartProject1.SetTimeRange(timeRange)
	chartProject1.AddIndicators(indicatorsToAdd...)

	return *chartProject1
}
