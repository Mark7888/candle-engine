package chartproject

import (
	"candle-engine/pkg/sources"
)

// ChartProjectConfig holds the initialization parameters for a ChartProject.
// Zero values use sensible built-in defaults (last 14 days, "1d" timeframe, red/green candle colours).
type ChartProjectConfig struct {
	TimeFrame    string                    `json:"timeFrame,omitempty"`
	TimeRange    sources.TimeRange         `json:"timeRange,omitempty"`
	CandleColors sources.CandleColorConfig `json:"candleColors,omitempty"`
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
