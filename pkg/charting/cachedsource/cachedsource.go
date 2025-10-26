package cachedsource

import (
	"testcharting/pkg/charting/sources"
	"time"
)

// TODO - Implement a caching layer for sources to avoid redundant API calls

type CachedSource struct {
	Source sources.Source
	// TODO - add caching fields
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

func (c *CachedSource) GetData(pairId string, interval string, start time.Time, end time.Time) ([]sources.ChartEntry, error) {
	// TODO - implement caching logic
	return c.Source.GetData(pairId, interval, start, end)
}

func NewCachedSource(source sources.Source) sources.Source {
	return &CachedSource{
		Source: source,
	}
}
