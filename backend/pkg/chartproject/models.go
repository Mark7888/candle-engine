package chartproject

type ProjectRenderData struct {
	ProjectIndex int
	Source       string
	TradingPair  string
	// Some more stuff, like current settings
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
