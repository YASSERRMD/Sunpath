package api

import "net/http"

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	stats := s.store.Stats()
	writeJSON(w, 200, envelope{Data: map[string]interface{}{
		"status": "ok",
		"store": map[string]int{
			"osm_extracts":     stats.OSMExtracts,
			"horizon_profiles": stats.HorizonProfiles,
		},
	}})
}
