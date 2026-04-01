package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func savedSearchRouter(t *testing.T) (*http.ServeMux, func(string) string) {
	t.Helper()
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	// Seed users for FK constraint
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at)
		VALUES (1, 'admin', 'admin@test.com', 'x', 'y', true, 'A', 'D', ?, ?)`, now, now)
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at)
		VALUES (5, 'user', 'user@test.com', 'x', 'y', false, 'U', 'S', ?, ?)`, now, now)

	makeToken := func(role string) string {
		isAdmin := role == "admin"
		userID := int64(1)
		if !isAdmin {
			userID = 5
		}
		tok, _ := jwtSvc.GenerateToken(userID, role, isAdmin)
		return tok
	}
	return mux, makeToken
}

func TestSavedSearches_CRUD(t *testing.T) {
	mux, makeToken := savedSearchRouter(t)
	tok := makeToken("admin")

	// List — empty
	rec := doRequest(mux, "GET", "/api/v1/saved_searches", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("list: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var searches []model.SavedSearch
	json.NewDecoder(rec.Body).Decode(&searches)
	if len(searches) != 0 {
		t.Errorf("expected 0 saved searches, got %d", len(searches))
	}

	// Create
	rec = doRequest(mux, "POST", "/api/v1/saved_searches", tok, map[string]interface{}{
		"name":    "Active Leads",
		"entity":  "leads",
		"filters": map[string]string{"status_eq": "new"},
	})
	if rec.Code != 201 {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created model.SavedSearch
	json.NewDecoder(rec.Body).Decode(&created)
	if created.Name != "Active Leads" {
		t.Errorf("expected name 'Active Leads', got %q", created.Name)
	}
	if created.Entity != "leads" {
		t.Errorf("expected entity 'leads', got %q", created.Entity)
	}

	// List — one
	rec = doRequest(mux, "GET", "/api/v1/saved_searches", tok, nil)
	json.NewDecoder(rec.Body).Decode(&searches)
	if len(searches) != 1 {
		t.Fatalf("expected 1 saved search, got %d", len(searches))
	}

	// Update
	rec = doRequest(mux, "PUT", "/api/v1/saved_searches/1", tok, map[string]interface{}{
		"name": "Hot Leads",
	})
	if rec.Code != 200 {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updated model.SavedSearch
	json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Name != "Hot Leads" {
		t.Errorf("expected updated name 'Hot Leads', got %q", updated.Name)
	}

	// Delete
	rec = doRequest(mux, "DELETE", "/api/v1/saved_searches/1", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("delete: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// List — empty again
	rec = doRequest(mux, "GET", "/api/v1/saved_searches", tok, nil)
	json.NewDecoder(rec.Body).Decode(&searches)
	if len(searches) != 0 {
		t.Errorf("expected 0 after delete, got %d", len(searches))
	}
}

func TestSavedSearches_CreateValidation(t *testing.T) {
	mux, makeToken := savedSearchRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/saved_searches", tok, map[string]interface{}{
		"name": "Missing Entity",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for missing entity, got %d", rec.Code)
	}
}

func TestSavedSearches_UserIsolation(t *testing.T) {
	mux, makeToken := savedSearchRouter(t)
	adminTok := makeToken("admin")
	userTok := makeToken("user")

	// Admin creates a saved search
	doRequest(mux, "POST", "/api/v1/saved_searches", adminTok, map[string]interface{}{
		"name": "Admin Search", "entity": "accounts", "filters": map[string]string{},
	})

	// Non-admin user should not see admin's saved searches
	rec := doRequest(mux, "GET", "/api/v1/saved_searches", userTok, nil)
	var searches []model.SavedSearch
	json.NewDecoder(rec.Body).Decode(&searches)
	if len(searches) != 0 {
		t.Errorf("user should not see admin's saved searches, got %d", len(searches))
	}

	// Non-admin user cannot delete admin's saved search
	rec = doRequest(mux, "DELETE", "/api/v1/saved_searches/1", userTok, nil)
	if rec.Code != 404 {
		t.Errorf("expected 404 when user tries to delete admin's search, got %d", rec.Code)
	}
}
