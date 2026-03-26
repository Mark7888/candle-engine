package cachedsource

import (
	"candle-engine/pkg/charting/sources"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// TODO - there is an anomaly, where if i cache data for 2 pairs at the same time, only one pair gets cached properly. Need to investigate.
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
	logger   *log.Logger
	logFile  *os.File
}

// logf writes a structured log entry to the log file when the logger is set.
// Format: TIMESTAMP [EVENT] key=value ...
func (c *CachedSource) logf(event, format string, args ...any) {
	if c.logger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	c.logger.Printf("[%s] %s", event, msg)
}

// Close releases the log file handle. It should be called when the CachedSource
// is no longer needed to avoid leaking file descriptors.
func (c *CachedSource) Close() {
	if c.logFile != nil {
		c.logf("CachedSource:CLOSE", "filePath=%q", c.filePath)
		c.logFile.Close()
		c.logFile = nil
		c.logger = nil
	}
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
	c.logf("GetData:START", "pairId=%q timeFrame=%q rangeStart=%s rangeEnd=%s",
		pairId, timeFrame, timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))

	c.mu.RLock()
	pairCache, pairExists := c.cache[pairId]
	var cachedRanges []cachedData
	var exists bool
	if pairExists {
		cachedRanges, exists = pairCache[timeFrame]
	}
	c.mu.RUnlock()

	if !exists {
		c.logf("GetData:CACHE_MISS", "pairId=%q timeFrame=%q no_cache_entry=true fetching_full_range=true", pairId, timeFrame)
		// No cache for this pair/timeFrame, fetch all data
		return c.fetchAndCache(pairId, timeFrame, timeRange)
	}

	c.logf("GetData:CACHE_EXISTS", "pairId=%q timeFrame=%q cached_segment_count=%d", pairId, timeFrame, len(cachedRanges))

	// Find what ranges we need to fetch
	missingRanges := c.findMissingRanges(cachedRanges, timeRange)

	c.logf("GetData:MISSING_RANGES", "pairId=%q timeFrame=%q missing_count=%d", pairId, timeFrame, len(missingRanges))
	for i, mr := range missingRanges {
		c.logf("GetData:MISSING_RANGE", "pairId=%q timeFrame=%q index=%d start=%s end=%s",
			pairId, timeFrame, i, mr.Start.UTC().Format(time.RFC3339), mr.End.UTC().Format(time.RFC3339))
	}

	// Fetch missing ranges
	for _, missingRange := range missingRanges {
		_, err := c.fetchAndCache(pairId, timeFrame, missingRange)
		if err != nil {
			c.logf("GetData:FETCH_ERROR", "pairId=%q timeFrame=%q start=%s end=%s error=%q",
				pairId, timeFrame, missingRange.Start.UTC().Format(time.RFC3339), missingRange.End.UTC().Format(time.RFC3339), err.Error())
			return nil, err
		}
	}

	// Now collect all relevant cached data for the requested range
	result := c.getFromCache(pairId, timeFrame, timeRange)
	c.logf("GetData:DONE", "pairId=%q timeFrame=%q returned_entries=%d", pairId, timeFrame, len(result))
	return result, nil
}

// findMissingRanges determines which time ranges are not covered by the cache
func (c *CachedSource) findMissingRanges(cachedRanges []cachedData, timeRange sources.TimeRange) []sources.TimeRange {
	c.logf("findMissingRanges:START", "cached_segment_count=%d requestedStart=%s requestedEnd=%s",
		len(cachedRanges), timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))

	if len(cachedRanges) == 0 {
		c.logf("findMissingRanges:RESULT", "no_cached_segments=true returning_full_range=true start=%s end=%s",
			timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))
		return []sources.TimeRange{{Start: timeRange.Start, End: timeRange.End}}
	}

	// Sort cached ranges by start time
	sortedRanges := make([]cachedData, len(cachedRanges))
	copy(sortedRanges, cachedRanges)
	sort.Slice(sortedRanges, func(i, j int) bool {
		return sortedRanges[i].Start.Before(sortedRanges[j].Start)
	})

	for i, sr := range sortedRanges {
		c.logf("findMissingRanges:CACHED_SEGMENT", "index=%d start=%s end=%s entry_count=%d",
			i, sr.Start.UTC().Format(time.RFC3339), sr.End.UTC().Format(time.RFC3339), len(sr.Entries))
	}

	var missing []sources.TimeRange
	currentStart := timeRange.Start

	for _, cached := range sortedRanges {
		// If there's a gap before this cached range
		if currentStart.Before(cached.Start) && cached.Start.Before(timeRange.End) {
			gapEnd := cached.Start
			if timeRange.End.Before(gapEnd) {
				gapEnd = timeRange.End
			}
			c.logf("findMissingRanges:GAP_FOUND", "gapStart=%s gapEnd=%s cachedSegmentStart=%s",
				currentStart.UTC().Format(time.RFC3339), gapEnd.UTC().Format(time.RFC3339), cached.Start.UTC().Format(time.RFC3339))
			missing = append(missing, sources.TimeRange{Start: currentStart, End: gapEnd})
		}

		// Move currentStart past this cached range
		if cached.End.After(currentStart) {
			c.logf("findMissingRanges:ADVANCE_CURSOR", "from=%s to=%s cachedEnd=%s",
				currentStart.UTC().Format(time.RFC3339), cached.End.UTC().Format(time.RFC3339), cached.End.UTC().Format(time.RFC3339))
			currentStart = cached.End
		}

		// If we've covered up to or past the requested end, we're done
		if !currentStart.Before(timeRange.End) {
			c.logf("findMissingRanges:FULLY_COVERED", "currentStart=%s requestedEnd=%s",
				currentStart.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))
			break
		}
	}

	// Check if there's a gap after all cached ranges
	if currentStart.Before(timeRange.End) {
		c.logf("findMissingRanges:TRAILING_GAP", "start=%s end=%s",
			currentStart.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))
		missing = append(missing, sources.TimeRange{Start: currentStart, End: timeRange.End})
	}

	c.logf("findMissingRanges:RESULT", "missing_range_count=%d", len(missing))
	for i, m := range missing {
		c.logf("findMissingRanges:MISSING", "index=%d start=%s end=%s", i, m.Start.UTC().Format(time.RFC3339), m.End.UTC().Format(time.RFC3339))
	}

	return missing
}

