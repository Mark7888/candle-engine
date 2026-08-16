package registry

import (
	"testing"
	"time"

	"candle-engine/pkg/sources"
)

// stubSource serves a fixed set of entries per pair.
type stubSource struct {
	name string
	data map[string][]sources.ChartEntry
}

func (s *stubSource) GetName() string                        { return s.name }
func (s *stubSource) GetTradingPairs() []sources.TradingPair { return nil }
func (s *stubSource) GetValidTimeFrames() []string           { return []string{"1h"} }
func (s *stubSource) GetTimeFrameSeconds(string) int64       { return 3600 }
func (s *stubSource) GetTimeFrameString(int64) string        { return "1h" }

func (s *stubSource) GetData(pairId, _ string, tr sources.TimeRange) ([]sources.ChartEntry, error) {
	var out []sources.ChartEntry
	for _, e := range s.data[pairId] {
		if !e.Timestamp.Before(tr.Start) && !e.Timestamp.After(tr.End) {
			out = append(out, e)
		}
	}
	return out, nil
}

func hour(n int) time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Hour)
}

func entries(fromHour, toHour int, price float64) []sources.ChartEntry {
	var out []sources.ChartEntry
	for h := fromHour; h <= toHour; h++ {
		out = append(out, sources.ChartEntry{
			Timestamp: hour(h),
			Open:      price,
			High:      price + 1,
			Low:       price - 1,
			Close:     price,
			Volume:    100,
		})
	}
	return out
}

// TestFetchSources_ShorterHistoryHasNoSentinelValues is the regression test for
// the chart drawing wick-only candles. A pair that starts trading later than
// the chart's other pair is padded with placeholder entries; when those kept
// their -999 sentinel, any indicator reading Close plotted it on the main
// chart's y-axis and squashed every real candle into a thin line.
func TestFetchSources_ShorterHistoryHasNoSentinelValues(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterDataSources(&stubSource{
		name: "Stub",
		data: map[string][]sources.ChartEntry{
			"OLD": entries(0, 9, 100), // full history
			"NEW": entries(5, 9, 200), // listed halfway through
		},
	})

	pairs := []sources.SourcePairId{
		{SourceName: "Stub", PairId: "OLD"},
		{SourceName: "Stub", PairId: "NEW"},
	}
	timeRange := sources.TimeRange{Start: hour(0), End: hour(9)}

	aligned, width, err := reg.FetchSources(timeRange, pairs, "1h")
	if err != nil {
		t.Fatalf("FetchSources failed: %v", err)
	}
	if width != 10 {
		t.Fatalf("expected 10 aligned entries, got %d", width)
	}

	newPair := aligned[sources.SourcePairId{SourceName: "Stub", PairId: "NEW"}]
	if len(newPair) != 10 {
		t.Fatalf("expected the short pair to be padded to 10 entries, got %d", len(newPair))
	}

	for i, e := range newPair {
		if e.Open < 0 || e.High < 0 || e.Low < 0 || e.Close < 0 || e.Volume < 0 {
			t.Errorf("entry %d leaks a sentinel value into chart data: %+v", i, e)
		}
	}

	// The padded entries must still be identifiable as placeholders so
	// validity-aware indicators can skip them and the candlestick renders a gap.
	for i := 0; i < 5; i++ {
		if !newPair[i].Invalid {
			t.Errorf("padded entry %d should be flagged Invalid", i)
		}
		if newPair[i].IsValid() {
			t.Errorf("padded entry %d must not report itself as valid", i)
		}
	}
	for i := 5; i < 10; i++ {
		if newPair[i].Invalid || !newPair[i].IsValid() {
			t.Errorf("real entry %d should be valid: %+v", i, newPair[i])
		}
	}
}
