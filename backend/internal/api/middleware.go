package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/auth"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func getUser(ctx context.Context) *store.UserRecord {
	u, _ := ctx.Value(userContextKey).(*store.UserRecord)
	return u
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			s.writeError(w, 401, "authorization header required")
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			s.writeError(w, 401, "bearer token required")
			return
		}

		tokenHash := auth.HashToken(token)
		session, err := s.store.GetSessionByTokenHash(r.Context(), tokenHash)
		if err != nil {
			s.writeError(w, 500, "failed to validate session")
			return
		}
		if session == nil {
			s.writeError(w, 401, "invalid session")
			return
		}
		if time.Now().After(session.ExpiresAt) {
			s.writeError(w, 401, "session expired")
			return
		}

		user, err := s.store.GetUserByID(r.Context(), session.UserID)
		if err != nil || user == nil {
			s.writeError(w, 401, "invalid session")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			next(w, r)
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if token == header {
			next(w, r)
			return
		}

		tokenHash := auth.HashToken(token)
		session, err := s.store.GetSessionByTokenHash(r.Context(), tokenHash)
		if err != nil || session == nil {
			next(w, r)
			return
		}
		if time.Now().After(session.ExpiresAt) {
			next(w, r)
			return
		}

		user, err := s.store.GetUserByID(r.Context(), session.UserID)
		if err != nil || user == nil {
			next(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}
