package sources

import "time"

type TradingPair struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Base   string `json:"base"`
	Quote  string `json:"quote"`
	Status string `json:"status"`
}

type ChartEntry struct {
	Timestamp time.Time `json:"timestamp"`

	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// TODO - Add extra methods and fields as needed to display the charts properly
type Source interface {
	GetName() string
	GetTradingPairs() []TradingPair
	GetValidIntervals() []string
	GetIntervalSeconds(interval string) []int64 // Required to calculate time ranges for each source correctly
	GetData(pairId string, interval string, start time.Time, end time.Time) ([]ChartEntry, error)
}
