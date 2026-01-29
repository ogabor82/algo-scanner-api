package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	return r
}
