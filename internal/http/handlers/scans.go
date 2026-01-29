package handlers

import (
	"context"
	"encoding/json"
	"net/http"

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

func NewScansHandler(db *pgxpool.Pool) *ScansHandler {
	return &ScansHandler{DB: db}
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
