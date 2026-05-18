package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/yasserrmd/sunpath/backend/internal/queue"
)

func (s *Server) handleHorizonJob(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		s.enqueueHorizonJob(w, r)
		return
	}
	if r.Method == "GET" {
		s.getHorizonJobResult(w, r)
		return
	}
	writeError(w, 405, "method not allowed")
}

func (s *Server) enqueueHorizonJob(w http.ResponseWriter, r *http.Request) {
	if s.jobQueue == nil {
		s.writeError(w, 503, "job queue not configured")
		return
	}

	lat, lng, h, useDSM, err := parseHorizonParams(r)
	if err != nil {
		s.writeError(w, 400, err.Error())
		return
	}

	jobID := uuid.New().String()
	job := &queue.HorizonJob{
		ID:     jobID,
		Lat:    lat,
		Lng:    lng,
		Height: h,
		UseDSM: useDSM,
	}

	if err := s.jobQueue.Enqueue(r.Context(), job); err != nil {
		s.writeError(w, 500, "failed to enqueue job")
		return
	}

	writeJSON(w, 202, envelope{Data: map[string]string{
		"job_id": jobID,
	}})
}

func (s *Server) getHorizonJobResult(w http.ResponseWriter, r *http.Request) {
	if s.jobQueue == nil {
		s.writeError(w, 503, "job queue not configured")
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/horizon/job/")
	if jobID == "" || jobID == r.URL.Path {
		s.writeError(w, 400, "job ID required")
		return
	}

	result, err := s.jobQueue.GetResult(r.Context(), jobID)
	if err != nil {
		s.writeError(w, 500, "failed to get job result")
		return
	}
	if result == nil {
		writeJSON(w, 404, envelope{Data: map[string]string{
			"job_id": jobID,
			"status": "pending",
		}})
		return
	}

	if result.Error != "" {
		writeJSON(w, 200, envelope{Data: map[string]interface{}{
			"job_id": jobID,
			"status": "failed",
			"error":  result.Error,
		}})
		return
	}

	var profile horizonJSON
	if err := json.Unmarshal(result.Profile, &profile); err != nil {
		s.writeError(w, 500, "failed to decode result")
		return
	}

	writeJSON(w, 200, envelope{Data: map[string]interface{}{
		"job_id":  jobID,
		"status":  "completed",
		"profile": profile,
	}})
}

func parseHorizonParams(r *http.Request) (lat, lng, h float64, useDSM bool, err error) {
	q := r.URL.Query()

	if err := parseQueryFloat(q, "lat", &lat); err != nil {
		return 0, 0, 0, false, err
	}
	if err := parseQueryFloat(q, "lng", &lng); err != nil {
		return 0, 0, 0, false, err
	}
	if lat < -90 || lat > 90 {
		return 0, 0, 0, false, fmt.Errorf("lat out of range")
	}
	if lng < -180 || lng > 180 {
		return 0, 0, 0, false, fmt.Errorf("lng out of range")
	}

	hVal := q.Get("h")
	if hVal == "" {
		h = 1.5
	} else {
		if err := parseQueryFloat(q, "h", &h); err != nil {
			return 0, 0, 0, false, err
		}
	}

	useDSM = q.Get("dsm") == "true"
	return lat, lng, h, useDSM, nil
}

func parseQueryFloat(q map[string][]string, key string, out *float64) error {
	s, ok := q[key]
	if !ok || len(s) == 0 {
		return fmt.Errorf("%s is required", key)
	}
	v, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return fmt.Errorf("invalid %s", key)
	}
	*out = v
	return nil
}

type horizonJSON struct {
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	Height     float64   `json:"height"`
	Horizon    []float64 `json:"horizon"`
	Confidence float64   `json:"confidence"`
	BuildCount int       `json:"build_count"`
	EstCount   int       `json:"est_count"`
	DataHash   string    `json:"data_hash"`
	UseDSM     bool      `json:"use_dsm"`
}
