package routes

import (
	"net/http"
	"strconv"
	"testcharting/internal/chartmanager"
	"testcharting/templates"
	templates_chartproject "testcharting/templates/chartproject"

	"github.com/a-h/templ"
)

func LoadRoutes() {
	http.Handle("/", templ.Handler(templates.MainApp()))

	http.Handle("/projects/", templ.Handler(templates_chartproject.ChartProjectCollection(chartmanager.GetChartManager().GetChartProjectsList())))
	http.HandleFunc("/project/{index}/", func(w http.ResponseWriter, r *http.Request) {
		index := r.PathValue("index")
		indexInt, err := strconv.Atoi(index)
		if err != nil {
			http.Error(w, "Invalid project index", http.StatusBadRequest)
			return
		}

		cm := chartmanager.GetChartManager()
		project, found := cm.GetChartProject(indexInt)
		if !found {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		renderData := []templates_chartproject.ProjectRenderData{}
		pairs := project.GetLoadedPairs()

		for _, pair := range pairs {
			renderData = append(renderData, templates_chartproject.ProjectRenderData{
				ProjectIndex: indexInt,
				Source:       pair.SourceName,
				TradingPair:  pair.PairId,
			})
		}

		templates_chartproject.ChartProject(renderData).Render(r.Context(), w)
	})
	http.HandleFunc("/project/{index}/render/", func(w http.ResponseWriter, r *http.Request) {
		index := r.PathValue("index")
		source := r.URL.Query().Get("source")
		pair := r.URL.Query().Get("pair")
		templates_chartproject.ProjectPairRender(index+" "+source+" "+pair).Render(r.Context(), w)
	})

	/*
		http.HandleFunc("/func/", func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			fmt.Println("Params: ", params)

			templates.SomeComponent().Render(r.Context(), w)
		})

		handlePostRequest("/post/", func(r *http.Request) {
			// handle post request
		})
	*/
}
