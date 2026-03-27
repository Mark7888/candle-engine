package cachedsource

import (
	"candle-engine/pkg/charting/sources"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------
// Mock source
// -----------------------------------------------------------------------------

// mockSource is a controllable sources.Source for testing.
type mockSource struct {
	mu        sync.Mutex
	callCount int32 // atomic
	// handler is called for every GetData invocation.
	// If nil, it returns the entries from dataStore that fall within the range.
	handler   func(pairId, timeFrame string, tr sources.TimeRange) ([]sources.ChartEntry, error)
	dataStore []sources.ChartEntry // used when handler is nil
	tfSeconds map[string]int64
}

func newMockSource(tfSeconds map[string]int64, entries []sources.ChartEntry) *mockSource {
	return &mockSource{tfSeconds: tfSeconds, dataStore: entries}
}

func (m *mockSource) GetName() string                            { return "mock" }
func (m *mockSource) GetTradingPairs() []sources.TradingPair    { return nil }
func (m *mockSource) GetValidTimeFrames() []string              { return nil }
func (m *mockSource) GetTimeFrameString(secs int64) string      { return "" }
func (m *mockSource) GetTimeFrameSeconds(tf string) int64 {
	if m.tfSeconds != nil {
		if v, ok := m.tfSeconds[tf]; ok {
			return v
		}
	}
	return 3600 // default: 1 h
}

func (m *mockSource) GetData(pairId, timeFrame string, tr sources.TimeRange) ([]sources.ChartEntry, error) {
	atomic.AddInt32(&m.callCount, 1)
	if m.handler != nil {
		return m.handler(pairId, timeFrame, tr)
	}
	return filterEntries(m.dataStore, tr), nil
}

func (m *mockSource) calls() int { return int(atomic.LoadInt32(&m.callCount)) }

// filterEntries returns entries whose Timestamp is within [tr.Start, tr.End].
func filterEntries(entries []sources.ChartEntry, tr sources.TimeRange) []sources.ChartEntry {
	var out []sources.ChartEntry
	for _, e := range entries {
		if !e.Timestamp.Before(tr.Start) && !e.Timestamp.After(tr.End) {
			out = append(out, e)
		}
	}
	return out
}

// -----------------------------------------------------------------------------
// Test helpers
// -----------------------------------------------------------------------------

// hour returns a UTC time that is base + n hours.
var base = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func hour(n int) time.Time { return base.Add(time.Duration(n) * time.Hour) }

func tr(startH, endH int) sources.TimeRange {
	return sources.TimeRange{Start: hour(startH), End: hour(endH)}
}

// makeEntries builds one ChartEntry per hour in [startH, endH].
func makeEntries(startH, endH int) []sources.ChartEntry {
	var out []sources.ChartEntry
	for h := startH; h <= endH; h++ {
		out = append(out, sources.ChartEntry{
			Timestamp: hour(h),
			Open:      float64(h),
			High:      float64(h) + 0.5,
			Low:       float64(h) - 0.5,
			Close:     float64(h) + 0.1,
			Volume:    float64(h) * 10,
		})
	}
	return out
}

// mustTimestamps returns the hour-offsets of the entries' timestamps.
func mustTimestamps(entries []sources.ChartEntry) []int {
	var out []int
	for _, e := range entries {
		out = append(out, int(e.Timestamp.Sub(base).Hours()))
	}
	return out
}

// newCS builds a CachedSource backed by mock and (optionally) a temp cache file.
func newCS(mock *mockSource, useFile bool, t *testing.T) sources.Source {
	t.Helper()
	filePath := ""
	if useFile {
		filePath = filepath.Join(t.TempDir(), "cache.json")
	}
	return NewCachedSource(mock, filePath)
}

// -----------------------------------------------------------------------------
// Unit tests for pure helper functions
// -----------------------------------------------------------------------------

func TestCovers(t *testing.T) {
	tests := []struct {
		name    string
		covered []interval
		want    sources.TimeRange
		expect  bool
	}{
		{"empty covered", nil, tr(0, 5), false},
		{"exact match", []interval{{hour(0), hour(5)}}, tr(0, 5), true},
		{"superset", []interval{{hour(0), hour(10)}}, tr(2, 8), true},
		{"gap in middle", []interval{{hour(0), hour(3)}, {hour(5), hour(10)}}, tr(0, 10), false},
		{"two segments fully cover", []interval{{hour(0), hour(5)}, {hour(5), hour(10)}}, tr(0, 10), true},
		{"want before covered", []interval{{hour(5), hour(10)}}, tr(0, 4), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := covers(tt.covered, tt.want); got != tt.expect {
				t.Errorf("covers() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestMissingRanges(t *testing.T) {
	tests := []struct {
		name    string
		covered []interval
		want    sources.TimeRange
		missing []sources.TimeRange
	}{
		{
			"nothing covered",
			nil, tr(0, 10),
			[]sources.TimeRange{tr(0, 10)},
		},
		{
			"fully covered",
			[]interval{{hour(0), hour(10)}}, tr(2, 8),
			nil,
		},
		{
			"gap at start",
			[]interval{{hour(5), hour(10)}}, tr(0, 10),
			[]sources.TimeRange{tr(0, 5)},
		},
		{
			"gap at end",
			[]interval{{hour(0), hour(5)}}, tr(0, 10),
			[]sources.TimeRange{tr(5, 10)},
		},
		{
			"gap in middle",
			[]interval{{hour(0), hour(3)}, {hour(7), hour(10)}}, tr(0, 10),
			[]sources.TimeRange{tr(3, 7)},
		},
		{
			"two gaps",
			[]interval{{hour(2), hour(4)}, {hour(6), hour(8)}}, tr(0, 10),
			[]sources.TimeRange{tr(0, 2), tr(4, 6), tr(8, 10)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := missingRanges(tt.covered, tt.want)
			if len(got) != len(tt.missing) {
				t.Fatalf("got %d missing ranges, want %d: %v", len(got), len(tt.missing), got)
			}
			for i, r := range got {
				if !r.Start.Equal(tt.missing[i].Start) || !r.End.Equal(tt.missing[i].End) {
					t.Errorf("range[%d]: got %v–%v, want %v–%v",
						i, r.Start, r.End, tt.missing[i].Start, tt.missing[i].End)
				}
			}
		})
	}
}

func TestAddCoveredInterval(t *testing.T) {
	tests := []struct {
		name    string
		covered []interval
		add     interval
		expect  []interval
	}{
		{
			"add to empty",
			nil, interval{hour(2), hour(5)},
			[]interval{{hour(2), hour(5)}},
		},
		{
			"prepend",
			[]interval{{hour(5), hour(8)}}, interval{hour(0), hour(3)},
			[]interval{{hour(0), hour(3)}, {hour(5), hour(8)}},
		},
		{
			"append",
			[]interval{{hour(0), hour(3)}}, interval{hour(5), hour(8)},
			[]interval{{hour(0), hour(3)}, {hour(5), hour(8)}},
		},
		{
			"overlap merges",
			[]interval{{hour(0), hour(5)}, {hour(7), hour(10)}}, interval{hour(3), hour(8)},
			[]interval{{hour(0), hour(10)}},
		},
		{
			"touching merges",
			[]interval{{hour(0), hour(5)}}, interval{hour(5), hour(10)},
			[]interval{{hour(0), hour(10)}},
		},
		{
			"inside existing – no change to bounds",
			[]interval{{hour(0), hour(10)}}, interval{hour(3), hour(7)},
			[]interval{{hour(0), hour(10)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addCoveredInterval(tt.covered, tt.add)
			if len(got) != len(tt.expect) {
				t.Fatalf("got %d intervals, want %d: %v", len(got), len(tt.expect), got)
			}
			for i, iv := range got {
				if !iv.Start.Equal(tt.expect[i].Start) || !iv.End.Equal(tt.expect[i].End) {
					t.Errorf("interval[%d]: got %v–%v, want %v–%v",
						i, iv.Start, iv.End, tt.expect[i].Start, tt.expect[i].End)
				}
			}
		})
	}
}

func TestMergeEntries(t *testing.T) {
	a := makeEntries(0, 3)
	b := makeEntries(2, 5) // overlapping timestamps 2–3

	// Overwrite entry at hour(2) with different Close value to verify precedence.
	b[0].Close = 999

	merged := mergeEntries(a, b)
	if len(merged) != 6 { // 0,1,2,3,4,5
		t.Fatalf("expected 6 merged entries, got %d", len(merged))
	}
	if merged[2].Close != 999 {
		t.Errorf("new entry did not overwrite existing: Close=%v", merged[2].Close)
	}
	// Verify sorted order
	for i := 1; i < len(merged); i++ {
		if !merged[i].Timestamp.After(merged[i-1].Timestamp) {
			t.Errorf("entries not sorted at index %d", i)
		}
	}
}

func TestMergeNearbyRanges(t *testing.T) {
	tfSecs := int64(3600) // 1 hour candles; maxGapCandles=10 → 10 h max gap

	// Two segments with a 5-hour gap → should merge
	missing := []sources.TimeRange{tr(0, 3), tr(8, 12)}
	merged := mergeNearbyRanges(missing, tfSecs)
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged range, got %d: %v", len(merged), merged)
	}
	if !merged[0].Start.Equal(hour(0)) || !merged[0].End.Equal(hour(12)) {
		t.Errorf("merged range wrong: %v–%v", merged[0].Start, merged[0].End)
	}

	// Two segments with a 15-hour gap → should NOT merge
	missing2 := []sources.TimeRange{tr(0, 3), tr(18, 22)}
	merged2 := mergeNearbyRanges(missing2, tfSecs)
	if len(merged2) != 2 {
		t.Fatalf("expected 2 separate ranges, got %d: %v", len(merged2), merged2)
	}
}

func TestExtractEntries(t *testing.T) {
	entries := makeEntries(0, 9)

	got := extractEntries(entries, tr(3, 6))
	ts := mustTimestamps(got)
	want := []int{3, 4, 5, 6}
	if len(ts) != len(want) {
		t.Fatalf("got %v, want %v", ts, want)
	}
	for i, v := range ts {
		if v != want[i] {
			t.Errorf("ts[%d]: got %d, want %d", i, v, want[i])
		}
	}
}

// -----------------------------------------------------------------------------
// Integration tests for GetData
// -----------------------------------------------------------------------------

func TestGetData_NoCacheHit(t *testing.T) {
	mock := newMockSource(nil, makeEntries(0, 9))
	cs := newCS(mock, false, t)

	got, err := cs.GetData("BTC", "1h", tr(0, 9))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 10 {
		t.Errorf("expected 10 entries, got %d", len(got))
	}
	if mock.calls() != 1 {
		t.Errorf("expected 1 source call, got %d", mock.calls())
	}
}

func TestGetData_FullCacheHit(t *testing.T) {
	mock := newMockSource(nil, makeEntries(0, 9))
	cs := newCS(mock, false, t)

	// Prime the cache
	if _, err := cs.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}

	// Second call – must NOT hit the source again
	got, err := cs.GetData("BTC", "1h", tr(2, 7))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 6 {
		t.Errorf("expected 6 entries, got %d", len(got))
	}
	if mock.calls() != 1 {
		t.Errorf("expected 1 source call total, got %d", mock.calls())
	}
}

func TestGetData_PartialCacheHit(t *testing.T) {
	mock := newMockSource(nil, makeEntries(0, 19))
	cs := newCS(mock, false, t)

	// Cache hours [0, 9]
	if _, err := cs.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}
	// Request [5, 14] – hours 10-14 are missing
	got, err := cs.GetData("BTC", "1h", tr(5, 14))
	if err != nil {
		t.Fatal(err)
	}
	ts := mustTimestamps(got)
	if len(ts) != 10 {
		t.Errorf("expected 10 entries (5-14), got %d: %v", len(ts), ts)
	}
	if mock.calls() != 2 {
		t.Errorf("expected 2 source calls, got %d", mock.calls())
	}
}

func TestGetData_MultipleGaps(t *testing.T) {
	// Hours 0-19 exist; we'll cache [0,4] and [10,14], leaving [5,9] and [15,19] as gaps.
	mock := newMockSource(nil, makeEntries(0, 19))
	cs := newCS(mock, false, t)

	if _, err := cs.GetData("BTC", "1h", tr(0, 4)); err != nil {
		t.Fatal(err)
	}
	if _, err := cs.GetData("BTC", "1h", tr(10, 14)); err != nil {
		t.Fatal(err)
	}
	callsBefore := mock.calls()

	// Request [0,19] – should fetch [5,9] and [15,19] as separate calls
	// (gap between them is 5 hours = 5 candles, which is ≤ maxGapCandles=10,
	//  so the fetcher will actually MERGE those two missing segments into one call).
	got, err := cs.GetData("BTC", "1h", tr(0, 19))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 20 {
		t.Errorf("expected 20 entries, got %d", len(got))
	}
	newCalls := mock.calls() - callsBefore
	// The two gaps ([5,9] gap=5h, [15,19] gap between them=10h from end of [5,9] to start of [15,19])
	// will be merged into one fetch because the gap between missing[0].End(9) and missing[1].Start(15) is 6h ≤ 10h.
	if newCalls != 1 {
		t.Errorf("expected 1 merged fetch for the two gaps, got %d", newCalls)
	}
}

func TestGetData_LargeGapNotMerged(t *testing.T) {
	// Use 5-minute candles so maxGap = 10 * 5min = 50 minutes.
	// Cache [0h,4h], [10h,12h], [20h,24h] – this creates two missing segments:
	//   [4h,10h] (gap=6h) and [12h,20h] (gap=8h).
	// The gap between the two missing segments is 12h-10h = 2h > 50 min,
	// so they must NOT be merged into a single fetch call.
	tfSecs := map[string]int64{"5m": 300}
	mock := newMockSource(tfSecs, makeEntries(0, 24))
	cs := NewCachedSource(mock, "")

	for _, r := range []sources.TimeRange{tr(0, 4), tr(10, 12), tr(20, 24)} {
		if _, err := cs.GetData("BTC", "5m", r); err != nil {
			t.Fatal(err)
		}
	}
	callsBefore := mock.calls()

	_, err := cs.GetData("BTC", "5m", tr(0, 24))
	if err != nil {
		t.Fatal(err)
	}
	newCalls := mock.calls() - callsBefore
	if newCalls != 2 {
		t.Errorf("expected 2 separate fetches (gap too large to merge), got %d", newCalls)
	}
}

func TestGetData_EmptyResult(t *testing.T) {
	// Source returns no data (e.g. a weekend gap).
	mock := newMockSource(nil, nil)
	cs := newCS(mock, false, t)

	got, err := cs.GetData("BTC", "1h", tr(0, 5))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 entries, got %d", len(got))
	}

	// Second call for same range must NOT call source again (range is now covered).
	if _, err := cs.GetData("BTC", "1h", tr(0, 5)); err != nil {
		t.Fatal(err)
	}
	if mock.calls() != 1 {
		t.Errorf("expected 1 source call total (empty range still covered), got %d", mock.calls())
	}
}

func TestGetData_Error(t *testing.T) {
	boom := errors.New("source unavailable")
	mock := &mockSource{
		handler: func(_, _ string, _ sources.TimeRange) ([]sources.ChartEntry, error) {
			return nil, boom
		},
	}
	cs := NewCachedSource(mock, "")

	_, err := cs.GetData("BTC", "1h", tr(0, 5))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Errorf("expected wrapped boom error, got: %v", err)
	}
}

func TestGetData_DifferentPairs_Independent(t *testing.T) {
	mock := newMockSource(nil, makeEntries(0, 9))
	cs := newCS(mock, false, t)

	if _, err := cs.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}
	// ETH has no cache – must hit source
	if _, err := cs.GetData("ETH", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}
	if mock.calls() != 2 {
		t.Errorf("expected 2 source calls (one per pair), got %d", mock.calls())
	}
	// Repeat BTC – cache hit, no new call
	if _, err := cs.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}
	if mock.calls() != 2 {
		t.Errorf("expected still 2 source calls after BTC cache hit, got %d", mock.calls())
	}
}

