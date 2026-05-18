package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	st, err := store.NewPostgresStore(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() {
		st.Close()
	})

	return NewServer(st, "https://overpass-api.de/api/interpreter")
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error != "" {
		t.Errorf("unexpected error: %s", env.Error)
	}
}

func TestHorizonMissingParams(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	tests := []struct {
		name string
		url  string
		code int
	}{
		{"missing all", "/api/horizon", 400},
		{"missing lng", "/api/horizon?lat=48.85", 400},
		{"invalid lat", "/api/horizon?lat=abc&lng=2.35", 400},
		{"lat out of range", "/api/horizon?lat=100&lng=2.35", 400},
		{"lng out of range", "/api/horizon?lat=48.85&lng=200", 400},
		{"invalid h", "/api/horizon?lat=48.85&lng=2.35&h=abc", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.code {
				t.Errorf("expected %d, got %d", tt.code, rec.Code)
			}

			var env envelope
			if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
				t.Fatal(err)
			}
			if env.Error == "" {
				t.Error("expected error message")
			}
		})
	}
}

func TestHorizonWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/horizon", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestGeocodeMissingQuery(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/geocode", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGeocodeWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/geocode", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestCorsHeaders(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("OPTIONS", "/api/horizon", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 204 {
		t.Errorf("expected 204 for OPTIONS, got %d", rec.Code)
	}
	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Errorf("expected CORS origin, got %s", origin)
	}
}

func TestGridMissingParams(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	tests := []struct {
		name string
		url  string
		code int
	}{
		{"missing all", "/api/grid", 400},
		{"missing lng1", "/api/grid?lat1=48.85", 400},
		{"missing lat2", "/api/grid?lat1=48.85&lng1=2.35", 400},
		{"invalid lat1", "/api/grid?lat1=abc&lng1=2.35&lat2=48.86&lng2=2.36", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.code {
				t.Errorf("expected %d, got %d", tt.code, rec.Code)
			}
		})
	}
}

func TestGridWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/grid?lat1=48.85&lng1=2.35&lat2=48.86&lng2=2.36", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestJSONErrorEnvelope(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/horizon", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := strings.TrimSpace(rec.Body.String())
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(body, "}") {
		t.Errorf("expected JSON object, got: %s", body)
	}
	if !strings.Contains(body, "error") {
		t.Errorf("expected error field in response, got: %s", body)
	}
}
