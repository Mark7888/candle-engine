package handlers

import (
	"net/http"
	"time"

	"candle-engine/pkg/chartproject"
	"candle-engine/pkg/chartmanager"
	"candle-engine/pkg/sources"
	"candle-engine/pkg/webserver/api/models"
	apival "candle-engine/pkg/webserver/api/validation"

	"github.com/gin-gonic/gin"
)

// UpdateProjectSettings godoc
//
//	@Summary		Update session settings for a project
//	@Description	Persists the active pair, time frame, date range, and candle colours for the project.
//	@Description	Only non-empty fields in the request body are applied; omitted fields leave existing values unchanged.
//	@Description	After updating, the frontend should re-fetch the render endpoint to get fresh chart data.
//	@Tags			projects
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Project UUID"
//	@Param			body	body		models.UpdateSettingsRequest	true	"Settings to update"
//	@Success		200		{object}	models.ProjectDetailResponse
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		404		{object}	models.ErrorResponse
//	@Router			/api/v1/projects/{id}/settings [patch]
func UpdateProjectSettings(c *gin.Context) {
	id := c.Param("id")

	cm := chartmanager.GetChartManager()
	project, found := cm.GetProjectByID(id)
	if !found {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	cfg := chartproject.ChartProjectConfig{}

	if req.ActivePair != nil {
		cfg.ActivePair = sources.SourcePairId{
			SourceName: req.ActivePair.SourceName,
			PairId:     req.ActivePair.PairID,
		}
	}

	if req.TimeFrame != "" {
		cfg.TimeFrame = req.TimeFrame
	}

	var rangeFrom, rangeTo time.Time
	if req.RangeFrom != "" {
		t, err := time.Parse(time.RFC3339, req.RangeFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid rangeFrom", Details: "rangeFrom must be an RFC3339 timestamp (e.g. 2024-01-01T00:00:00Z)"})
			return
		}
		rangeFrom = t
	}
	if req.RangeTo != "" {
		t, err := time.Parse(time.RFC3339, req.RangeTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid rangeTo", Details: "rangeTo must be an RFC3339 timestamp (e.g. 2024-12-31T23:59:59Z)"})
			return
		}
		rangeTo = t
	}

	// Validate the time range when at least one boundary is provided.
	if !rangeFrom.IsZero() || !rangeTo.IsZero() {
		base := project.GetTimeRange()
		resolved := base
		if !rangeFrom.IsZero() {
			resolved.Start = rangeFrom
		}
		if !rangeTo.IsZero() {
			resolved.End = rangeTo
		}
		if validationErrs := apival.ValidateTimeRange(resolved); len(validationErrs) != 0 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid time range",
				Details: validationErrs[0].Error(),
			})
			return
		}
		cfg.TimeRange = resolved
	}

	if req.CandleBullishColor != "" {
		cfg.CandleColors.BullishColor = req.CandleBullishColor
	}
	if req.CandleBearishColor != "" {
		cfg.CandleColors.BearishColor = req.CandleBearishColor
	}

	project.UpdateConfig(cfg)

	// Return the updated project detail so the frontend can refresh its state in one round trip.
	rawPairs := project.GetLoadedPairs()
	pairs := make([]models.SourcePair, len(rawPairs))
	for i, p := range rawPairs {
		pairs[i] = models.SourcePair{SourceName: p.SourceName, PairID: p.PairId}
	}

	tr := project.GetTimeRange()
	ap := project.GetActivePair()
	colors := project.GetColors()

	c.JSON(http.StatusOK, models.ProjectDetailResponse{
		ID:                  id,
		Title:               project.GetTitle(),
		TimeFrame:           project.GetTimeFrame(),
		AvailableTimeFrames: project.GetAvailableTimeFrames(),
		Pairs:               pairs,
		TimeRange:           models.TimeRange{Start: tr.Start, End: tr.End},
		ActivePair:          models.SourcePair{SourceName: ap.SourceName, PairID: ap.PairId},
		CandleColors: models.CandleColors{
			BullishColor: colors.BullishColor,
			BearishColor: colors.BearishColor,
		},
	})
}
