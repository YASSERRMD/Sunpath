package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBatchHorizonWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/horizon/batch", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestBatchHorizonEmptyBody(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/horizon/batch", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchHorizonTooManyPoints(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	points := make([]map[string]interface{}, 21)
	for i := 0; i < 21; i++ {
		points[i] = map[string]interface{}{"lat": 48.85, "lng": 2.35, "height": 1.5}
	}
	body, _ := json.Marshal(map[string]interface{}{"points": points})
	req := httptest.NewRequest("POST", "/api/horizon/batch", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchHorizonEmptyPoints(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	body := `{"points":[]}`
	req := httptest.NewRequest("POST", "/api/horizon/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestBatchHorizonInvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/horizon/batch", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
