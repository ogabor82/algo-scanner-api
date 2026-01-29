package http

import (
	"net/http"
	"os"

	"algosphera/scanner-api/internal/http/handlers"

	"github.com/go-chi/chi/v5"
)

func NewRouter() (http.Handler, error) {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	readerBaseURL := os.Getenv("READER_BASE_URL")
	if readerBaseURL != "" {
		h, err := handlers.NewResultsHandler(readerBaseURL)
		if err != nil {
			return nil, err
		}
		r.Get("/scans/{jobID}/results", h.GetResults)
	}

	return r, nil
}
