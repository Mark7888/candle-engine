package chartproject

import (
	"testcharting/pkg/charting/sources"
)

// TODO - What should this store? Sources, selected pairs, fetched data? Should be separate these concerns?
type ChartSourceManager struct {
	loadedSources map[string]sources.Source // key is source name
	selectedPairs []string                  // e.g., "Coinbase:BTC-USD"

	fetchedData []sources.ChartEntry
}

func NewChartSourceManager() *ChartSourceManager {
	return &ChartSourceManager{
		loadedSources: make(map[string]sources.Source),
		selectedPairs: []string{},
		fetchedData:   []sources.ChartEntry{},
	}
}

func (c *ChartSourceManager) AddSources(sources_ ...sources.Source) {
	if c.loadedSources == nil {
		c.loadedSources = make(map[string]sources.Source)
	}

	for _, source := range sources_ {
		c.loadedSources[source.GetName()] = source
	}
}

// TODO - implement fetching logic
// Maybe these could run in parallel for each source?
func (c *ChartSourceManager) FetchSources(timeRange sources.TimeRange, pairIds []string, interval string) (map[string][]sources.ChartEntry, error) {
	// TODO Implement
	return nil, nil
}

func (c *ChartSourceManager) AddTradingPairs(pairs ...string) {
	for _, pair := range pairs {
		c.addTradingPair(pair)
	}
}

func (c *ChartSourceManager) addTradingPair(pair string) {
	for _, p := range c.selectedPairs {
		if p == pair {
			return
		}
	}

	c.selectedPairs = append(c.selectedPairs, pair)
}
