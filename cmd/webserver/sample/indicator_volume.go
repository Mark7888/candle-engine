package sample

import (
	"testcharting/pkg/charting/sources"
	"testcharting/pkg/indicators"
)

func GetSampleVolumeIndicator() indicators.BaseIndicator {
	return &VolumeIndicator{}
}

type VolumeIndicator struct {
}

func (r *VolumeIndicator) GetType() indicators.IndicatorType {
	return indicators.IndicatorTypeBarChart
}

func (r *VolumeIndicator) GetName() string {
	return "Relative Strength Index (RSI)"
}

func (r *VolumeIndicator) GetConfig() indicators.IndicatorConfig {
	return indicators.IndicatorConfig{
		TargetChart: "1_volume", // empty means main chart
		BarChartIndicator: indicators.BarChartIndicatorConfig{
			BarTitles: []string{"Volume"},
		},
	}
}

func (r *VolumeIndicator) HandleEntry(points indicators.PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config indicators.ChartConfig, handler indicators.IndicatorOutputHandler) (indicators.IndicatorOutputHandler, error) {
	point := indicators.PlotPoint{
		Value: chartsData.GetEntry(config.ActivePair, 0).Volume,
		Color: "#fa5bff",
	}
	handler.SetSingleOutput(point)
	return handler, nil
}
