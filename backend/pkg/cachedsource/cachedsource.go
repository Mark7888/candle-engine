// Package cachedsource wraps a sources.Source with an in-memory cache that
// persists to a JSON file. It tracks which time ranges have already been
// fetched per (pairId, timeFrame), avoids redundant upstream calls, merges
// nearby uncovered windows into a single fetch, and serialises all fetch
// operations for the same key so that concurrent callers share results.
package cachedsource

import (
	"candle-engine/pkg/sources"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// -----------------------------------------------------------------------------
// Internal types
// -----------------------------------------------------------------------------

// interval is an inclusive [Start, End] time range that has been fetched from
// the upstream source.
type interval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// tfCache holds all cached data for a single (pairId, timeFrame) pair.
// Entries is always sorted ascending by Timestamp.
// Covered is always sorted and non-overlapping.
type tfCache struct {
	Entries []sources.ChartEntry `json:"entries"`
	Covered []interval           `json:"covered"`
}

// persistFormat is the on-disk JSON envelope.
type persistFormat struct {
	Version int                `json:"version"`
	Data    map[string]tfCache `json:"data"` // key: cacheKey(pairId, timeFrame)
}

const (
	persistVersion = 1
	// maxGapCandles is the number of candle-widths between two uncovered
	// segments that will still be merged into a single upstream request.
	// Fetching one slightly-larger window is cheaper than two separate calls.
	maxGapCandles = 10
)

// fetchLock serialises upstream fetches for a single cache key so that
// concurrent callers share results rather than all hitting the source.
type fetchLock struct {
	mu      sync.Mutex
	waiters int
}

// CachedSource is a caching decorator around a sources.Source.
// All cached data lives in memory; the file (if configured) is only read on
// startup and written after every mutation.
type CachedSource struct {
	source   sources.Source
	filePath string

	mu   sync.RWMutex        // protects data
	data map[string]*tfCache // key: cacheKey(pairId, timeFrame)

	fetchMu    sync.Mutex            // protects fetchLocks map
	fetchLocks map[string]*fetchLock // one lock per cache key
}

// -----------------------------------------------------------------------------
// Constructor
// -----------------------------------------------------------------------------

// NewCachedSource wraps source with a caching layer.
// If filePath is non-empty the cache is loaded from that file at startup and
// written back after every update. Pass an empty string for memory-only mode.
func NewCachedSource(source sources.Source, filePath string) sources.Source {
	cs := &CachedSource{
		source:     source,
		filePath:   filePath,
		data:       make(map[string]*tfCache),
		fetchLocks: make(map[string]*fetchLock),
	}

	if filePath != "" {
		if err := cs.loadFromFile(); err != nil {
			fmt.Printf("⚠️  [CachedSource] Could not load cache from '%s' (starting empty): %v\n", filePath, err)
		} else {
			fmt.Printf("✅ [CachedSource] Cache loaded from: %s\n", filePath)
		}
	}

	return cs
}

// -----------------------------------------------------------------------------
// sources.Source delegation
// -----------------------------------------------------------------------------

func (c *CachedSource) GetName() string                        { return c.source.GetName() }
func (c *CachedSource) GetTradingPairs() []sources.TradingPair { return c.source.GetTradingPairs() }
func (c *CachedSource) GetValidTimeFrames() []string           { return c.source.GetValidTimeFrames() }
func (c *CachedSource) GetTimeFrameSeconds(tf string) int64    { return c.source.GetTimeFrameSeconds(tf) }
func (c *CachedSource) GetTimeFrameString(secs int64) string {
	return c.source.GetTimeFrameString(secs)
}

// -----------------------------------------------------------------------------
// GetData – the main entry point
// -----------------------------------------------------------------------------

// GetData returns chart entries for the requested window, fetching only the
// sub-ranges that are not already in cache.
func (c *CachedSource) GetData(pairId, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	key := cacheKey(pairId, timeFrame)

	// Fast path: everything already cached (read lock only).
	c.mu.RLock()
	if tf, ok := c.data[key]; ok && covers(tf.Covered, timeRange) {
		result := extractEntries(tf.Entries, timeRange)
		c.mu.RUnlock()
		return result, nil
	}
	c.mu.RUnlock()

	// Slow path: acquire per-key fetch lock so that concurrent callers for the
	// same (pairId, timeFrame) queue up rather than all hitting the source.
	c.acquireFetchLock(key)
	defer c.releaseFetchLock(key)

	// Re-check after acquiring the lock – a previous waiter may have populated
	// the cache while we were queued.
	c.mu.RLock()
	var covered []interval
	if tf, ok := c.data[key]; ok {
		covered = tf.Covered
	}
	c.mu.RUnlock()

	missing := missingRanges(covered, timeRange)
	if len(missing) == 0 {
		c.mu.RLock()
		result := extractEntries(c.data[key].Entries, timeRange)
		c.mu.RUnlock()
		return result, nil
	}

	// Merge nearby missing segments to minimise upstream calls.
	tfSeconds := c.source.GetTimeFrameSeconds(timeFrame)
	toFetch := mergeNearbyRanges(missing, tfSeconds)

	for _, r := range toFetch {
		entries, err := c.source.GetData(pairId, timeFrame, r)
		if err != nil {
			return nil, fmt.Errorf("cachedsource: fetch %s/%s [%s – %s]: %w",
				pairId, timeFrame, r.Start.Format(time.RFC3339), r.End.Format(time.RFC3339), err)
		}
		// Mark the requested range as covered even if the source returned nothing
		// (e.g. a weekend gap), so we do not re-fetch it next time.
		c.updateCache(key, entries, r)
	}

	c.mu.RLock()
	var result []sources.ChartEntry
	if tf, ok := c.data[key]; ok {
		result = extractEntries(tf.Entries, timeRange)
	}
	c.mu.RUnlock()
	return result, nil
}

// -----------------------------------------------------------------------------
// Cache update
// -----------------------------------------------------------------------------

// updateCache merges newEntries into the in-memory cache for key, marks
// fetchedRange as covered, and persists to disk if a file path is configured.
// Must NOT be called with mu held.
func (c *CachedSource) updateCache(key string, newEntries []sources.ChartEntry, fetchedRange sources.TimeRange) {
	c.mu.Lock()
	tf, ok := c.data[key]
	if !ok {
		tf = &tfCache{}
		c.data[key] = tf
	}
	tf.Entries = mergeEntries(tf.Entries, newEntries)
	tf.Covered = addCoveredInterval(tf.Covered, interval{Start: fetchedRange.Start, End: fetchedRange.End})
	c.mu.Unlock()

	if c.filePath != "" {
		if err := c.saveToFile(); err != nil {
			fmt.Printf("⚠️  [CachedSource] Failed to save cache to '%s': %v\n", c.filePath, err)
		}
	}
}

// -----------------------------------------------------------------------------
// Per-key fetch lock helpers
// -----------------------------------------------------------------------------

func (c *CachedSource) acquireFetchLock(key string) {
	c.fetchMu.Lock()
	fl, ok := c.fetchLocks[key]
	if !ok {
		fl = &fetchLock{}
		c.fetchLocks[key] = fl
	}
	fl.waiters++
	c.fetchMu.Unlock()

	fl.mu.Lock()
}

func (c *CachedSource) releaseFetchLock(key string) {
	c.fetchMu.Lock()
	fl := c.fetchLocks[key]
	c.fetchMu.Unlock()

	fl.mu.Unlock()

	c.fetchMu.Lock()
	fl.waiters--
	if fl.waiters == 0 {
		delete(c.fetchLocks, key)
	}
	c.fetchMu.Unlock()
}

// -----------------------------------------------------------------------------
// File persistence
// -----------------------------------------------------------------------------

func (c *CachedSource) loadFromFile() error {
	raw, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no file yet – start empty
		}
		return fmt.Errorf("read: %w", err)
	}
	if len(raw) == 0 {
		return nil
	}

	var pf persistFormat
	if err := json.Unmarshal(raw, &pf); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	if pf.Version != persistVersion {
		fmt.Printf("⚠️  [CachedSource] Cache version mismatch (file=%d, want=%d) – ignoring cached data\n",
			pf.Version, persistVersion)
		return nil
	}

	c.mu.Lock()
	for k, tf := range pf.Data {
		copied := tf
		c.data[k] = &copied
	}
	c.mu.Unlock()
	return nil
}

