package chartmanager

import (
	"sync"
)

var chartManager *ChartManager
var chartManagerMutex sync.RWMutex

// TODO - this is a temporary singleton pattern for ChartManager
func GetChartManager() *ChartManager {
	chartManagerMutex.Lock()
	defer chartManagerMutex.Unlock()

	if chartManager == nil {
		chartManager = NewChartManager()
	}
	return chartManager
}
