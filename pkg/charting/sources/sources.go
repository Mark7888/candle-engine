package sources

type Source interface {
	GetName() string
	GetTradingPairs() []TradingPair
	GetValidTimeFrames() []string
	GetTimeFrameSeconds(timeFrame string) int64 // Required to calculate time ranges for each source correctly
	GetData(pairId string, timeFrame string, timeRange TimeRange) ([]ChartEntry, error)
}
