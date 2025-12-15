package indicators

import "time"

type PlotPointCollection struct {
	Points  []PlotPoint
	Invalid bool
}

func NewPlotPointCollection(points []PlotPoint) PlotPointCollection {
	return PlotPointCollection{
		Points: points,
	}
}

func NewInvalidPlotPointCollection() PlotPointCollection {
	return PlotPointCollection{
		Invalid: true,
	}
}

func (ppc *PlotPointCollection) AddPoints(points ...PlotPoint) {
	ppc.Points = append(ppc.Points, points...)
}

func (ppc *PlotPointCollection) IsInvalid() bool {
	return ppc.Invalid
}

func (ppc *PlotPointCollection) Length() int {
	return len(ppc.Points)
}

type IndicatorLogLevel int

const (
	LogDebug IndicatorLogLevel = iota
	LogInfo
	LogWarning
	LogError
)

type IndicatorLog struct {
	Level   IndicatorLogLevel
	Time    time.Time
	Message string
}

type IndicatorOutputHandler struct {
	Time time.Time

	Data PlotPointCollection
	Logs []IndicatorLog
}

func NewIndicatorOutputHandler(activeEntryTime time.Time) IndicatorOutputHandler {
	return IndicatorOutputHandler{
		Time: activeEntryTime,
		Data: NewPlotPointCollection([]PlotPoint{}),
		Logs: []IndicatorLog{},
	}
}

func (ioh *IndicatorOutputHandler) SetOutput(data PlotPointCollection) {
	ioh.Data = data
}

func (ioh *IndicatorOutputHandler) GetOutput() PlotPointCollection {
	return ioh.Data
}

func (ioh *IndicatorOutputHandler) LogDebug(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Level:   LogDebug,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogInfo(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Level:   LogInfo,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogWarning(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Level:   LogWarning,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogError(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Level:   LogError,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) GetLogs() []IndicatorLog {
	return ioh.Logs
}