// saveToFile writes the in-memory cache to disk atomically (temp-file + rename).
// Must NOT be called with mu held.
func (c *CachedSource) saveToFile() error {
	c.mu.RLock()
	pf := persistFormat{
		Version: persistVersion,
		Data:    make(map[string]tfCache, len(c.data)),
	}
	for k, tf := range c.data {
		pf.Data[k] = *tf
	}
	c.mu.RUnlock()

	raw, err := json.Marshal(pf)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tmp := c.filePath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, c.filePath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Pure helper functions (no shared state – easy to unit-test)
// -----------------------------------------------------------------------------

// cacheKey returns the map key for a (pairId, timeFrame) pair.
func cacheKey(pairId, timeFrame string) string {
	return pairId + "|" + timeFrame
}

// covers reports whether the sorted, non-overlapping covered list fully spans want.
func covers(covered []interval, want sources.TimeRange) bool {
	current := want.Start
	for _, iv := range covered {
		if !current.Before(want.End) {
			return true
		}
		if iv.Start.After(current) {
			return false // uncovered gap
		}
		if iv.End.After(current) {
			current = iv.End
		}
	}
	return !current.Before(want.End)
}

// missingRanges returns the sub-ranges of want that are not covered.
// covered must be sorted and non-overlapping.
func missingRanges(covered []interval, want sources.TimeRange) []sources.TimeRange {
	var missing []sources.TimeRange
	current := want.Start

	for _, iv := range covered {
		if !current.Before(want.End) {
			break
		}
		if iv.End.Before(current) || iv.End.Equal(current) {
			continue // interval ends at or before our cursor
		}
		if iv.Start.After(current) {
			// Gap before this interval
			end := iv.Start
			if end.After(want.End) {
				end = want.End
			}
			missing = append(missing, sources.TimeRange{Start: current, End: end})
		}
		if iv.End.After(current) {
			current = iv.End
		}
	}

	if current.Before(want.End) {
		missing = append(missing, sources.TimeRange{Start: current, End: want.End})
	}
	return missing
}

// mergeNearbyRanges collapses adjacent missing ranges whose separating gap is
// smaller than maxGapCandles candle-widths into a single wider fetch window.
// This trades fetching a few extra (already-cached) candles for fewer round trips.
func mergeNearbyRanges(missing []sources.TimeRange, tfSeconds int64) []sources.TimeRange {
	if len(missing) <= 1 || tfSeconds <= 0 {
		return missing
	}
	maxGap := time.Duration(maxGapCandles*tfSeconds) * time.Second
	result := []sources.TimeRange{missing[0]}

	for i := 1; i < len(missing); i++ {
		last := &result[len(result)-1]
		gap := missing[i].Start.Sub(last.End)
		if gap <= maxGap {
			last.End = missing[i].End // merge
		} else {
			result = append(result, missing[i])
		}
	}
	return result
}

// addCoveredInterval inserts newIv into the sorted, non-overlapping covered
// slice, merging any intervals that now overlap or touch.
func addCoveredInterval(covered []interval, newIv interval) []interval {
	result := make([]interval, 0, len(covered)+1)
	merged := newIv

	for _, iv := range covered {
		switch {
		case iv.End.Before(merged.Start):
			// iv is entirely before merged – keep it as-is
			result = append(result, iv)
		case iv.Start.After(merged.End):
			// iv is entirely after merged – flush merged then keep iv
			result = append(result, merged)
			merged = iv
		default:
			// Overlap or touching – expand merged to cover both
			if iv.Start.Before(merged.Start) {
				merged.Start = iv.Start
			}
			if iv.End.After(merged.End) {
				merged.End = iv.End
			}
		}
	}
	result = append(result, merged)
	return result
}

// mergeEntries returns a sorted, deduplicated union of two entry slices.
// Both inputs must be sorted ascending by Timestamp.
// When the same timestamp appears in both slices, newEntries takes precedence.
func mergeEntries(existing, newEntries []sources.ChartEntry) []sources.ChartEntry {
	if len(newEntries) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return newEntries
	}

	combined := make(map[time.Time]sources.ChartEntry, len(existing)+len(newEntries))
	for _, e := range existing {
		combined[e.Timestamp] = e
	}
	for _, e := range newEntries {
		combined[e.Timestamp] = e
	}

	result := make([]sources.ChartEntry, 0, len(combined))
	for _, e := range combined {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	return result
}

// extractEntries returns a copy of the entries in [want.Start, want.End] using
// binary search on the sorted slice.
func extractEntries(entries []sources.ChartEntry, want sources.TimeRange) []sources.ChartEntry {
	lo := sort.Search(len(entries), func(i int) bool {
		return !entries[i].Timestamp.Before(want.Start)
	})
	hi := sort.Search(len(entries), func(i int) bool {
		return entries[i].Timestamp.After(want.End)
	})
	if lo >= hi {
		return nil
	}
	result := make([]sources.ChartEntry, hi-lo)
	copy(result, entries[lo:hi])
	return result
}
