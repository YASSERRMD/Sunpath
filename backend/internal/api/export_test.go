package api

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExportCSVMissingParams(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	tests := []struct {
		name string
		url  string
		code int
	}{
		{"missing all", "/api/export/csv", 400},
		{"missing lng", "/api/export/csv?lat=48.85", 400},
		{"invalid lat", "/api/export/csv?lat=abc&lng=2.35", 400},
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

func TestExportCSVWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/export/csv", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestExportPDFMissingParams(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/export/pdf", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestExportPDFWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("DELETE", "/api/export/pdf", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestExportCSVContentType(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/export/csv?lat=48.85&lng=2.35&h=1.5", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == 200 {
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/csv") {
			t.Errorf("expected CSV content type, got %s", ct)
		}
		disp := rec.Header().Get("Content-Disposition")
		if !strings.Contains(disp, "attachment") {
			t.Errorf("expected attachment disposition, got %s", disp)
		}
	}
}
