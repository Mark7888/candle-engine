package chartproject

import (
	"candle-engine/pkg/sources"
	"time"
)

// ChartProjectConfig holds the initialization parameters for a ChartProject.
// Zero values use sensible built-in defaults (last 14 days, "1d" timeframe, red/green candle colours).
type ChartProjectConfig struct {
	TimeFrame    string                    `json:"timeFrame,omitempty"`
	TimeRange    sources.TimeRange         `json:"timeRange,omitempty"`
	CandleColors sources.CandleColorConfig `json:"candleColors,omitempty"`
	ActivePair   sources.SourcePairId      `json:"activePair,omitempty"`
}

// RenderConfig holds per-request render parameters. Zero-value fields fall back to the project's configured defaults.
type RenderConfig struct {
	ActivePair   sources.SourcePairId       `json:"activePair,omitempty"`
	TimeFrame    string                     `json:"timeFrame,omitempty"`
	RangeFrom    time.Time                  `json:"rangeFrom,omitempty"`
	RangeTo      time.Time                  `json:"rangeTo,omitempty"`
	CandleColors *sources.CandleColorConfig `json:"candleColors,omitempty"` // pointer so nil means "use project default"
}

type LogLevelEnum string

const (
	LogLevelDebug LogLevelEnum = "DEBUG"
	LogLevelInfo  LogLevelEnum = "INFO"
	LogLevelWarn  LogLevelEnum = "WARN"
	LogLevelError LogLevelEnum = "ERROR"
)

type ProjectLogData struct {
	LogLevel      LogLevelEnum
	TimeFormatted string
	Message       string
}
