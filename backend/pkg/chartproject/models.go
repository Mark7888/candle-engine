package chartproject

import (
	"candle-engine/pkg/sources"
	"time"
)

// ChartProjectConfig holds the initialization parameters for a ChartProject.
// Zero values use sensible built-in defaults (last 14 days, "1d" timeframe, red/green candle colours).
type ChartProjectConfig struct {
	TimeFrame    string
	TimeRange    sources.TimeRange
	CandleColors sources.CandleColorConfig
}

// RenderConfig holds per-request render parameters. Zero-value fields fall back to the project's configured defaults.
type RenderConfig struct {
	ActivePair   sources.SourcePairId
	TimeFrame    string
	RangeFrom    time.Time
	RangeTo      time.Time
	CandleColors *sources.CandleColorConfig // pointer so nil means "use project default"
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
