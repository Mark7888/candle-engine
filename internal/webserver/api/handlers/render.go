package handlers

import (
	"net/http"
	"strconv"

	"candle-engine/internal/chartmanager"
	"candle-engine/internal/webserver/api/models"
	"candle-engine/pkg/charting/sources"

	"github.com/gin-gonic/gin"
)

// RenderProject godoc
//
//	@Summary		Render chart data for a project
//	@Description	Fetches OHLCV data and computes all configured indicators for the requested trading pair.
//	@Description	The returned `echartOptions` object can be passed directly to an Apache ECharts instance.
//	@Tags			projects
//	@Produce		json
//	@Param			id		path		int		true	"Project ID (zero-based index)"
//	@Param			source	query		string	true	"Data source name (e.g. 'Binance')"
//	@Param			pair	query		string	true	"Trading pair ID (e.g. 'BTCUSDC')"
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
	// GetChartProject returns a value copy so each request has its own working copy,
	// which is safe for concurrent use when setting the active pair.
	project, found := cm.GetChartProject(id)
	if !found {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	if err := project.SetActivePair(sources.SourcePairId{SourceName: source, PairId: pair}); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid trading pair", Details: err.Error()})
		return
	}

	render, err := project.GetRender()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to render chart", Details: err.Error()})
		return
	}

	// Collect all indicator logs and rendering errors into a flat list.
	var logs []models.LogEntry
	for _, indicatorGroup := range render.IndicatorCharts {
		for _, indicator := range indicatorGroup {
			if indicator.DataError {
				for _, renderErr := range indicator.RenderingErrors {
					logs = append(logs, models.LogEntry{
						Indicator: indicator.Name,
						Level:     "ERROR",
						Message:   "data inconsistency: " + renderErr.Error(),
					})
				}
			}
			for _, logEntry := range indicator.Logs {
				logs = append(logs, models.LogEntry{
					Indicator: indicator.Name,
					Level:     string(logEntry.Level),
					Time:      logEntry.Time,
					Message:   logEntry.Message,
				})
			}
		}
	}
	if logs == nil {
		logs = []models.LogEntry{}
	}

	ap := render.ActivePair
	c.JSON(http.StatusOK, models.ProjectRenderResponse{
		Title:         render.Title,
		ActivePair:    models.SourcePair{SourceName: ap.SourceName, PairID: ap.PairId},
		EchartOptions: render.GetEchartsOptions(),
		Logs:          logs,
	})
}
