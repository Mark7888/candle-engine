package sources

import "sync"

// MultiSourceEntries manages chart entries for multiple source pairs.
// Index 0 represents the most recent (rightmost) entry, increasing leftward.
type MultiSourceEntries struct {
	entries map[SourcePairId][]ChartEntry

	entriesMutex sync.RWMutex
}

func NewMultiSourceEntries() *MultiSourceEntries {
	return &MultiSourceEntries{
		entries: make(map[SourcePairId][]ChartEntry),
	}
}

// AddEntries adds or replaces the entries for a given source pair.
func (m *MultiSourceEntries) AddEntries(sourcePairId SourcePairId, entriesList []ChartEntry) {
	m.entriesMutex.Lock()
	defer m.entriesMutex.Unlock()

	m.entries[sourcePairId] = entriesList
}

func (m *MultiSourceEntries) ParseEntriesMap(entriesMap map[SourcePairId][]ChartEntry) {
	m.entriesMutex.Lock()
	defer m.entriesMutex.Unlock()

	for pairId, entriesList := range entriesMap {
		m.entries[pairId] = entriesList
	}
}

// GetEntry retrieves a single entry by source pair ID and index.
// Index 0 is the most recent (rightmost) entry, increasing as we go left.
// Returns nil if the pair doesn't exist or index is out of bounds.
func (m *MultiSourceEntries) GetEntry(pairId SourcePairId, index int) *ChartEntry {
	m.entriesMutex.RLock()
	defer m.entriesMutex.RUnlock()

	entries, exists := m.entries[pairId]
	if !exists {
		return nil
	}

	// Index 0 is most recent (rightmost), so we reverse the array access
	actualIndex := len(entries) - 1 - index
	if actualIndex < 0 || actualIndex >= len(entries) {
		return nil
	}

	return &entries[actualIndex]
}

// GetEntries retrieves all entries for a given source pair.
// Returns nil if the pair doesn't exist.
func (m *MultiSourceEntries) GetEntries(pairId SourcePairId) []ChartEntry {
	m.entriesMutex.RLock()
	defer m.entriesMutex.RUnlock()

	if entries, exists := m.entries[pairId]; exists {
		return entries
	}

	return nil
}

// GetEntryCount returns the number of entries for a given source pair.
// Returns 0 if the pair doesn't exist.
func (m *MultiSourceEntries) GetEntryCount(pairId SourcePairId) int {
	m.entriesMutex.RLock()
	defer m.entriesMutex.RUnlock()

	if entries, exists := m.entries[pairId]; exists {
		return len(entries)
	}

	return 0
}
