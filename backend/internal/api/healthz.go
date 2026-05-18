package api

import "net/http"

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, envelope{Data: map[string]string{"status": "ok"}})
}
