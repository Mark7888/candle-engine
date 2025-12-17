package sample

import (
	"testcharting/pkg/charting/cachedsource"
	"testcharting/pkg/charting/chartproject"
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/charting/sources/binance"
	"testcharting/pkg/charting/sources/coinbase"
	"testcharting/pkg/indicators"
	"time"
)

func GetSampleChartProject() *chartproject.ChartProject {
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
		SourceName: "Binance",
		PairId:     "BTC-USDC",
	}
	allPairsToAdd := []sources.SourcePairId{
		// {SourceName: "Coinbase", PairId: "BTC-USD"},
		// {SourceName: "Coinbase", PairId: "ETH-USD"},
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
		// GetSampleMACDIndicator(),
		// GetSampleSMAIndicator(),
		// GetSampleEMAIndicator(),
		// GetSampleBollingerBandsIndicator(),
		// GetSampleStochasticIndicator(),
		// GetSampleATRIndicator(),
	}

	chartProject1 := chartproject.NewChartProject("Sample Project")
	chartProject1.AddSources(sourcesToAdd...)
	chartProject1.AddTradingPairs(allPairsToAdd...)
	chartProject1.SetActivePair(activePair)
	chartProject1.SetTimeFrame("1d")
	chartProject1.SetTimeRange(timeRange)
	chartProject1.AddIndicators(indicatorsToAdd...)

	return chartProject1
}
