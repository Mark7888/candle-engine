package chartproject

import (
	"fmt"
	"slices"
	"sync"
	"testcharting/pkg/charting/sources"
	"time"
)

type ChartSourceManager struct {
	mu sync.RWMutex

	loadedSources map[string]sources.Source // key is source name
	selectedPairs []sources.SourcePairId    // e.g., "Coinbase:BTC-USD"

	fetchedData map[string][]sources.ChartEntry // key is sourcePairId, data is list of entries
}

func NewChartSourceManager() *ChartSourceManager {
	return &ChartSourceManager{
		loadedSources: make(map[string]sources.Source),
		selectedPairs: []sources.SourcePairId{},
		fetchedData:   make(map[string][]sources.ChartEntry),
	}
}

func (c *ChartSourceManager) AddSources(sources_ ...sources.Source) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loadedSources == nil {
		c.loadedSources = make(map[string]sources.Source)
	}

	for _, source := range sources_ {
		c.loadedSources[source.GetName()] = source
	}
}

func (c *ChartSourceManager) AddTradingPairs(pairs ...sources.SourcePairId) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, pair := range pairs {
		if slices.Contains(c.selectedPairs, pair) {
			return
		}

		c.selectedPairs = append(c.selectedPairs, pair)
	}
}

func (c *ChartSourceManager) GetLoadedPairs() []sources.SourcePairId {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.selectedPairs
}

func (c *ChartSourceManager) GetTimeFrameSeconds(timeFrame string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.loadedSources) > 0 {
		// Just get the time frames from the first source
		for _, source := range c.loadedSources {
			return source.GetTimeFrameSeconds(timeFrame)
		}
	}

	return 0
}

// FetchSources fetches data from multiple sources for the given pairs and time range,
// then aligns the data across all sources to ensure consistent timestamps.
// All source fetches are executed concurrently using goroutines.
func (c *ChartSourceManager) FetchSources(timeRange sources.TimeRange, pairIds []sources.SourcePairId, timeFrame string) (entries map[sources.SourcePairId][]sources.ChartEntry, width int, err error) {
	rawData := make(map[sources.SourcePairId][]sources.ChartEntry)
	var dataMu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(pairIds))

	// Fetch data from each source-pair combination concurrently
	for _, pairId := range pairIds {
		c.mu.RLock()
		source, exists := c.loadedSources[pairId.SourceName]
		c.mu.RUnlock()

		if !exists {
			return nil, 0, fmt.Errorf("source '%s' not found in loaded sources", pairId.SourceName)
		}

		wg.Add(1)
		go func(src sources.Source, pid sources.SourcePairId) {
			defer wg.Done()

			data, err := src.GetData(pid.PairId, timeFrame, timeRange)
			if err != nil {
				errChan <- err
				return
			}

			dataMu.Lock()
			rawData[pid] = data
			dataMu.Unlock()
		}(source, pairId)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	if err := <-errChan; err != nil {
		return nil, 0, err
	}

	// Align data across all sources to have consistent timestamps
	alignedData, err := c.alignChartData(rawData)
	if err != nil {
		return nil, 0, err
	}

	// Calculate width (number of aligned timestamps)
	dataWidth := 0
	for _, entries := range alignedData {
		dataWidth = len(entries)
		break // All entries have same length after alignment
	}

	return alignedData, dataWidth, nil
}

// alignChartData ensures all sources have data points at the same timestamps,
// filling gaps with invalid entries where data is missing.
func (c *ChartSourceManager) alignChartData(rawData map[sources.SourcePairId][]sources.ChartEntry) (map[sources.SourcePairId][]sources.ChartEntry, error) {
	if len(rawData) == 0 {
		return rawData, nil
	}

	// Collect all unique timestamps across all sources
	timestampSet := c.collectAllTimestamps(rawData)

	// Create aligned data with entries for each timestamp
	alignedData := make(map[sources.SourcePairId][]sources.ChartEntry)

	for pairId, entries := range rawData {
		alignedEntries := c.createAlignedEntries(entries, timestampSet)
		alignedData[pairId] = alignedEntries
	}

	return alignedData, nil
}

// collectAllTimestamps gathers all unique timestamps from all data sources
// and returns them in a sorted slice.
func (c *ChartSourceManager) collectAllTimestamps(rawData map[sources.SourcePairId][]sources.ChartEntry) []time.Time {
	timestampMap := make(map[time.Time]bool)

	for _, entries := range rawData {
		for _, entry := range entries {
			timestampMap[entry.Timestamp] = true
		}
	}

	timestamps := make([]time.Time, 0, len(timestampMap))
	for timestamp := range timestampMap {
		timestamps = append(timestamps, timestamp)
	}

	// Sort timestamps chronologically
	slices.SortFunc(timestamps, func(a, b time.Time) int {
		if a.Before(b) {
			return -1
		}
		if a.After(b) {
			return 1
		}
		return 0
	})

	return timestamps
}

// createAlignedEntries creates a complete set of entries for all timestamps,
// filling in missing timestamps with invalid entries.
func (c *ChartSourceManager) createAlignedEntries(entries []sources.ChartEntry, allTimestamps []time.Time) []sources.ChartEntry {
	// Create a map for quick lookup of existing entries
	entryMap := make(map[time.Time]sources.ChartEntry)
	for _, entry := range entries {
		entryMap[entry.Timestamp] = entry
	}

	// Build aligned entries for all timestamps
	alignedEntries := make([]sources.ChartEntry, 0, len(allTimestamps))
	for _, timestamp := range allTimestamps {
		if entry, exists := entryMap[timestamp]; exists {
			alignedEntries = append(alignedEntries, entry)
		} else {
			alignedEntries = append(alignedEntries, sources.GenerateInvalidEntry(timestamp))
		}
	}

	return alignedEntries
}
