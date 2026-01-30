package http

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"algosphera/scanner-api/internal/http/handlers"
	tickersets "algosphera/scanner-api/internal/tickersets"
	"encoding/json"
)

func NewRouter(db *pgxpool.Pool, cat *tickersets.Catalog) (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
	}))

	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	r.Get("/ticker-sets", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"items": cat.ListSummaries(),
		})
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
