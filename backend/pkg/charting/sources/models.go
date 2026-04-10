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

func (c ChartEntry) IsValid() bool {
	return c.Open >= 0 && c.High >= 0 && c.Low >= 0 && c.Close >= 0 && c.Volume >= 0
}

// ForwardFillInvalidEntries fills OHLCV values of invalid entries from the last valid entry.
func ForwardFillInvalidEntries(entries []ChartEntry) []ChartEntry {
	result := make([]ChartEntry, len(entries))
	copy(result, entries)

	lastValidIdx := -1
	for i := range result {
		if !result[i].Invalid {
			lastValidIdx = i
		} else if lastValidIdx >= 0 {
			result[i].Open = result[lastValidIdx].Open
			result[i].High = result[lastValidIdx].High
			result[i].Low = result[lastValidIdx].Low
			result[i].Close = result[lastValidIdx].Close
			result[i].Volume = result[lastValidIdx].Volume
		}
	}
	return result
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

type CandleColorConfig struct {
	BullishColor string
	BearishColor string
}
