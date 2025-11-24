package sources

type Source interface {
	GetName() string
	GetTradingPairs() []TradingPair
	GetValidIntervals() []string
	GetIntervalSeconds(interval string) []int64 // Required to calculate time ranges for each source correctly
	GetData(pairId string, interval string, timeRange TimeRange) ([]ChartEntry, error)
}
