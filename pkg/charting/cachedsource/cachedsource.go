package cachedsource

import (
	"sort"
	"sync"
	"testcharting/pkg/charting/sources"
	"time"
)

// cacheKey uniquely identifies cached data by pair and interval
type cacheKey struct {
	pairId   string
	interval string
}

// cachedData stores chart entries with their time range
type cachedData struct {
	entries []sources.ChartEntry
	start   time.Time
	end     time.Time
}

type CachedSource struct {
	Source sources.Source
	cache  map[cacheKey][]cachedData
	mu     sync.RWMutex
	// TODO - save cache to file?
}

func (c *CachedSource) GetName() string {
	return c.Source.GetName()
}

func (c *CachedSource) GetTradingPairs() []sources.TradingPair {
	return c.Source.GetTradingPairs()
}

func (c *CachedSource) GetValidIntervals() []string {
	return c.Source.GetValidIntervals()
}

func (c *CachedSource) GetIntervalSeconds(interval string) []int64 {
	return c.Source.GetIntervalSeconds(interval)
}

func (c *CachedSource) GetData(pairId string, interval string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	key := cacheKey{pairId: pairId, interval: interval}

	c.mu.RLock()
	cachedRanges, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		// No cache for this pair/interval, fetch all data
		return c.fetchAndCache(key, timeRange)
	}

	// Find what ranges we need to fetch
	missingRanges := c.findMissingRanges(cachedRanges, timeRange)
	// Fetch missing ranges
	for _, missingRange := range missingRanges {
		_, err := c.fetchAndCache(key, missingRange)
		if err != nil {
			return nil, err
		}
	}

	// Now collect all relevant cached data for the requested range
	return c.getFromCache(key, timeRange), nil
}

// findMissingRanges determines which time ranges are not covered by the cache
func (c *CachedSource) findMissingRanges(cachedRanges []cachedData, timeRange sources.TimeRange) []sources.TimeRange {
	if len(cachedRanges) == 0 {
		return []sources.TimeRange{{Start: timeRange.Start, End: timeRange.End}}
	}

	// Sort cached ranges by start time
	sortedRanges := make([]cachedData, len(cachedRanges))
	copy(sortedRanges, cachedRanges)
	sort.Slice(sortedRanges, func(i, j int) bool {
		return sortedRanges[i].start.Before(sortedRanges[j].start)
	})

	var missing []sources.TimeRange
	currentStart := timeRange.Start

	for _, cached := range sortedRanges {
		// If there's a gap before this cached range
		if currentStart.Before(cached.start) && cached.start.Before(timeRange.End) {
			gapEnd := cached.start
			if timeRange.End.Before(gapEnd) {
				gapEnd = timeRange.End
			}
			missing = append(missing, sources.TimeRange{Start: currentStart, End: gapEnd})
		}

		// Move currentStart past this cached range
		if cached.end.After(currentStart) {
			currentStart = cached.end
		}

		// If we've covered up to or past the requested end, we're done
		if !currentStart.Before(timeRange.End) {
			break
		}
	}

	// Check if there's a gap after all cached ranges
	if currentStart.Before(timeRange.End) {
		missing = append(missing, sources.TimeRange{Start: currentStart, End: timeRange.End})
	}

	return missing
}

// fetchAndCache fetches data from the source and adds it to the cache
func (c *CachedSource) fetchAndCache(key cacheKey, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	entries, err := c.Source.GetData(key.pairId, key.interval, timeRange)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return entries, nil
	}

	// Determine actual time range from the data
	actualStart := entries[0].Timestamp
	actualEnd := entries[len(entries)-1].Timestamp

	newCache := cachedData{
		entries: entries,
		start:   actualStart,
		end:     actualEnd,
	}

	c.mu.Lock()
	if c.cache[key] == nil {
		c.cache[key] = []cachedData{}
	}
	c.cache[key] = append(c.cache[key], newCache)
	c.mu.Unlock()

	return entries, nil
}

// getFromCache retrieves all cached entries within the specified time range
func (c *CachedSource) getFromCache(key cacheKey, timeRange sources.TimeRange) []sources.ChartEntry {
	c.mu.RLock()
	cachedRanges := c.cache[key]
	c.mu.RUnlock()

	var result []sources.ChartEntry
	seen := make(map[time.Time]bool)

	for _, cached := range cachedRanges {
		// Check if this cached range overlaps with requested range
		if cached.end.Before(timeRange.Start) || cached.start.After(timeRange.End) {
			continue
		}

		for _, entry := range cached.entries {
			// Check if entry is within requested range and not already added
			if !entry.Timestamp.Before(timeRange.Start) && !entry.Timestamp.After(timeRange.End) && !seen[entry.Timestamp] {
				result = append(result, entry)
				seen[entry.Timestamp] = true
			}
		}
	}

	// Sort results by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	return result
}

func NewCachedSource(source sources.Source) sources.Source {
	return &CachedSource{
		Source: source,
		cache:  make(map[cacheKey][]cachedData),
	}
}
