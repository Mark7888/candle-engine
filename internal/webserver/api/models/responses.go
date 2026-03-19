// Package models defines the JSON request and response types for the REST API.
// These types are designed for use with React + Axios on the frontend.
package models

import "time"

// ProjectSummary is a brief representation of a chart project.
type ProjectSummary struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ProjectListResponse is returned by GET /api/v1/projects.
type ProjectListResponse struct {
	Projects []ProjectSummary `json:"projects"`
}

// SourcePair identifies a trading pair from a specific data source.
type SourcePair struct {
	SourceName string `json:"sourceName"`
	PairID     string `json:"pairId"`
}

// TimeRange represents a start/end time window.
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ProjectDetailResponse is returned by GET /api/v1/projects/:id.
type ProjectDetailResponse struct {
	ID                  int          `json:"id"`
	Title               string       `json:"title"`
	TimeFrame           string       `json:"timeFrame"`
	AvailableTimeFrames []string     `json:"availableTimeFrames"`
	Pairs               []SourcePair `json:"pairs"`
	ActivePair          SourcePair   `json:"activePair"`
	TimeRange           TimeRange    `json:"timeRange"`
}

// LogEntry represents a single log line produced by an indicator.
type LogEntry struct {
	Indicator string    `json:"indicator"`
	Level     string    `json:"level"`
	Time      time.Time `json:"time"`
	Message   string    `json:"message"`
}

// ProjectRenderResponse is returned by GET /api/v1/projects/:id/render.
// The EchartOptions field can be passed directly to an Apache ECharts instance.
type ProjectRenderResponse struct {
	Title         string      `json:"title"`
	ActivePair    SourcePair  `json:"activePair"`
	EchartOptions interface{} `json:"echartOptions"`
	Logs          []LogEntry  `json:"logs"`
}

// HealthResponse is returned by GET /api/v1/health.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse is returned for any API error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
