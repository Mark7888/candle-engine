// Package models defines the JSON request and response types for the REST API.
// These types are designed for use with React + Axios on the frontend.
package models

import "time"

// ProjectSummary is a brief representation of a chart project.
type ProjectSummary struct {
	ID    string `json:"id"`
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
	ID                  string       `json:"id"`
	Title               string       `json:"title"`
	TimeFrame           string       `json:"timeFrame"`
	AvailableTimeFrames []string     `json:"availableTimeFrames"`
	Pairs               []SourcePair `json:"pairs"`
	TimeRange           TimeRange    `json:"timeRange"`
	ActivePair          SourcePair   `json:"activePair"`
	CandleColors        CandleColors `json:"candleColors"`
}

// CandleEntry is a single OHLCV data point.
type CandleEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Invalid   bool      `json:"invalid"`
}

// CandleSeries holds the OHLCV data for one source/pair combination.
type CandleSeries struct {
	SourceName string        `json:"sourceName"`
	PairID     string        `json:"pairId"`
	Entries    []CandleEntry `json:"entries"`
}

// CandleColors holds the configurable color scheme for candle rendering.
type CandleColors struct {
	BullishColor string `json:"bullishColor"`
	BearishColor string `json:"bearishColor"`
}

// PlotPoint is a single value in an indicator series.
type PlotPoint struct {
	Value   float64 `json:"value"`
	Color   string  `json:"color"`
	Invalid bool    `json:"invalid"`
}

// PlotPointCollection holds all plot values for a single time step.
// Each collection is index-aligned with the entries of the active pair's CandleSeries.
type PlotPointCollection struct {
	Points  []PlotPoint `json:"points"`
	Invalid bool        `json:"invalid"`
}

// IndicatorData holds the computed output for one indicator.
// Config contains type-specific configuration fields; the relevant sub-object
// is selected based on the Type field (e.g. "line", "stepLine", "bar", "area", "table").
type IndicatorData struct {
	Name        string                `json:"name"`
	Type        string                `json:"type"`
	TargetChart string                `json:"targetChart"`
	Config      interface{}           `json:"config"`
	PlotPoints  []PlotPointCollection `json:"plotPoints"`
	DataError   bool                  `json:"dataError"`
	Errors      []string              `json:"errors,omitempty"`
}

// LogEntry represents a single log line produced by an indicator.
type LogEntry struct {
	Indicator string    `json:"indicator"`
	Level     string    `json:"level"`
	Time      time.Time `json:"time"`
	Message   string    `json:"message"`
}

// ProjectRenderResponse is returned by GET /api/v1/projects/:id/render.
// It contains raw OHLCV data and computed indicator values so the frontend
// can render the chart without any server-side ECharts coupling.
// Indicator PlotPoints arrays are index-aligned with the entries in the
// active pair's CandleSeries.
type ProjectRenderResponse struct {
	Title               string          `json:"title"`
	TimeRange           TimeRange       `json:"timeRange"`
	AvailableTimeFrames []string        `json:"availableTimeFrames"`
	ActiveTimeFrame     string          `json:"activeTimeFrame"`
	ActivePair          SourcePair      `json:"activePair"`
	CandleColors        CandleColors    `json:"candleColors"`
	Candles             []CandleSeries  `json:"candles"`
	Indicators          []IndicatorData `json:"indicators"`
	Logs                []LogEntry      `json:"logs"`
}

// HealthResponse is returned by GET /api/v1/health.
type HealthResponse struct {
	Status     string    `json:"status"`
	ServerTime time.Time `json:"serverTime"`
}

// ErrorResponse is returned for any API error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
