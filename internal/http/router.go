package http

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"algosphera/scanner-api/internal/http/handlers"
)

func NewRouter(db *pgxpool.Pool) (http.Handler, error) {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	// job create
	scans := handlers.NewScansHandler(db)
	r.Post("/scans", scans.CreateScan)
	r.Get("/scans/{jobID}", scans.GetScan)

	// results proxy
	readerBaseURL := os.Getenv("READER_BASE_URL")
	if readerBaseURL != "" {
		h, err := handlers.NewResultsHandler(readerBaseURL)
		if err != nil {
			return nil, err
		}
		runs, err := handlers.NewRunsHandler(readerBaseURL)
		if err != nil {
			return nil, err
		}
		r.Get("/scans/{jobID}/results", h.GetResults)
		r.Get("/runs", runs.GetRuns)
	}

	return r, nil
}
