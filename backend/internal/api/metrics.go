package api

import (
	"net/http"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

type metricsResponse struct {
	Cache  store.CacheStats       `json:"cache"`
	Errors map[string]int64       `json:"errors"`
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.writeError(w, 405, "method not allowed")
		return
	}

	cacheStats := s.store.Stats()
	errMap := make(map[string]int64, len(s.errorCounts))
	for code, counter := range s.errorCounts {
		errMap[code] = *counter
	}

	writeJSON(w, 200, envelope{Data: metricsResponse{
		Cache:  cacheStats,
		Errors: errMap,
	}})
}

func (s *Server) handleCacheEvict(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, 405, "method not allowed")
		return
	}

	removed, err := s.store.EvictOlderThan(evictionTTL)
	if err != nil {
		s.writeError(w, 500, "eviction failed")
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]int{"removed": removed}})
}