// fetchAndCache fetches data from the source and adds it to the cache
func (c *CachedSource) fetchAndCache(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	c.logf("fetchAndCache:START", "pairId=%q timeFrame=%q start=%s end=%s",
		pairId, timeFrame, timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))

	entries, err := c.Source.GetData(pairId, timeFrame, timeRange)
	if err != nil {
		c.logf("fetchAndCache:API_ERROR", "pairId=%q timeFrame=%q error=%q", pairId, timeFrame, err.Error())
		return nil, err
	}

	c.logf("fetchAndCache:API_RESULT", "pairId=%q timeFrame=%q entry_count=%d", pairId, timeFrame, len(entries))

	if len(entries) == 0 {
		c.logf("fetchAndCache:EMPTY_RESULT", "pairId=%q timeFrame=%q no_entries=true skipping_cache_save=true", pairId, timeFrame)
		return entries, nil
	}

	// Determine actual time range from the data
	actualStart := entries[0].Timestamp
	actualEnd := entries[len(entries)-1].Timestamp

	c.logf("fetchAndCache:ACTUAL_RANGE", "pairId=%q timeFrame=%q actualStart=%s actualEnd=%s requested_start=%s requested_end=%s",
		pairId, timeFrame,
		actualStart.UTC().Format(time.RFC3339), actualEnd.UTC().Format(time.RFC3339),
		timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))

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

	segmentCount := len(c.cache[pairId][timeFrame])
	c.logf("fetchAndCache:SAVED_TO_CACHE", "pairId=%q timeFrame=%q segment_index=%d total_segments=%d start=%s end=%s entry_count=%d",
		pairId, timeFrame, segmentCount-1, segmentCount,
		actualStart.UTC().Format(time.RFC3339), actualEnd.UTC().Format(time.RFC3339), len(entries))

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
	c.logf("getFromCache:START", "pairId=%q timeFrame=%q start=%s end=%s",
		pairId, timeFrame, timeRange.Start.UTC().Format(time.RFC3339), timeRange.End.UTC().Format(time.RFC3339))

	c.mu.RLock()
	var cachedRanges []cachedData
	if pairCache, exists := c.cache[pairId]; exists {
		cachedRanges = pairCache[timeFrame]
	}
	c.mu.RUnlock()

	c.logf("getFromCache:SEGMENTS", "pairId=%q timeFrame=%q segment_count=%d", pairId, timeFrame, len(cachedRanges))

	var result []sources.ChartEntry
	seen := make(map[time.Time]bool)

	for i, cached := range cachedRanges {
		// Check if this cached range overlaps with requested range
		if cached.End.Before(timeRange.Start) || cached.Start.After(timeRange.End) {
			c.logf("getFromCache:SEGMENT_SKIP", "segment_index=%d start=%s end=%s reason=no_overlap",
				i, cached.Start.UTC().Format(time.RFC3339), cached.End.UTC().Format(time.RFC3339))
			continue
		}

		entriesBefore := len(result)
		for _, entry := range cached.Entries {
			// Check if entry is within requested range and not already added
			if !entry.Timestamp.Before(timeRange.Start) && !entry.Timestamp.After(timeRange.End) && !seen[entry.Timestamp] {
				result = append(result, entry)
				seen[entry.Timestamp] = true
			}
		}
		c.logf("getFromCache:SEGMENT_USE", "segment_index=%d start=%s end=%s entries_added=%d",
			i, cached.Start.UTC().Format(time.RFC3339), cached.End.UTC().Format(time.RFC3339), len(result)-entriesBefore)
	}

	// Sort results by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})

	c.logf("getFromCache:RESULT", "pairId=%q timeFrame=%q total_entries_returned=%d", pairId, timeFrame, len(result))
	return result
}

