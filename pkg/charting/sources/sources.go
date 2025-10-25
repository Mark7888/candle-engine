package sources

import "time"

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
	GetValidIntervals() []string
	GetData(interval string) ([]ChartEntry, error)
}
