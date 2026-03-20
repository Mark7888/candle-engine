// Package handlers contains the Gin handler functions for the REST API.
package handlers

import (
	"net/http"
	"strconv"

	"candle-engine/internal/chartmanager"
	"candle-engine/internal/webserver/api/models"

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

	ap := project.GetActivePair()
	tr := project.GetTimeRange()

	c.JSON(http.StatusOK, models.ProjectDetailResponse{
		ID:                  id,
		Title:               project.GetTitle(),
		TimeFrame:           project.GetTimeFrame(),
		AvailableTimeFrames: project.GetAvailableTimeFrames(),
		Pairs:               pairs,
		ActivePair:          models.SourcePair{SourceName: ap.SourceName, PairID: ap.PairId},
		TimeRange:           models.TimeRange{Start: tr.Start, End: tr.End},
	})
}
