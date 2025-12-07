package cachedsource

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"testcharting/pkg/charting/sources"
	"time"
)

// cachedData stores chart entries with their time range
type cachedData struct {
	Entries []sources.ChartEntry `json:"entries"`
	Start   time.Time            `json:"start"`
	End     time.Time            `json:"end"`
}

type CachedSource struct {
	Source   sources.Source
	cache    map[string]map[string][]cachedData // pairId -> timeFrame -> []cachedData
	mu       sync.RWMutex
	filePath string // empty string disables file persistence
}

func (c *CachedSource) GetName() string {
	return c.Source.GetName()
}

func (c *CachedSource) GetTradingPairs() []sources.TradingPair {
	return c.Source.GetTradingPairs()
}

func (c *CachedSource) GetValidTimeFrames() []string {
	return c.Source.GetValidTimeFrames()
}

func (c *CachedSource) GetTimeFrameSeconds(timeFrame string) int64 {
	return c.Source.GetTimeFrameSeconds(timeFrame)
}

func (c *CachedSource) GetTimeFrameString(timeFrameSeconds int64) string {
	return c.Source.GetTimeFrameString(timeFrameSeconds)
}

func (c *CachedSource) GetData(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	c.mu.RLock()
	pairCache, pairExists := c.cache[pairId]
	var cachedRanges []cachedData
	var exists bool
	if pairExists {
		cachedRanges, exists = pairCache[timeFrame]
	}
	c.mu.RUnlock()

	if !exists {
		// No cache for this pair/timeFrame, fetch all data
		return c.fetchAndCache(pairId, timeFrame, timeRange)
	}

	// Find what ranges we need to fetch
	missingRanges := c.findMissingRanges(cachedRanges, timeRange)
	// Fetch missing ranges
	for _, missingRange := range missingRanges {
		_, err := c.fetchAndCache(pairId, timeFrame, missingRange)
		if err != nil {
			return nil, err
		}
	}

	// Now collect all relevant cached data for the requested range
	return c.getFromCache(pairId, timeFrame, timeRange), nil
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
		return sortedRanges[i].Start.Before(sortedRanges[j].Start)
	})

	var missing []sources.TimeRange
	currentStart := timeRange.Start

	for _, cached := range sortedRanges {
		// If there's a gap before this cached range
		if currentStart.Before(cached.Start) && cached.Start.Before(timeRange.End) {
			gapEnd := cached.Start
			if timeRange.End.Before(gapEnd) {
				gapEnd = timeRange.End
			}
			missing = append(missing, sources.TimeRange{Start: currentStart, End: gapEnd})
		}

		// Move currentStart past this cached range
		if cached.End.After(currentStart) {
			currentStart = cached.End
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
func (c *CachedSource) fetchAndCache(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	entries, err := c.Source.GetData(pairId, timeFrame, timeRange)
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
		Entries: entries,
		Start:   actualStart,
		End:     actualEnd,
	}

	c.mu.Lock()
	if c.cache[pairId] == nil {
		c.cache[pairId] = make(map[string][]cachedData)
	}
	if c.cache[pairId][timeFrame] == nil {
		c.cache[pairId][timeFrame] = []cachedData{}
	}
	c.cache[pairId][timeFrame] = append(c.cache[pairId][timeFrame], newCache)

	// Export to file if persistence is enabled
	if c.filePath != "" {
		if err := c.exportCacheToFile(); err != nil {
			fmt.Printf("⚠️  [CachedSource] Failed to export cache to file '%s': %v\n", c.filePath, err)
		}
	}
	c.mu.Unlock()

	return entries, nil
}

// getFromCache retrieves all cached entries within the specified time range
func (c *CachedSource) getFromCache(pairId string, timeFrame string, timeRange sources.TimeRange) []sources.ChartEntry {
	c.mu.RLock()
	var cachedRanges []cachedData
	if pairCache, exists := c.cache[pairId]; exists {
		cachedRanges = pairCache[timeFrame]
	}
	c.mu.RUnlock()

	var result []sources.ChartEntry
	seen := make(map[time.Time]bool)

	for _, cached := range cachedRanges {
		// Check if this cached range overlaps with requested range
		if cached.End.Before(timeRange.Start) || cached.Start.After(timeRange.End) {
			continue
		}

		for _, entry := range cached.Entries {
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

// exportCacheToFile writes the cache to disk
// Must be called with mu Lock held
func (c *CachedSource) exportCacheToFile() error {
	fmt.Printf("💾 [CachedSource] Exporting cache to file: %s\n", c.filePath)
	data, err := json.MarshalIndent(c.cache, "", "  ")
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to marshal cache data: %v\n", err)
		return err
	}
	err = os.WriteFile(c.filePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to write cache file '%s': %v\n", c.filePath, err)
		return err
	}
	fmt.Printf("✅ [CachedSource] Cache exported successfully to: %s\n", c.filePath)
	return nil
}

// importCacheFromFile loads the cache from disk
// Must be called with mu Lock held
func (c *CachedSource) importCacheFromFile() error {
	fmt.Printf("📂 [CachedSource] Importing cache from file: %s\n", c.filePath)
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ℹ️  [CachedSource] Cache file does not exist yet: %s\n", c.filePath)
			return nil
		}
		fmt.Printf("❌ [CachedSource] Failed to read cache file '%s': %v\n", c.filePath, err)
		return err
	}

	if len(data) == 0 {
		fmt.Printf("ℹ️  [CachedSource] Cache file is empty: %s\n", c.filePath)
		return nil
	}

	err = json.Unmarshal(data, &c.cache)
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to unmarshal cache data from '%s': %v\n", c.filePath, err)
		return err
	}
	fmt.Printf("✅ [CachedSource] Cache imported successfully from: %s\n", c.filePath)
	return nil
}

func NewCachedSource(source sources.Source, file string) sources.Source {
	cs := &CachedSource{
		Source:   source,
		cache:    make(map[string]map[string][]cachedData),
		filePath: file,
	}

	// Import existing cache if file persistence is enabled
	if file != "" {
		cs.mu.Lock()
		if err := cs.importCacheFromFile(); err != nil {
			fmt.Printf("⚠️  [CachedSource] Failed to import cache on initialization, starting with empty cache: %v\n", err)
		}
		cs.mu.Unlock()
	}

	return cs
}
