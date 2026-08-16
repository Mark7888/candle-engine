package registry

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"
)

type Registry struct {
	indicators  map[string]indicators.BaseIndicator
	dataSources map[string]sources.Source
}

func NewRegistry() *Registry {
	return &Registry{
		indicators:  make(map[string]indicators.BaseIndicator),
		dataSources: make(map[string]sources.Source),
	}
}

func (r *Registry) RegisterIndicators(indicators ...indicators.BaseIndicator) {
	for _, indicator := range indicators {
		name := indicator.GetName()
		if _, exists := r.indicators[name]; !exists {
			r.indicators[name] = indicator
		}
	}
}

func (r *Registry) UnregisterIndicator(name string) {
	delete(r.indicators, name)
}

func (r *Registry) ClearIndicators() {
	r.indicators = make(map[string]indicators.BaseIndicator)
}

func (r *Registry) GetIndicator(name string) (indicators.BaseIndicator, bool) {
	indicator, exists := r.indicators[name]
	return indicator, exists
}

func (r *Registry) ListIndicators() []string {
	return slices.Collect(maps.Keys(r.indicators))
}

func (r *Registry) RegisterDataSources(sources ...sources.Source) {
	for _, source := range sources {
		name := source.GetName()
		if _, exists := r.dataSources[name]; !exists {
			r.dataSources[name] = source
		}
	}
}

func (r *Registry) UnregisterDataSource(name string) {
	delete(r.dataSources, name)
}

func (r *Registry) ClearDataSources() {
	r.dataSources = make(map[string]sources.Source)
}

func (r *Registry) GetDataSource(name string) (sources.Source, bool) {
	source, exists := r.dataSources[name]
	return source, exists
}

func (r *Registry) ListDataSources() []string {
	return slices.Collect(maps.Keys(r.dataSources))
}

func (r *Registry) FetchSources(timeRange sources.TimeRange, pairIds []sources.SourcePairId, timeFrame string) (map[sources.SourcePairId][]sources.ChartEntry, int, error) {
	rawData := make(map[sources.SourcePairId][]sources.ChartEntry)
	var dataMu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(pairIds))

	for _, pairId := range pairIds {
		src, exists := r.dataSources[pairId.SourceName]
		if !exists {
			return nil, 0, fmt.Errorf("source '%s' not found in registry", pairId.SourceName)
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
		}(src, pairId)
	}

	wg.Wait()
	close(errChan)
	if err := <-errChan; err != nil {
		return nil, 0, err
	}

	alignedData, err := r.alignChartData(rawData)
	if err != nil {
		return nil, 0, err
	}

	dataWidth := 0
	for _, entries := range alignedData {
		dataWidth = len(entries)
		break
	}

	return alignedData, dataWidth, nil
}

func (r *Registry) alignChartData(rawData map[sources.SourcePairId][]sources.ChartEntry) (map[sources.SourcePairId][]sources.ChartEntry, error) {
	if len(rawData) == 0 {
		return rawData, nil
	}
	timestamps := r.collectAllTimestamps(rawData)
	aligned := make(map[sources.SourcePairId][]sources.ChartEntry)
	for pairId, entries := range rawData {
		aligned[pairId] = sources.FillInvalidEntries(r.createAlignedEntries(entries, timestamps))
	}
	return aligned, nil
}

func (r *Registry) collectAllTimestamps(rawData map[sources.SourcePairId][]sources.ChartEntry) []time.Time {
	tsMap := make(map[time.Time]bool)
	for _, entries := range rawData {
		for _, e := range entries {
			tsMap[e.Timestamp] = true
		}
	}
	timestamps := make([]time.Time, 0, len(tsMap))
	for ts := range tsMap {
		timestamps = append(timestamps, ts)
	}
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

func (r *Registry) createAlignedEntries(entries []sources.ChartEntry, allTimestamps []time.Time) []sources.ChartEntry {
	entryMap := make(map[time.Time]sources.ChartEntry, len(entries))
	for _, e := range entries {
		entryMap[e.Timestamp] = e
	}
	aligned := make([]sources.ChartEntry, 0, len(allTimestamps))
	for _, ts := range allTimestamps {
		if e, ok := entryMap[ts]; ok {
			aligned = append(aligned, e)
		} else {
			aligned = append(aligned, sources.GenerateInvalidEntry(ts))
		}
	}
	return aligned
}

func (r *Registry) GetCommonTimeFrames(sourceNames []string) []string {
	if len(sourceNames) == 0 {
		return []string{}
	}

	first, ok := r.dataSources[sourceNames[0]]
	if !ok {
		return []string{}
	}

	commonSeconds := make(map[int64]bool)
	for _, tf := range first.GetValidTimeFrames() {
		commonSeconds[first.GetTimeFrameSeconds(tf)] = true
	}

	for _, name := range sourceNames[1:] {
		src, ok := r.dataSources[name]
		if !ok {
			continue
		}
		srcSeconds := make(map[int64]bool)
		for _, tf := range src.GetValidTimeFrames() {
			srcSeconds[src.GetTimeFrameSeconds(tf)] = true
		}
		for sec := range commonSeconds {
			if !srcSeconds[sec] {
				delete(commonSeconds, sec)
			}
		}
	}

	keys := make([]int64, 0, len(commonSeconds))
	for sec := range commonSeconds {
		keys = append(keys, sec)
	}
	slices.Sort(keys)

	result := make([]string, 0, len(keys))
	for _, sec := range keys {
		result = append(result, first.GetTimeFrameString(sec))
	}
	return result
}

func (r *Registry) GetTimeFrameSeconds(sourceNames []string, timeFrame string) int64 {
	for _, name := range sourceNames {
		if src, ok := r.dataSources[name]; ok {
			return src.GetTimeFrameSeconds(timeFrame)
		}
	}
	return 0
}

func (r *Registry) GetValidPairsForSource(sourceName string) ([]sources.TradingPair, error) {
	src, ok := r.dataSources[sourceName]
	if !ok {
		return nil, fmt.Errorf("source '%s' not found in registry", sourceName)
	}
	return src.GetTradingPairs(), nil
}
