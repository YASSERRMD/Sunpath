package api

import (
	"net/http"
)

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	stats := s.store.Stats()
	writeJSON(w, 200, envelope{Data: stats})
}

func (s *Server) handleCacheEvict(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	removed, err := s.store.EvictOlderThan(evictionTTL)
	if err != nil {
		writeError(w, 500, "eviction failed")
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]int{"removed": removed}})
}
