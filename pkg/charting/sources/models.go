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

func GenerateInvalidEntry(timestamp time.Time) ChartEntry {
	return ChartEntry{
		Timestamp: timestamp,
		Open:      -999,
		High:      -999,
		Low:       -999,
		Close:     -999,
		Volume:    -999,
	}
}

func (c ChartEntry) IsValid() bool {
	return c.Open >= 0 && c.High >= 0 && c.Low >= 0 && c.Close >= 0 && c.Volume >= 0
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type SourcePairId struct {
	SourceName string
	PairId     string
}

func (s SourcePairId) String() string {
	return s.SourceName + ":" + s.PairId
}