// -----------------------------------------------------------------------------
// Concurrency tests
// -----------------------------------------------------------------------------

func TestGetData_ConcurrentSamePair(t *testing.T) {
	// All goroutines request the same range; the source should be called exactly once.
	var (
		fetchSem = make(chan struct{}, 1) // ensures only 1 fetch at a time
		ready    = make(chan struct{})
	)

	mock := &mockSource{
		handler: func(_, _ string, _ sources.TimeRange) ([]sources.ChartEntry, error) {
			// Signal that we're inside a fetch, then block until released.
			fetchSem <- struct{}{}
			<-ready
			return makeEntries(0, 9), nil
		},
	}
	cs := NewCachedSource(mock, "")

	const goroutines = 10
	var wg sync.WaitGroup
	results := make([][]sources.ChartEntry, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = cs.GetData("BTC", "1h", tr(0, 9))
		}(i)
	}

	// Wait until first fetch has started, then release it.
	<-fetchSem
	close(ready)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d error: %v", i, err)
		}
	}
	if mock.calls() != 1 {
		t.Errorf("source called %d times; want exactly 1", mock.calls())
	}
	for i, got := range results {
		if len(got) != 10 {
			t.Errorf("goroutine %d: expected 10 entries, got %d", i, len(got))
		}
	}
}

