package handlers

import (
	"net/http"
	"strconv"
	"time"

	"candle-engine/pkg/chartmanager"
	"candle-engine/pkg/chartproject"
	"candle-engine/pkg/indicators"
	"candle-engine/pkg/sources"
	"candle-engine/pkg/webserver/api/models"
	apival "candle-engine/pkg/webserver/api/validation"

	"github.com/gin-gonic/gin"
)

// indicatorTypeName converts the internal IndicatorType int to a stable string
// name that the frontend can use to select the correct rendering strategy.
func indicatorTypeName(t indicators.IndicatorType) string {
	switch t {
	case indicators.IndicatorTypeLine:
		return "line"
	case indicators.IndicatorTypeStepLine:
		return "stepLine"
	case indicators.IndicatorTypeBarChart:
		return "bar"
	case indicators.IndicatorTypeTable:
		return "table"
	case indicators.IndicatorTypeAreaChart:
		return "area"
	case indicators.IndicatorTypePieChart:
		return "pieChart"
	case indicators.IndicatorTypeCandleStick:
		return "candleStick"
	default:
		return "unknown"
	}
}

// indicatorConfig extracts only the type-relevant configuration sub-object so
// the frontend receives a clean, narrowly-typed config instead of a union struct.
func indicatorConfig(t indicators.IndicatorType, cfg indicators.IndicatorConfig) interface{} {
	switch t {
	case indicators.IndicatorTypeLine:
		return cfg.LineIndicator
	case indicators.IndicatorTypeStepLine:
		return cfg.StepLineIndicator
	case indicators.IndicatorTypeBarChart:
		return cfg.BarChartIndicator
	case indicators.IndicatorTypeTable:
		return cfg.TableIndicator
	case indicators.IndicatorTypeAreaChart:
		return cfg.AreaChartIndicator
	case indicators.IndicatorTypePieChart:
		return cfg.PieChartIndicator
	case indicators.IndicatorTypeCandleStick:
		return cfg.CandleStickIndicator
	default:
		return nil
	}
}

// emptyRenderResponse converts a RenderableChartProject that contains only
// ProjectErrors (no candle/indicator data) into a minimal ProjectRenderResponse.
// This is used to surface validation or early-exit errors to the frontend
// without returning an HTTP error status, so that callers always receive the
// same response shape.
func emptyRenderResponse(render chartproject.RenderableChartProject) models.ProjectRenderResponse {
	var logs []models.LogEntry
	for _, projErr := range render.ProjectErrors {
		logs = append(logs, models.LogEntry{
			Indicator: "ProjectErrors",
			Level:     "ERROR",
			Message:   projErr.Error(),
		})
	}
	if logs == nil {
		logs = []models.LogEntry{}
	}
	return models.ProjectRenderResponse{
		Candles:    []models.CandleSeries{},
		Indicators: []models.IndicatorData{},
		Logs:       logs,
	}
}

