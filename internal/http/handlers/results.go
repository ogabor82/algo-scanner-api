package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type ResultsHandler struct {
	Proxy *httputil.ReverseProxy
}

func NewResultsHandler(readerBaseURL string) (*ResultsHandler, error) {
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

	return &ResultsHandler{Proxy: p}, nil
}

// GET /scans/{jobID}/results  ->  GET {READER_BASE_URL}/query?job_id={jobID}&...
func (h *ResultsHandler) GetResults(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")

	// Rewrite path to reader endpoint
	r.URL.Path = "/query"

	// Pass through query + inject job_id
	q := r.URL.Query()
	q.Set("job_id", jobID)
	r.URL.RawQuery = q.Encode()

	h.Proxy.ServeHTTP(w, r)
}