func TestGetData_ConcurrentDifferentPairs(t *testing.T) {
	mock := newMockSource(nil, makeEntries(0, 9))
	cs := newCS(mock, false, t)

	pairs := []string{"BTC", "ETH", "LTC", "XRP", "BNB"}
	var wg sync.WaitGroup

	for _, p := range pairs {
		wg.Add(1)
		go func(pairId string) {
			defer wg.Done()
			if _, err := cs.GetData(pairId, "1h", tr(0, 9)); err != nil {
				t.Errorf("pair %s error: %v", pairId, err)
			}
		}(p)
	}
	wg.Wait()

	if mock.calls() != len(pairs) {
		t.Errorf("expected %d source calls (one per pair), got %d", len(pairs), mock.calls())
	}

	// Second round – everything cached, no new calls.
	for _, p := range pairs {
		wg.Add(1)
		go func(pairId string) {
			defer wg.Done()
			if _, err := cs.GetData(pairId, "1h", tr(2, 7)); err != nil {
				t.Errorf("pair %s error: %v", pairId, err)
			}
		}(p)
	}
	wg.Wait()

	if mock.calls() != len(pairs) {
		t.Errorf("expected no additional source calls after cache warm, got %d total", mock.calls())
	}
}

func TestGetData_ConcurrentMixedPairsAndTimeframes(t *testing.T) {
	mock := newMockSource(map[string]int64{"1h": 3600, "4h": 14400}, makeEntries(0, 23))
	cs := newCS(mock, false, t)

	type job struct {
		pair string
		tf   string
	}
	jobs := []job{
		{"BTC", "1h"}, {"BTC", "4h"},
		{"ETH", "1h"}, {"ETH", "4h"},
	}

	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			if _, err := cs.GetData(j.pair, j.tf, tr(0, 23)); err != nil {
				t.Errorf("%s/%s: %v", j.pair, j.tf, err)
			}
		}(j)
	}
	wg.Wait()

	if mock.calls() != len(jobs) {
		t.Errorf("expected %d source calls, got %d", len(jobs), mock.calls())
	}
}

