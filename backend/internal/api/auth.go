package api

import (
	"net/http"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}

	email := r.FormValue("email")
	if email == "" {
		email = r.URL.Query().Get("email")
	}
	if email == "" {
		s.writeError(w, 400, "email is required")
		return
	}

	code, err := auth.GenerateMagicCode()
	if err != nil {
		s.writeError(w, 500, "failed to generate code")
		return
	}

	if err := s.store.CreateMagicLink(r.Context(), email, code, time.Now().Add(15*time.Minute)); err != nil {
		s.writeError(w, 500, "failed to create magic link")
		return
	}

	writeJSON(w, 200, envelope{Data: map[string]string{
		"message": "magic link sent (check console for dev mode)",
		"code":    code,
	}})
}

func (s *Server) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		s.writeError(w, 400, "code is required")
		return
	}

	email, err := s.store.ConsumeMagicLink(r.Context(), code)
	if err != nil {
		s.writeError(w, 500, "failed to verify code")
		return
	}
	if email == nil {
		s.writeError(w, 401, "invalid or expired code")
		return
	}

	user, err := s.store.CreateUser(r.Context(), *email, *email)
	if err != nil {
		s.writeError(w, 500, "failed to create user")
		return
	}

	token, tokenHash, err := auth.GenerateToken()
	if err != nil {
		s.writeError(w, 500, "failed to generate session")
		return
	}

	if _, err := s.store.CreateSession(r.Context(), user.ID, tokenHash, time.Now().Add(30*24*time.Hour)); err != nil {
		s.writeError(w, 500, "failed to create session")
		return
	}

	writeJSON(w, 200, envelope{Data: map[string]interface{}{
		"user":  user,
		"token": token,
	}})
}
