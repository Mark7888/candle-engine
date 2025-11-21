package sources

import "sync"

type MultiSourceEntries struct {
	entries map[string][]ChartEntry // key is pair ID, like "Coinbase:BTC-USD"

	entriesMutex sync.RWMutex
}

func NewMultiSourceEntries() *MultiSourceEntries {
	return &MultiSourceEntries{
		entries: make(map[string][]ChartEntry),
	}
}

func (m *MultiSourceEntries) AddEntries(sourcePairId string, entries []ChartEntry) {
	m.entriesMutex.Lock()
	defer m.entriesMutex.Unlock()

	m.entries[sourcePairId] = entries
}

func (m *MultiSourceEntries) GetEntries(pairId string) []ChartEntry {
	m.entriesMutex.RLock()
	defer m.entriesMutex.RUnlock()

	if entries, exists := m.entries[pairId]; exists {
		return entries
	}

	return nil
}
