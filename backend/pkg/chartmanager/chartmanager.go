package chartmanager

import (
	"candle-engine/pkg/chartproject"
	"candle-engine/pkg/registry"
	"candle-engine/pkg/sources"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// ProjectSummary is a brief representation of a chart project for list responses.
type ProjectSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ChartManager manages a collection of chart projects in a thread-safe manner.
type ChartManager struct {
	chartProjects map[string]*chartproject.ChartProject
	mu            sync.RWMutex
	reg           *registry.Registry
}

func NewChartManager(reg *registry.Registry) *ChartManager {
	return &ChartManager{
		chartProjects: make(map[string]*chartproject.ChartProject),
		reg:           reg,
	}
}

func (cm *ChartManager) CreateProject(title string, cfg chartproject.ChartProjectConfig) *chartproject.ChartProject {
	id := uuid.NewString()
	project := chartproject.NewChartProject(id, title, cfg)
	project.SetRegistry(cm.reg)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.chartProjects[id] = project
	return project
}

func (cm *ChartManager) AddChartProject(project *chartproject.ChartProject) {
	if project.GetID() == "" {
		project.ID = uuid.NewString()
	}
	project.SetRegistry(cm.reg)

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.chartProjects[project.GetID()] = project
}

func (cm *ChartManager) DeleteProject(id string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	_, exists := cm.chartProjects[id]
	if exists {
		delete(cm.chartProjects, id)
	}
	return exists
}

func (cm *ChartManager) GetProjectByID(id string) (*chartproject.ChartProject, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	project, exists := cm.chartProjects[id]
	return project, exists
}

func (cm *ChartManager) GetChartProjectsList() []ProjectSummary {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	summaries := make([]ProjectSummary, 0, len(cm.chartProjects))
	for _, project := range cm.chartProjects {
		summaries = append(summaries, ProjectSummary{ID: project.GetID(), Title: project.GetTitle()})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Title != summaries[j].Title {
			return summaries[i].Title < summaries[j].Title
		}
		return summaries[i].ID < summaries[j].ID
	})
	return summaries
}

func (cm *ChartManager) AddPairToProject(id string, pair sources.SourcePairId) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	project, exists := cm.chartProjects[id]
	if !exists {
		return false
	}
	project.AddPair(pair)
	return true
}

func (cm *ChartManager) RemovePairFromProject(id string, pair sources.SourcePairId) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	project, exists := cm.chartProjects[id]
	if !exists {
		return false
	}
	project.RemovePair(pair)
	return true
}

func (cm *ChartManager) AddIndicatorToProject(id, indicatorName string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	project, exists := cm.chartProjects[id]
	if !exists {
		return false
	}
	project.AddIndicatorByName(indicatorName)
	return true
}

func (cm *ChartManager) RemoveIndicatorFromProject(id, indicatorName string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	project, exists := cm.chartProjects[id]
	if !exists {
		return false
	}
	project.RemoveIndicatorByName(indicatorName)
	return true
}
