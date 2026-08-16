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

	Invalid bool `json:"invalid"`
}

func GenerateInvalidEntry(timestamp time.Time) ChartEntry {
	return ChartEntry{
		Timestamp: timestamp,
		Open:      -999,
		High:      -999,
		Low:       -999,
		Close:     -999,
		Volume:    -999,
		Invalid:   true,
	}
}

// IsValid reports whether the entry holds real market data.
//
// Entries that were only produced to align this pair with another pair's
// timestamps are never valid, even after FillInvalidEntries has replaced their
// sentinel numbers with plausible neighbouring values – the Invalid flag is the
// authoritative signal, the value check only guards against sentinels that were
// never filled.
func (c ChartEntry) IsValid() bool {
	if c.Invalid {
		return false
	}
	return c.Open >= 0 && c.High >= 0 && c.Low >= 0 && c.Close >= 0 && c.Volume >= 0
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type SourcePairId struct {
	SourceName string `json:"sourceName"`
	PairId     string `json:"pairId"`
}

func (s SourcePairId) String() string {
	return s.SourceName + ":" + s.PairId
}

type CandleColorConfig struct {
	BullishColor string `json:"bullishColor"`
	BearishColor string `json:"bearishColor"`
}