// -----------------------------------------------------------------------------
// File persistence tests
// -----------------------------------------------------------------------------

func TestPersistence_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cache.json")

	mock := newMockSource(nil, makeEntries(0, 9))
	cs1 := NewCachedSource(mock, file)

	if _, err := cs1.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}

	// Verify file was written
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("cache file not created: %v", err)
	}

	// Load a new CachedSource from the same file.
	mock2 := newMockSource(nil, makeEntries(0, 9))
	cs2 := NewCachedSource(mock2, file)

	got, err := cs2.GetData("BTC", "1h", tr(0, 9))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 10 {
		t.Errorf("expected 10 entries from loaded cache, got %d", len(got))
	}
	// The second source should NOT have been called because the cache was loaded.
	if mock2.calls() != 0 {
		t.Errorf("expected 0 source calls after loading from file, got %d", mock2.calls())
	}
}

func TestPersistence_VersionMismatch(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "cache.json")

	// Write a file with an incompatible version number.
	badFile := persistFormat{Version: 999, Data: map[string]tfCache{
		"BTC|1h": {
			Entries: makeEntries(0, 9),
			Covered: []interval{{hour(0), hour(9)}},
		},
	}}
	raw, _ := json.Marshal(badFile)
	if err := os.WriteFile(file, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	mock := newMockSource(nil, makeEntries(0, 9))
	cs := NewCachedSource(mock, file)

	// Cache should have been ignored due to version mismatch.
	if _, err := cs.GetData("BTC", "1h", tr(0, 9)); err != nil {
		t.Fatal(err)
	}
	if mock.calls() != 1 {
		t.Errorf("expected source to be called (version mismatch → empty cache), got %d calls", mock.calls())
	}
}

func TestPersistence_MissingFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "nonexistent", "cache.json")

	mock := newMockSource(nil, makeEntries(0, 9))
	cs := NewCachedSource(mock, file)

	// Should not panic; just fetch normally.
	got, err := cs.GetData("BTC", "1h", tr(0, 5))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected some entries")
	}

	// File should have been created (including parent dir).
	if _, err := os.Stat(file); err != nil {
		t.Errorf("cache file not created: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Parallel write + read stress test
// -----------------------------------------------------------------------------

func TestGetData_StressParallelReadsAndWrites(t *testing.T) {
	// 20 goroutines each request a slightly different window; verify no races
	// and no panics. Use the race detector by running with -race.
	mock := newMockSource(nil, makeEntries(0, 99))
	cs := newCS(mock, false, t)

	const goroutines = 20
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			start := offset * 2
			end := start + 10
			if end > 99 {
				end = 99
			}
			_, _ = cs.GetData("BTC", "1h", tr(start, end))
		}(i % 10)
	}
	wg.Wait()
}
