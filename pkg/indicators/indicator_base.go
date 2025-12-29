package indicators

import "testcharting/pkg/charting/sources"

type BaseIndicator interface {
	GetName() string
	GetType() IndicatorType
	GetConfig() IndicatorConfig
	HandleEntry(points PlotPointCollectionHistory, chartsData sources.MultiSourceEntries, config ChartConfig, handler IndicatorOutputHandler) (IndicatorOutputHandler, error)
}
