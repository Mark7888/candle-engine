package chartproject

import (
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
)

type RenderableIndicator struct {
	Name            string
	Type            indicators.IndicatorType
	IndicatorConfig indicators.IndicatorConfig
	PlotPoints      []indicators.PlotPointCollection

	DataError       bool // if true, indicates that there was an error in data rendering (e.g., inconsistent data lengths)
	RenderingErrors []error
	Logs            []indicators.IndicatorLog
}

// This will contain all the data needed to render the chart and indicators in the frontend
type RenderableChartProject struct {
	Title               string
	TimeRange           sources.TimeRange
	AvailableTimeFrames []string
	ActiveTimeFrame     string
	LoadedPairs         []sources.SourcePairId
	ActivePair          sources.SourcePairId

	CandleData   map[sources.SourcePairId][]sources.ChartEntry
	CandleColors sources.CandleColorConfig

	IndicatorCharts map[string][]RenderableIndicator

	ProjectErrors []error // errors that prevented the project from rendering (e.g. source fetch failures)

	// Other metadata
}

func NewRenderableChartProject() RenderableChartProject {
	return RenderableChartProject{}
}

func (rcp *RenderableChartProject) GetLogs() map[string][]ProjectLogData {
	var logList map[string][]ProjectLogData
	logList = make(map[string][]ProjectLogData)

	if len(rcp.ProjectErrors) > 0 {
		logList["ProjectErrors"] = []ProjectLogData{}
		for _, err := range rcp.ProjectErrors {
			logList["ProjectErrors"] = append(logList["ProjectErrors"], ProjectLogData{
				LogLevel: LogLevelError,
				Message:  err.Error(),
			})
		}
	}

	for _, indicators := range rcp.IndicatorCharts {
		for _, indicator := range indicators {
			// First collect rendering errors
			if indicator.DataError {
				var renderErrorLogs []ProjectLogData
				for _, err := range indicator.RenderingErrors {
					renderErrorLogs = append(renderErrorLogs, ProjectLogData{
						LogLevel:      LogLevelError,
						TimeFormatted: "",
						Message:       "Rendering error for indicator " + indicator.Name + ": " + err.Error(),
					})
				}

				if logList["RenderingErrors"] == nil {
					logList["RenderingErrors"] = []ProjectLogData{}
				}
				logList["RenderingErrors"] = append(logList["RenderingErrors"], renderErrorLogs...)
			}

			var indicatorLogList []ProjectLogData
			for _, logEntry := range indicator.Logs {
				indicatorLogList = append(indicatorLogList, ProjectLogData{
					LogLevel:      LogLevelEnum(logEntry.Level),
					TimeFormatted: logEntry.Time.Format("2006-01-02 15:04:05"),
					Message:       logEntry.Message,
				})
			}

			if logList[indicator.Name] == nil {
				logList[indicator.Name] = []ProjectLogData{}
			}
			logList[indicator.Name] = append(logList[indicator.Name], indicatorLogList...)
		}
	}

	return logList
}
