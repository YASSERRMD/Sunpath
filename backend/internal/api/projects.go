package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.listProjects(w, r)
	case "POST":
		s.createProject(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.writeError(w, 400, "invalid project id")
		return
	}

	switch r.Method {
	case "GET":
		s.getProject(w, r, id)
	case "PUT":
		s.updateProject(w, r, id)
	case "DELETE":
		s.deleteProject(w, r, id)
	default:
		writeError(w, 405, "method not allowed")
	}
}

type projectInput struct {
	Name   string  `json:"name"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Height float64 `json:"height"`
	UseDSM bool    `json:"use_dsm"`
}

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	user := getUser(r.Context())
	if user == nil {
		s.writeError(w, 401, "authentication required")
		return
	}

	projects, err := s.store.ListProjects(r.Context(), user.ID)
	if err != nil {
		s.writeError(w, 500, "failed to list projects")
		return
	}
	if projects == nil {
		projects = []store.ProjectRecord{}
	}
	writeJSON(w, 200, envelope{Data: projects})
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	user := getUser(r.Context())
	if user == nil {
		s.writeError(w, 401, "authentication required")
		return
	}

	var input projectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		s.writeError(w, 400, "invalid JSON body")
		return
	}
	if input.Name == "" {
		s.writeError(w, 400, "name is required")
		return
	}

	project, err := s.store.CreateProject(r.Context(), user.ID, input.Name, input.Lat, input.Lng, input.Height, input.UseDSM)
	if err != nil {
		s.writeError(w, 500, "failed to create project")
		return
	}
	writeJSON(w, 201, envelope{Data: project})
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request, id int64) {
	user := getUser(r.Context())
	if user == nil {
		s.writeError(w, 401, "authentication required")
		return
	}

	project, err := s.store.GetProject(r.Context(), id, user.ID)
	if err != nil {
		s.writeError(w, 500, "failed to get project")
		return
	}
	if project == nil {
		s.writeError(w, 404, "project not found")
		return
	}
	writeJSON(w, 200, envelope{Data: project})
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request, id int64) {
	user := getUser(r.Context())
	if user == nil {
		s.writeError(w, 401, "authentication required")
		return
	}

	var input projectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		s.writeError(w, 400, "invalid JSON body")
		return
	}

	project, err := s.store.UpdateProject(r.Context(), id, user.ID, input.Name, input.Lat, input.Lng, input.Height, input.UseDSM)
	if err != nil {
		s.writeError(w, 500, "failed to update project")
		return
	}
	if project == nil {
		s.writeError(w, 404, "project not found")
		return
	}
	writeJSON(w, 200, envelope{Data: project})
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request, id int64) {
	user := getUser(r.Context())
	if user == nil {
		s.writeError(w, 401, "authentication required")
		return
	}

	if err := s.store.DeleteProject(r.Context(), id, user.ID); err != nil {
		s.writeError(w, 500, "failed to delete project")
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]string{"status": "deleted"}})
}
