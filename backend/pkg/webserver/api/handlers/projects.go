// Package handlers contains the Gin handler functions for the REST API.
package handlers

import (
	"net/http"

	"candle-engine/pkg/chartmanager"
	"candle-engine/pkg/webserver/api/models"

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
	summaries := cm.GetChartProjectsList()

	projects := make([]models.ProjectSummary, len(summaries))
	for i, s := range summaries {
		projects[i] = models.ProjectSummary{ID: s.ID, Title: s.Title}
	}

	c.JSON(http.StatusOK, models.ProjectListResponse{Projects: projects})
}

// GetProject godoc
//
//	@Summary		Get project details
//	@Description	Returns full details for a chart project, including available trading pairs and time frames.
//	@Tags			projects
//	@Produce		json
//	@Param			id	path		string	true	"Project UUID"
//	@Success		200	{object}	models.ProjectDetailResponse
//	@Failure		404	{object}	models.ErrorResponse
//	@Router			/api/v1/projects/{id} [get]
func GetProject(c *gin.Context) {
	id := c.Param("id")

	cm := chartmanager.GetChartManager()
	project, found := cm.GetProjectByID(id)
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
	colors := project.GetColors()

	c.JSON(http.StatusOK, models.ProjectDetailResponse{
		ID:                  id,
		Title:               project.GetTitle(),
		TimeFrame:           project.GetTimeFrame(),
		AvailableTimeFrames: project.GetAvailableTimeFrames(),
		Pairs:               pairs,
		TimeRange:           models.TimeRange{Start: tr.Start, End: tr.End},
		CandleColors: models.CandleColors{
			BullishColor: colors.BullishColor,
			BearishColor: colors.BearishColor,
		},
	})
}
