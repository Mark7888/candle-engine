package routes

import (
	"net/http"
	"testcharting/internal/chartmanager"
	"testcharting/templates"
	templates_chartproject "testcharting/templates/chartproject"

	"github.com/a-h/templ"
)

func LoadRoutes() {
	http.Handle("/", templ.Handler(templates.MainApp()))

	http.Handle("/projects", templ.Handler(templates_chartproject.ChartProjectCollection(chartmanager.GetChartManager().GetChartProjects())))

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
