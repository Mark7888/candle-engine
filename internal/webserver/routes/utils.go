package routes

import "net/http"

func handlePostRequest(uri string, handler func(r *http.Request)) {
	http.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		handler(r)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	})
}
