package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getAuthToken(t *testing.T, srv *Server, handler http.Handler, email string) string {
	t.Helper()
	loginBody := strings.NewReader("email=" + email)
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
		t.Fatalf("login decode: %v", err)
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
		t.Fatalf("callback decode: %v", err)
	}
	return cbEnv.Data.Token
}

func TestProjectsRequireAuth(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	tests := []struct {
		name string
		url  string
		meth string
	}{
		{"list", "/api/projects", "GET"},
		{"create", "/api/projects", "POST"},
		{"get", "/api/projects/1", "GET"},
		{"update", "/api/projects/1", "PUT"},
		{"delete", "/api/projects/1", "DELETE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.meth, tt.url, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != 401 {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestProjectsCreateAndList(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()
	token := getAuthToken(t, srv, handler, "test-projects@example.com")

	body := `{"name":"My Test","lat":48.85,"lng":2.35,"height":10,"use_dsm":false}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var env envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if env.Error != "" {
		t.Fatalf("unexpected error: %s", env.Error)
	}

	// List
	listReq := httptest.NewRequest("GET", "/api/projects", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)

	if listRec.Code != 200 {
		t.Fatalf("expected 200, got %d", listRec.Code)
	}

	var listEnv envelope
	if err := json.NewDecoder(listRec.Body).Decode(&listEnv); err != nil {
		t.Fatal(err)
	}
	projects, ok := listEnv.Data.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", listEnv.Data)
	}
	if len(projects) < 1 {
		t.Error("expected at least 1 project")
	}
}

func TestProjectsCreateRequiresName(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()
	token := getAuthToken(t, srv, handler, "test-namereq@example.com")

	body := `{"lat":48.85,"lng":2.35,"height":1.5,"use_dsm":false}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProjectsGetUpdateDelete(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()
	token := getAuthToken(t, srv, handler, "test-crud@example.com")

	createBody := `{"name":"CRUD Test","lat":48.86,"lng":2.36,"height":5,"use_dsm":true}`
	req := httptest.NewRequest("POST", "/api/projects", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var createResp struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("create decode: %v", err)
	}
	projectID := createResp.Data.ID

	actualURL := fmt.Sprintf("/api/projects/%d", projectID)
	getReq2 := httptest.NewRequest("GET", actualURL, nil)
	getReq2.Header.Set("Authorization", "Bearer "+token)
	getRec2 := httptest.NewRecorder()
	handler.ServeHTTP(getRec2, getReq2)
	if getRec2.Code != 200 {
		t.Errorf("expected 200, got %d", getRec2.Code)
	}

	// Update
	updateBody := `{"name":"Updated","lat":48.87,"lng":2.37,"height":20,"use_dsm":false}`
	putReq := httptest.NewRequest("PUT", actualURL, strings.NewReader(updateBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putRec := httptest.NewRecorder()
	handler.ServeHTTP(putRec, putReq)
	if putRec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", putRec.Code, putRec.Body.String())
	}

	// Delete
	delReq := httptest.NewRequest("DELETE", actualURL, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	handler.ServeHTTP(delRec, delReq)
	if delRec.Code != 200 {
		t.Errorf("expected 200, got %d", delRec.Code)
	}

	// Verify deleted
	getDelReq := httptest.NewRequest("GET", actualURL, nil)
	getDelReq.Header.Set("Authorization", "Bearer "+token)
	getDelRec := httptest.NewRecorder()
	handler.ServeHTTP(getDelRec, getDelReq)
	if getDelRec.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", getDelRec.Code)
	}
}

func TestProjectsWrongMethod(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()

	req := httptest.NewRequest("PUT", "/api/projects", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestProjectsDeleteNotFound(t *testing.T) {
	srv := newTestServer(t)
	handler := srv.Routes()
	token := getAuthToken(t, srv, handler, "test-delete@example.com")

	delReq := httptest.NewRequest("DELETE", "/api/projects/99999", nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delRec := httptest.NewRecorder()
	handler.ServeHTTP(delRec, delReq)
	if delRec.Code != 200 {
		t.Errorf("expected 200 (delete succeeds even if not found), got %d", delRec.Code)
	}
}
