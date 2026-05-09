package registry

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"maps"
	"slices"
)

// This package will provide a registry for indicator and data source implementations.
// It will allow us to register all new indicators and data sources once at startup, and load/unload them in the application dynamically.

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

// Indicators
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

// Data Sources
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
