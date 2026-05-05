package chartmanager

import (
	"candle-engine/pkg/chartproject"
	"sync"
)

// ChartManager manages a collection of chart projects in a thread-safe manner
type ChartManager struct {
	chartProjects []chartproject.ChartProject
	mu            sync.RWMutex
}

// NewChartManager creates a new ChartManager instance
func NewChartManager() *ChartManager {
	return &ChartManager{
		chartProjects: make([]chartproject.ChartProject, 0),
	}
}

// GetChartProjects returns a copy of all chart projects
func (cm *ChartManager) GetChartProjects() []chartproject.ChartProject {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a copy of the slice
	projects := make([]chartproject.ChartProject, len(cm.chartProjects))
	copy(projects, cm.chartProjects)
	return projects
}

func (cm *ChartManager) GetChartProjectsList() (titles []string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, project := range cm.chartProjects {
		titles = append(titles, project.GetTitle())
	}
	return titles
}

// GetChartProject returns a copy of a chart project by index
func (cm *ChartManager) GetChartProject(index int) (chartproject.ChartProject, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if index < 0 || index >= len(cm.chartProjects) {
		return chartproject.ChartProject{}, false
	}
	return cm.chartProjects[index], true
}

// AddChartProject adds a new chart project to the manager
func (cm *ChartManager) AddChartProjects(project ...chartproject.ChartProject) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.chartProjects = append(cm.chartProjects, project...)
}

// RemoveChartProject removes a chart project by index
func (cm *ChartManager) RemoveChartProject(index int) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if index < 0 || index >= len(cm.chartProjects) {
		return false
	}

	cm.chartProjects = append(cm.chartProjects[:index], cm.chartProjects[index+1:]...)
	return true
}

// Count returns the number of chart projects
func (cm *ChartManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.chartProjects)
}

// Clear removes all chart projects
func (cm *ChartManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.chartProjects = make([]chartproject.ChartProject, 0)
}
