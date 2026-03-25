// Package handlers contains the Gin handler functions for the REST API.
package handlers

import (
	"net/http"
	"slices"
	"strconv"

	"candle-engine/internal/chartmanager"
	"candle-engine/internal/webserver/api/models"
	"candle-engine/pkg/charting/sources"

	"github.com/gin-gonic/gin"
)

// ListProjects godoc
//
//	@Summary		List all chart projects
//	@Description	Returns a summary list of all available chart projects.
//	@Tags			projects
//	@Produce		json
//	@Success		200	{object}	models.ProjectListResponse
//	@Router			/api/v1/projects [get]
func ListProjects(c *gin.Context) {
	cm := chartmanager.GetChartManager()
	titles := cm.GetChartProjectsList()

	projects := make([]models.ProjectSummary, len(titles))
	for i, title := range titles {
		projects[i] = models.ProjectSummary{ID: i, Title: title}
	}

	c.JSON(http.StatusOK, models.ProjectListResponse{Projects: projects})
}

// GetProject godoc
//
//	@Summary		Get project details
//	@Description	Returns full details for a chart project, including available trading pairs and time frames.
//	@Tags			projects
//	@Produce		json
//	@Param			id	path		int	true	"Project ID (zero-based index)"
//	@Success		200	{object}	models.ProjectDetailResponse
//	@Failure		400	{object}	models.ErrorResponse
//	@Failure		404	{object}	models.ErrorResponse
//	@Router			/api/v1/projects/{id} [get]
func GetProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid project id", Details: "id must be a non-negative integer"})
		return
	}

	cm := chartmanager.GetChartManager()
	project, found := cm.GetChartProject(id)
	if !found {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	rawPairs := project.GetLoadedPairs()
	pairs := make([]models.SourcePair, len(rawPairs))
	for i, p := range rawPairs {
		pairs[i] = models.SourcePair{SourceName: p.SourceName, PairID: p.PairId}
	}

	tr := project.GetTimeRange()

	c.JSON(http.StatusOK, models.ProjectDetailResponse{
		ID:                  id,
		Title:               project.GetTitle(),
		TimeFrame:           project.GetTimeFrame(),
		AvailableTimeFrames: project.GetAvailableTimeFrames(),
		Pairs:               pairs,
		TimeRange:           models.TimeRange{Start: tr.Start, End: tr.End},
	})
}

// UpdateProjectSettings godoc
//
//	@Summary		Update project settings
//	@Description	Updates the timeFrame and/or timeRange for a chart project. Only supplied fields are changed.
//	@Tags			projects
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"Project ID (zero-based index)"
//	@Param			body	body		models.UpdateSettingsRequest	true	"Settings to update"
//	@Success		200		{object}	models.ProjectDetailResponse
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		404		{object}	models.ErrorResponse
//	@Router			/api/v1/projects/{id}/settings [patch]
func UpdateProjectSettings(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid project id", Details: "id must be a non-negative integer"})
		return
	}

	var req models.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body", Details: err.Error()})
		return
	}

	cm := chartmanager.GetChartManager()
	project, found := cm.GetChartProject(id)
	if !found {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	// Validate timeFrame against available options for this project.
	if req.TimeFrame != nil {
		available := project.GetAvailableTimeFrames()
		if !slices.Contains(available, *req.TimeFrame) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid timeFrame",
				Details: "timeFrame is not available for this project",
			})
			return
		}
	}

	// Validate timeRange if provided.
	if req.TimeRange != nil && !req.TimeRange.Start.Before(req.TimeRange.End) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid timeRange",
			Details: "start must be before end",
		})
		return
	}

	var tr *sources.TimeRange
	if req.TimeRange != nil {
		tr = &sources.TimeRange{Start: req.TimeRange.Start, End: req.TimeRange.End}
	}

	if !cm.UpdateProjectSettings(id, req.TimeFrame, tr) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "project not found"})
		return
	}

	// Return the updated project detail so the frontend can refresh its state.
	updated, _ := cm.GetChartProject(id)
	rawPairs := updated.GetLoadedPairs()
	pairs := make([]models.SourcePair, len(rawPairs))
	for i, p := range rawPairs {
		pairs[i] = models.SourcePair{SourceName: p.SourceName, PairID: p.PairId}
	}
	utr := updated.GetTimeRange()

	c.JSON(http.StatusOK, models.ProjectDetailResponse{
		ID:                  id,
		Title:               updated.GetTitle(),
		TimeFrame:           updated.GetTimeFrame(),
		AvailableTimeFrames: updated.GetAvailableTimeFrames(),
		Pairs:               pairs,
		TimeRange:           models.TimeRange{Start: utr.Start, End: utr.End},
	})
}
