// Package legacy contains the original Templ + HTMX based HTML routes.
// These are preserved for backward compatibility and can be deleted once
// the React frontend is in place.
package legacy

import (
	"candle-engine/internal/chartmanager"
	"candle-engine/pkg/charting/sources"
	"candle-engine/templates"
	templates_chartproject "candle-engine/templates/chartproject"
	"errors"
	"strconv"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all legacy Templ/HTMX HTML routes on the provided Gin router group.
func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/", gin.WrapH(templ.Handler(templates.MainApp())))

	rg.GET("/projects/", func(c *gin.Context) {
		cm := chartmanager.GetChartManager()
		templ.Handler(templates_chartproject.ChartProjectCollection(cm.GetChartProjectsList())).ServeHTTP(c.Writer, c.Request)
	})

	rg.GET("/project/:index/", func(c *gin.Context) {
		index := c.Param("index")
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			templates_chartproject.ChartProjectError(errors.New("Invalid project index format")).Render(c.Request.Context(), c.Writer)
			return
		}

		cm := chartmanager.GetChartManager()
		project, found := cm.GetChartProject(indexInt)
		if !found {
			templates_chartproject.ChartProjectError(errors.New("Project index out of range")).Render(c.Request.Context(), c.Writer)
			return
		}

		renderData := []templates_chartproject.ProjectRenderData{}
		for _, pair := range project.GetLoadedPairs() {
			renderData = append(renderData, templates_chartproject.ProjectRenderData{
				ProjectIndex: indexInt,
				Source:       pair.SourceName,
				TradingPair:  pair.PairId,
			})
		}

		templates_chartproject.ChartProject(renderData).Render(c.Request.Context(), c.Writer)
	})

	rg.GET("/project/:index/render/", func(c *gin.Context) {
		index := c.Param("index")
		source := c.Query("source")
		pair := c.Query("pair")

		cm := chartmanager.GetChartManager()
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			templates_chartproject.ChartProjectError(errors.New("Invalid project index format")).Render(c.Request.Context(), c.Writer)
			return
		}

		project, found := cm.GetChartProject(indexInt)
		if !found {
			templates_chartproject.ChartProjectError(errors.New("Project index out of range")).Render(c.Request.Context(), c.Writer)
			return
		}

		sourcePair := sources.SourcePairId{
			SourceName: source,
			PairId:     pair,
		}

		if err := project.SetActivePair(sourcePair); err != nil {
			templates_chartproject.ChartProjectError(err).Render(c.Request.Context(), c.Writer)
			return
		}

		render, err := project.GetRender()
		if err != nil {
			templates_chartproject.ChartProjectError(err).Render(c.Request.Context(), c.Writer)
			return
		}

		logData := render.GetLogs()
		echartOptions := render.GetEchartsOptions()
		templates_chartproject.ProjectPairRender(echartOptions, logData).Render(c.Request.Context(), c.Writer)
	})
}
