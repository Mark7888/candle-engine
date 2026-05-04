package indicators

import "time"

type PlotPoint struct {
	Value   float64
	Color   string
	Invalid bool
}

func NewInvalidPlotPoint() PlotPoint {
	return PlotPoint{
		Value:   -999,
		Invalid: true,
	}
}

type PlotPointCollection struct {
	Points  []PlotPoint
	Invalid bool
}

func NewPlotPointCollection(points ...PlotPoint) PlotPointCollection {
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

type PlotPointCollectionHistory struct {
	Collections []PlotPointCollection
}

func NewPlotPointCollectionHistory(collections []PlotPointCollection) PlotPointCollectionHistory {
	return PlotPointCollectionHistory{
		Collections: collections,
	}
}

func (pcch *PlotPointCollectionHistory) AddCollection(collection PlotPointCollection) {
	pcch.Collections = append(pcch.Collections, collection)
}

func (pcch *PlotPointCollectionHistory) Length() int {
	return len(pcch.Collections)
}

func (pcch *PlotPointCollectionHistory) GetCollection(index int) *PlotPointCollection {
	if index < 0 || index >= len(pcch.Collections) {
		return nil
	}
	reversedIndex := len(pcch.Collections) - 1 - index
	return &pcch.Collections[reversedIndex]
}

func (pcch *PlotPointCollectionHistory) GetAllCollections() []PlotPointCollection {
	return pcch.Collections
}

type IndicatorLogLevel string

const (
	LogDebug   IndicatorLogLevel = "DEBUG"
	LogInfo    IndicatorLogLevel = "INFO"
	LogWarning IndicatorLogLevel = "WARNING"
	LogError   IndicatorLogLevel = "ERROR"
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
		Data: NewPlotPointCollection(),
		Logs: []IndicatorLog{},
	}
}

func (ioh *IndicatorOutputHandler) SetOutput(data PlotPointCollection) {
	ioh.Data = data
}

func (ioh *IndicatorOutputHandler) SetSingleOutput(point PlotPoint) {
	ioh.Data = NewPlotPointCollection(point)
}

func (ioh *IndicatorOutputHandler) GetOutput() PlotPointCollection {
	return ioh.Data
}

func (ioh *IndicatorOutputHandler) LogDebug(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Time:    ioh.Time,
		Level:   LogDebug,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogInfo(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Time:    ioh.Time,
		Level:   LogInfo,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogWarning(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Time:    ioh.Time,
		Level:   LogWarning,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) LogError(message string) {
	ioh.Logs = append(ioh.Logs, IndicatorLog{
		Time:    ioh.Time,
		Level:   LogError,
		Message: message,
	})
}

func (ioh *IndicatorOutputHandler) GetLogs() []IndicatorLog {
	return ioh.Logs
}
