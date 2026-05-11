package chartmanager

import (
	"candle-engine/pkg/registry"
	"sync"
)

var chartManager *ChartManager
var chartManagerMutex sync.RWMutex

// InitChartManager creates the singleton ChartManager; must be called once at startup before GetChartManager.
func InitChartManager(reg *registry.Registry) *ChartManager {
	chartManagerMutex.Lock()
	defer chartManagerMutex.Unlock()

	if chartManager != nil {
		panic("chartmanager: InitChartManager called more than once")
	}
	chartManager = NewChartManager(reg)
	return chartManager
}

// GetChartManager returns the singleton ChartManager; panics if InitChartManager was not called first.
func GetChartManager() *ChartManager {
	chartManagerMutex.RLock()
	defer chartManagerMutex.RUnlock()

	if chartManager == nil {
		panic("chartmanager: GetChartManager called before InitChartManager")
	}
	return chartManager
}
