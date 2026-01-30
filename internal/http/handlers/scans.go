package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"net/http/httputil"

	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateScanRequest struct {
	JobType string         `json:"job_type"` // "calendar"
	Params  map[string]any `json:"params"`
}

type CreateScanResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ScansHandler struct {
	DB *pgxpool.Pool
}

type RunsHandler struct {
	Proxy *httputil.ReverseProxy
}

func NewScansHandler(db *pgxpool.Pool) *ScansHandler {
	return &ScansHandler{DB: db}
}

func NewRunsHandler(readerBaseURL string) (*RunsHandler, error) {
	u, err := url.Parse(readerBaseURL)
	if err != nil {
		return nil, err
	}

	p := httputil.NewSingleHostReverseProxy(u)

	// Ensure Host header matches the target (helps in some envs)
	orig := p.Director
	p.Director = func(r *http.Request) {
		orig(r)
		r.Host = u.Host
	}

	return &RunsHandler{Proxy: p}, nil
}

type GetScanResponse struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	JobType    string     `json:"job_type"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	ResultRows *int            `json:"result_rows,omitempty"`
	Summary    json.RawMessage `json:"summary_json,omitempty"`
	ErrorText  *string         `json:"error_text,omitempty"`
}

func (h *ScansHandler) CreateScan(w http.ResponseWriter, r *http.Request) {
	var req CreateScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if req.JobType == "" {
		req.JobType = "calendar"
	}

	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		http.Error(w, "invalid params", http.StatusBadRequest)
		return
	}

	var id string
	err = h.DB.QueryRow(
		context.Background(),
		`insert into scan_jobs (id, status, job_type, params_json)
		 values (gen_random_uuid(), 'queued', $1, $2::jsonb)
		 returning id`,
		req.JobType, string(paramsBytes),
	).Scan(&id)

	if err != nil {
		http.Error(w, "db insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateScanResponse{ID: id, Status: "queued"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ScansHandler) GetScan(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	var resp GetScanResponse
	var startedAt *time.Time
	var finishedAt *time.Time
	var resultRows *int
	var summary json.RawMessage
	var errorText *string

	err := h.DB.QueryRow(
		r.Context(),
		`select
			id::text,
			status,
			job_type,
			created_at,
			started_at,
			finished_at,
			result_rows,
			summary_json,
			error_text
		from scan_jobs
		where id = $1`,
		jobID,
	).Scan(
		&resp.ID,
		&resp.Status,
		&resp.JobType,
		&resp.CreatedAt,
		&startedAt,
		&finishedAt,
		&resultRows,
		&summary,
		&errorText,
	)

	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp.StartedAt = startedAt
	resp.FinishedAt = finishedAt
	resp.ResultRows = resultRows
	resp.Summary = summary
	resp.ErrorText = errorText

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /scans  ->  GET {READER_BASE_URL}/runs?strategy=calendar&limit=20

func (h *RunsHandler) GetRuns(w http.ResponseWriter, r *http.Request) {
	strategy := r.URL.Query().Get("strategy")
	limit := r.URL.Query().Get("limit")

	r.URL.Path = "/scan_jobs"
	q := r.URL.Query()
	q.Set("strategy", strategy)
	q.Set("limit", limit)
	r.URL.RawQuery = q.Encode()

	h.Proxy.ServeHTTP(w, r)
}
