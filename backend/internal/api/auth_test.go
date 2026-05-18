package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthLoginMissingEmail(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error == "" {
		t.Error("expected error message")
	}
}

func TestAuthLoginWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/auth/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAuthLoginSuccess(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	body := strings.NewReader("email=test@example.com")
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error != "" {
		t.Errorf("unexpected error: %s", env.Error)
	}
}

func TestAuthCallbackMissingCode(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/auth/callback", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAuthCallbackWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("POST", "/api/auth/callback", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestAuthCallbackInvalidCode(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("GET", "/api/auth/callback?code=bogus", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthCallbackSuccess(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	body := strings.NewReader("email=callback@example.com")
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("login failed: %d", rec.Code)
	}

	var loginResp struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}

	cbReq := httptest.NewRequest("GET", "/api/auth/callback?code="+loginResp.Data.Code, nil)
	cbRec := httptest.NewRecorder()
	handler.ServeHTTP(cbRec, cbReq)

	if cbRec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", cbRec.Code, cbRec.Body.String())
	}

	var cbEnv envelope
	if err := json.NewDecoder(cbRec.Body).Decode(&cbEnv); err != nil {
		t.Fatal(err)
	}
	if cbEnv.Error != "" {
		t.Errorf("unexpected error: %s", cbEnv.Error)
	}
}

func TestAuthCallbackCodeReuse(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	body := strings.NewReader("email=reuse@example.com")
	req := httptest.NewRequest("POST", "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var loginResp struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}

	cbReq1 := httptest.NewRequest("GET", "/api/auth/callback?code="+loginResp.Data.Code, nil)
	cbRec1 := httptest.NewRecorder()
	handler.ServeHTTP(cbRec1, cbReq1)
	if cbRec1.Code != 200 {
		t.Fatalf("first callback failed: %d", cbRec1.Code)
	}

	cbReq2 := httptest.NewRequest("GET", "/api/auth/callback?code="+loginResp.Data.Code, nil)
	cbRec2 := httptest.NewRecorder()
	handler.ServeHTTP(cbRec2, cbReq2)
	if cbRec2.Code != 401 {
		t.Errorf("expected 401 for reused code, got %d", cbRec2.Code)
	}
}

func TestRequireAuthMissingHeader(t *testing.T) {
	srv := newTestServer(t)
	middleware := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	middleware(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	srv := newTestServer(t)
	middleware := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rec := httptest.NewRecorder()
	middleware(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthValidToken(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	loginBody := strings.NewReader("email=validtoken@example.com")
	loginReq := httptest.NewRequest("POST", "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	var loginResp struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}

	cbReq := httptest.NewRequest("GET", "/api/auth/callback?code="+loginResp.Data.Code, nil)
	cbRec := httptest.NewRecorder()
	handler.ServeHTTP(cbRec, cbReq)

	var cbEnv struct {
		Data struct {
			User  interface{} `json:"user"`
			Token string      `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(cbRec.Body).Decode(&cbEnv); err != nil {
		t.Fatal(err)
	}

	called := false
	middleware := srv.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		u := getUser(r.Context())
		if u == nil {
			t.Error("expected user in context")
		}
		w.WriteHeader(200)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+cbEnv.Data.Token)
	rec := httptest.NewRecorder()
	middleware(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOptionalAuthNoHeader(t *testing.T) {
	srv := newTestServer(t)
	called := false
	middleware := srv.optionalAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		u := getUser(r.Context())
		if u != nil {
			t.Error("expected no user without auth header")
		}
		w.WriteHeader(200)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	middleware(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestOptionalAuthWithToken(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	loginBody := strings.NewReader("email=optionalauth@example.com")
	loginReq := httptest.NewRequest("POST", "/api/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	var loginResp struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}

	cbReq := httptest.NewRequest("GET", "/api/auth/callback?code="+loginResp.Data.Code, nil)
	cbRec := httptest.NewRecorder()
	handler.ServeHTTP(cbRec, cbReq)

	var cbEnv struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(cbRec.Body).Decode(&cbEnv); err != nil {
		t.Fatal(err)
	}

	called := false
	middleware := srv.optionalAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		u := getUser(r.Context())
		if u == nil {
			t.Error("expected user with valid token")
		}
		w.WriteHeader(200)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+cbEnv.Data.Token)
	rec := httptest.NewRecorder()
	middleware(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("handler was not called")
	}
}