// exportCacheToFile writes the cache to disk
// Must be called with mu Lock held
func (c *CachedSource) exportCacheToFile() error {
	fmt.Printf("💾 [CachedSource] Exporting cache to file: %s\n", c.filePath)
	c.logf("exportCacheToFile:START", "filePath=%q", c.filePath)

	data, err := json.MarshalIndent(c.cache, "", "  ")
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to marshal cache data: %v\n", err)
		c.logf("exportCacheToFile:ERROR", "stage=marshal error=%q", err.Error())
		return err
	}

	// Create directory structure if it doesn't exist
	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ [CachedSource] Failed to create directory '%s': %v\n", dir, err)
		c.logf("exportCacheToFile:ERROR", "stage=mkdir dir=%q error=%q", dir, err.Error())
		return err
	}

	err = os.WriteFile(c.filePath, data, 0644)
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to write cache file '%s': %v\n", c.filePath, err)
		c.logf("exportCacheToFile:ERROR", "stage=write error=%q", err.Error())
		return err
	}
	fmt.Printf("✅ [CachedSource] Cache exported successfully to: %s\n", c.filePath)
	c.logf("exportCacheToFile:SUCCESS", "filePath=%q bytes_written=%d", c.filePath, len(data))
	return nil
}

// importCacheFromFile loads the cache from disk
// Must be called with mu Lock held
func (c *CachedSource) importCacheFromFile() error {
	fmt.Printf("📂 [CachedSource] Importing cache from file: %s\n", c.filePath)
	c.logf("importCacheFromFile:START", "filePath=%q", c.filePath)

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ℹ️  [CachedSource] Cache file does not exist yet: %s\n", c.filePath)
			c.logf("importCacheFromFile:NOT_FOUND", "filePath=%q starting_empty=true", c.filePath)
			return nil
		}
		fmt.Printf("❌ [CachedSource] Failed to read cache file '%s': %v\n", c.filePath, err)
		c.logf("importCacheFromFile:ERROR", "stage=read filePath=%q error=%q", c.filePath, err.Error())
		return err
	}

	if len(data) == 0 {
		fmt.Printf("ℹ️  [CachedSource] Cache file is empty: %s\n", c.filePath)
		c.logf("importCacheFromFile:EMPTY", "filePath=%q starting_empty=true", c.filePath)
		return nil
	}

	err = json.Unmarshal(data, &c.cache)
	if err != nil {
		fmt.Printf("❌ [CachedSource] Failed to unmarshal cache data from '%s': %v\n", c.filePath, err)
		c.logf("importCacheFromFile:ERROR", "stage=unmarshal filePath=%q error=%q", c.filePath, err.Error())
		return err
	}

	// Count loaded segments for logging
	totalSegments := 0
	for pairId, tfMap := range c.cache {
		for tf, segs := range tfMap {
			c.logf("importCacheFromFile:LOADED_PAIR", "pairId=%q timeFrame=%q segment_count=%d", pairId, tf, len(segs))
			totalSegments += len(segs)
		}
	}

	fmt.Printf("✅ [CachedSource] Cache imported successfully from: %s\n", c.filePath)
	c.logf("importCacheFromFile:SUCCESS", "filePath=%q bytes_read=%d total_segments=%d", c.filePath, len(data), totalSegments)
	return nil
}

func NewCachedSource(source sources.Source, file string) sources.Source {
	cs := &CachedSource{
		Source:   source,
		cache:    make(map[string]map[string][]cachedData),
		filePath: file,
	}

	// Set up file logger when a file path is configured
	if file != "" {
		logPath := file + ".log"
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("⚠️  [CachedSource] Failed to create log directory '%s': %v\n", logDir, err)
		} else {
			lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("⚠️  [CachedSource] Failed to open log file '%s': %v\n", logPath, err)
			} else {
				cs.logFile = lf
				cs.logger = log.New(lf, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC)
				cs.logf("NewCachedSource:INIT", "filePath=%q logPath=%q sourceName=%q", file, logPath, source.GetName())
			}
		}
	}

	// Import existing cache if file persistence is enabled
	if file != "" {
		cs.mu.Lock()
		if err := cs.importCacheFromFile(); err != nil {
			fmt.Printf("⚠️  [CachedSource] Failed to import cache on initialization, starting with empty cache: %v\n", err)
			cs.logf("NewCachedSource:IMPORT_ERROR", "error=%q", err.Error())
		}
		cs.mu.Unlock()
	}

	return cs
}
