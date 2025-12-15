package routes

import (
	"net/http"
	"testcharting/templates"

	"github.com/a-h/templ"
)

func LoadRoutes() {
	http.Handle("/", templ.Handler(templates.MainApp()))

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
