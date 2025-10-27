package sources

import "time"

type TradingPair struct {
	Id     string
	Name   string
	Base   string
	Quote  string
	Status string
}

type ChartEntry struct {
	Timestamp time.Time

	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// TODO - Add extra methods and fields as needed to display the charts properly
type Source interface {
	GetName() string
	GetTradingPairs() []TradingPair
	GetValidIntervals() []string
	GetIntervalSeconds(interval string) []int64 // Required to calculate time ranges for each source correctly
	GetData(pairId string, interval string, start time.Time, end time.Time) ([]ChartEntry, error)
}
