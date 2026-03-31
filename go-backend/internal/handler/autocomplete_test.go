package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
)

func autocompleteRouter(t *testing.T) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Acme Corporation', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 1, 'Beta Industries', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (3, 1, 1, 'Acme Widgets', 'Public', ?, ?)", n, n)

	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 1, 'John', 'Smith', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (2, 1, 1, 'Jane', 'Doe', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (3, 1, 1, 'Bob', 'Johnson', 'Public', ?, ?)", n, n)

	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 1, 'Alice', 'Wonder', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (2, 1, 1, 'Charlie', 'Brown', 'Public', ?, ?)", n, n)

	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Spring Campaign', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 1, 'Fall Promo', 'Public', ?, ?)", n, n)

	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Big Deal', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 1, 'Small Deal', 'Public', ?, ?)", n, n)

	return mux, jwtSvc
}

func TestAutocomplete_Accounts(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=acme", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Should be sorted by name
	if results[0].Name != "Acme Corporation" {
		t.Errorf("expected first result 'Acme Corporation', got %q", results[0].Name)
	}
}

func TestAutocomplete_Contacts(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/contacts/autocomplete?q=john", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// "John Smith" matches on first_name, "Bob Johnson" matches on last_name
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %+v", len(results), results)
	}
}

func TestAutocomplete_Leads(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/leads/autocomplete?q=alice", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Alice Wonder" {
		t.Errorf("expected 'Alice Wonder', got %q", results[0].Name)
	}
}

func TestAutocomplete_Campaigns(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/campaigns/autocomplete?q=camp", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (Spring Campaign), got %d", len(results))
	}
}

func TestAutocomplete_Opportunities(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/opportunities/autocomplete?q=deal", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestAutocomplete_EmptyQuery(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty query, got %d", len(results))
	}
}

func TestAutocomplete_NoQuery(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for missing query, got %d", len(results))
	}
}

func TestAutocomplete_NoAuth(t *testing.T) {
	mux, _ := autocompleteRouter(t)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=acme", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAutocomplete_NoResults(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=zzzzz", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAutocomplete_CaseInsensitive(t *testing.T) {
	mux, jwtSvc := autocompleteRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=ACME", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var results []autocompleteItem
	if err := json.NewDecoder(rec.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for case-insensitive search, got %d", len(results))
	}
}

func TestAutocomplete_DeletedExcluded(t *testing.T) {
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 1, 'Acme Corp', 'Public', ?, ?)", n, n)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at, deleted_at) VALUES (2, 1, 1, 'Acme Deleted', 'Public', ?, ?, ?)", n, n, n)

	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/accounts/autocomplete?q=acme", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var results []autocompleteItem
	json.NewDecoder(rec.Body).Decode(&results)
	if len(results) != 1 {
		t.Errorf("expected 1 non-deleted result, got %d", len(results))
	}
}
