package main

import (
	"candle-engine/cmd/webserver/sample"
	"candle-engine/pkg/cachedsource"
	"candle-engine/pkg/chartmanager"
	"candle-engine/pkg/registry"
	"candle-engine/pkg/sources/binance"
	"candle-engine/pkg/sources/coinbase"
	"candle-engine/pkg/webserver"
)

func main() {
	reg := registry.NewRegistry()

	sourceCoinbase := coinbase.NewCoinbaseSource()
	sourceBinance := binance.NewBinanceSource()
	reg.RegisterDataSources(
		cachedsource.NewCachedSource(sourceCoinbase, "./tmp/coinbase_cache.json"),
		cachedsource.NewCachedSource(sourceBinance, "./tmp/binance_cache.json"),
	)
	reg.RegisterIndicators(
		sample.GetSampleRSIIndicator(),
		sample.GetSampleKAMAIndicator(),
		sample.GetSampleVolumeIndicator(),
		sample.GetSampleMAIndicator(),
		sample.GetDemoLineIndicator(),
		sample.GetDemoStepLineIndicator(),
		sample.GetDemoBarChartIndicator(),
		sample.GetDemoAreaChartIndicator(),
		sample.GetDemoCandleStickIndicator(),
		sample.GetDemoTableIndicator(),
		sample.GetDemoPieChartIndicator(),
		sample.GetDemoPieChart2Indicator(),
	)

	cm := chartmanager.InitChartManager(reg)
	cm.AddChartProject(sample.GetSampleChartProject())
	cm.AddChartProject(sample.GetSampleChartProject2())
	cm.AddChartProject(sample.GetDemoChartProject())

	webserver.InitServer()
}