// RenderProject godoc
//
//	@Summary		Render chart data for a project
//	@Description	Fetches OHLCV data and computes all configured indicators for the requested trading pair.
//	@Description	The response contains raw candle and indicator data for the frontend to render.
//	@Description	Indicator PlotPoints arrays are index-aligned with the active pair's candle entries.
//	@Description	Optional query parameters timeFrame, rangeFrom, and rangeTo override the project defaults
//	@Description	for this request only; no server-side state is mutated.
//	@Tags			projects
//	@Produce		json
//	@Param			id			path		int		true	"Project ID (zero-based index)"
//	@Param			source		query		string	true	"Data source name (e.g. 'Binance')"
//	@Param			pair		query		string	true	"Trading pair ID (e.g. 'BTCUSDC')"
//	@Param			timeFrame	query		string	false	"Override time frame (e.g. '1h', '4h', '1d')"
//	@Param			rangeFrom	query		string	false	"Override range start as RFC3339 timestamp"
//	@Param			rangeTo		query		string	false	"Override range end as RFC3339 timestamp"
//	@Success		200		{object}	models.ProjectRenderResponse
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		404		{object}	models.ErrorResponse
//	@Failure		500		{object}	models.ErrorResponse
//	@Router			/api/v1/projects/{id}/render [get]
func RenderProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid project id", Details: "id must be a non-negative integer"})
		return
	}

	source := c.Query("source")
	pair := c.Query("pair")
	if source == "" || pair == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing query parameters", Details: "both 'source' and 'pair' are required"})
		return
	}

	cm := chartmanager.GetChartManager()
	// GetChartProject returns a value copy so each request works on its own
	// instance, making concurrent pair switching safe.
	project, found := cm.GetChartProject(id)
	if !found {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	// Build a RenderConfig from query parameters; unset fields fall back to
	// the project's configured defaults inside GetRender.
	renderCfg := chartproject.RenderConfig{
		ActivePair: sources.SourcePairId{SourceName: source, PairId: pair},
	}

	if tf := c.Query("timeFrame"); tf != "" {
		renderCfg.TimeFrame = tf
	}

	if rangeFromStr := c.Query("rangeFrom"); rangeFromStr != "" {
		t, parseErr := time.Parse(time.RFC3339, rangeFromStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid rangeFrom", Details: "rangeFrom must be an RFC3339 timestamp (e.g. 2024-01-01T00:00:00Z)"})
			return
		}
		renderCfg.RangeFrom = t
	}

	if rangeToStr := c.Query("rangeTo"); rangeToStr != "" {
		t, parseErr := time.Parse(time.RFC3339, rangeToStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid rangeTo", Details: "rangeTo must be an RFC3339 timestamp (e.g. 2024-12-31T23:59:59Z)"})
			return
		}
		renderCfg.RangeTo = t
	}

	if bullish := c.Query("candleBullishColor"); bullish != "" {
		if renderCfg.CandleColors == nil {
			renderCfg.CandleColors = &sources.CandleColorConfig{}
		}
		renderCfg.CandleColors.BullishColor = bullish
	}

	if bearish := c.Query("candleBearishColor"); bearish != "" {
		if renderCfg.CandleColors == nil {
			renderCfg.CandleColors = &sources.CandleColorConfig{}
		}
		renderCfg.CandleColors.BearishColor = bearish
	}

	// Semantic validation of the resolved time range when at least one boundary
	// was supplied. Start from the project defaults and merge in the overrides
	// so the validator sees the complete range.
	if !renderCfg.RangeFrom.IsZero() || !renderCfg.RangeTo.IsZero() {
		base := project.GetTimeRange()
		resolved := base
		if !renderCfg.RangeFrom.IsZero() {
			resolved.Start = renderCfg.RangeFrom
		}
		if !renderCfg.RangeTo.IsZero() {
			resolved.End = renderCfg.RangeTo
		}
		if validationErrs := apival.ValidateTimeRange(resolved); len(validationErrs) != 0 {
			partial := chartproject.NewRenderableChartProject()
			partial.ProjectErrors = validationErrs
			c.JSON(http.StatusOK, emptyRenderResponse(partial))
			return
		}
	}

	render, err := project.GetRender(renderCfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to render chart", Details: err.Error()})
		return
	}

	// --- Candle data ---
	// Only include the active pair's candle data. Iterating over a Go map yields
	// non-deterministic order, so sending all pairs would let the frontend pick
	// the wrong series. The indicators are already index-aligned with the active
	// pair, so only that series is meaningful here.
	var candles []models.CandleSeries
	if entries, ok := render.CandleData[render.ActivePair]; ok {
		candleEntries := make([]models.CandleEntry, len(entries))
		for i, e := range entries {
			candleEntries[i] = models.CandleEntry{
				Timestamp: e.Timestamp,
				Open:      e.Open,
				High:      e.High,
				Low:       e.Low,
				Close:     e.Close,
				Volume:    e.Volume,
				Invalid:   e.Invalid,
			}
		}
		candles = []models.CandleSeries{{
			SourceName: render.ActivePair.SourceName,
			PairID:     render.ActivePair.PairId,
			Entries:    candleEntries,
		}}
	} else {
		candles = []models.CandleSeries{}
	}

	// --- Indicator data ---
	// Flatten the per-chart map into a single slice; TargetChart is included on
	// each indicator so the frontend can re-group them as needed.
	var indicatorData []models.IndicatorData
	var logs []models.LogEntry

	for _, indicatorGroup := range render.IndicatorCharts {
		for _, ind := range indicatorGroup {
			// Convert plot point collections.
			plotPoints := make([]models.PlotPointCollection, len(ind.PlotPoints))
			for i, coll := range ind.PlotPoints {
				points := make([]models.PlotPoint, len(coll.Points))
				for j, pt := range coll.Points {
					points[j] = models.PlotPoint{
						Value:   pt.Value,
						Color:   pt.Color,
						Invalid: pt.Invalid,
					}
				}
				plotPoints[i] = models.PlotPointCollection{
					Points:  points,
					Invalid: coll.Invalid,
				}
			}

			// Collect error strings (errors.Error() is the only safe serialization).
			var errStrings []string
			for _, renderErr := range ind.RenderingErrors {
				errStrings = append(errStrings, renderErr.Error())
			}

			indicatorData = append(indicatorData, models.IndicatorData{
				Name:        ind.Name,
				Type:        indicatorTypeName(ind.Type),
				TargetChart: ind.IndicatorConfig.TargetChart,
				Config:      indicatorConfig(ind.Type, ind.IndicatorConfig),
				PlotPoints:  plotPoints,
				DataError:   ind.DataError,
				Errors:      errStrings,
			})

			// Collect logs for this indicator.
			if ind.DataError {
				for _, renderErr := range ind.RenderingErrors {
					logs = append(logs, models.LogEntry{
						Indicator: ind.Name,
						Level:     "ERROR",
						Message:   "data inconsistency: " + renderErr.Error(),
					})
				}
			}
			for _, logEntry := range ind.Logs {
				logs = append(logs, models.LogEntry{
					Indicator: ind.Name,
					Level:     string(logEntry.Level),
					Time:      logEntry.Time,
					Message:   logEntry.Message,
				})
			}
		}
	}

	if indicatorData == nil {
		indicatorData = []models.IndicatorData{}
	}

	for _, projErr := range render.ProjectErrors {
		logs = append(logs, models.LogEntry{
			Indicator: "ProjectErrors",
			Level:     "ERROR",
			Message:   projErr.Error(),
		})
	}

	if logs == nil {
		logs = []models.LogEntry{}
	}

	ap := render.ActivePair
	renderTr := render.TimeRange

	c.JSON(http.StatusOK, models.ProjectRenderResponse{
		Title: render.Title,
		TimeRange: models.TimeRange{
			Start: renderTr.Start,
			End:   renderTr.End,
		},
		AvailableTimeFrames: render.AvailableTimeFrames,
		ActiveTimeFrame:     render.ActiveTimeFrame,
		ActivePair:          models.SourcePair{SourceName: ap.SourceName, PairID: ap.PairId},
		CandleColors: models.CandleColors{
			BullishColor: render.CandleColors.BullishColor,
			BearishColor: render.CandleColors.BearishColor,
		},
		Candles:    candles,
		Indicators: indicatorData,
		Logs:       logs,
	})
}
