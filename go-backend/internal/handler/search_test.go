package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
)

func searchRouter(t *testing.T) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	// Seed data
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Acme Corporation', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 1, 'Beta Industries', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 1, 'John', 'Acme', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 1, 'Jane', 'Doe', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Acme Deal', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Acme Campaign', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, created_at, updated_at) VALUES (1, 1, 1, 'Call Acme', ?, ?)", n, n)

	return mux, jwtSvc
}

func TestSearch_AllEntities(t *testing.T) {
	mux, jwtSvc := searchRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/search?q=Acme", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result SearchResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Query != "Acme" {
		t.Errorf("expected query 'Acme', got %q", result.Query)
	}
	if len(result.Accounts) != 1 {
		t.Errorf("expected 1 account match, got %d", len(result.Accounts))
	}
	if len(result.Contacts) != 1 {
		t.Errorf("expected 1 contact match (last_name=Acme), got %d", len(result.Contacts))
	}
	if len(result.Opportunities) != 1 {
		t.Errorf("expected 1 opportunity match, got %d", len(result.Opportunities))
	}
	if len(result.Campaigns) != 1 {
		t.Errorf("expected 1 campaign match, got %d", len(result.Campaigns))
	}
	if len(result.Tasks) != 1 {
		t.Errorf("expected 1 task match, got %d", len(result.Tasks))
	}
	if result.TotalCount != 5 {
		t.Errorf("expected total_count 5, got %d", result.TotalCount)
	}
}

func TestSearch_EntityFilter(t *testing.T) {
	mux, jwtSvc := searchRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/search?q=Acme&entity=accounts", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result SearchResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(result.Accounts))
	}
	// Other entity lists should be nil (not searched)
	if result.Contacts != nil {
		t.Errorf("expected nil contacts when filtering by accounts, got %d", len(result.Contacts))
	}
}

func TestSearch_NoResults(t *testing.T) {
	mux, jwtSvc := searchRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/search?q=nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result SearchResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.TotalCount != 0 {
		t.Errorf("expected 0 results, got %d", result.TotalCount)
	}
}

func TestSearch_MissingQuery(t *testing.T) {
	mux, jwtSvc := searchRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/search", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400 for missing q, got %d", rec.Code)
	}
}

func TestSearch_NoAuth(t *testing.T) {
	mux, _ := searchRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/search?q=test", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
