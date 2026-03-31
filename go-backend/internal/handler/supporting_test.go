package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func supportingRouter(t *testing.T, db *gorm.DB) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux, jwtSvc
}

func TestListComments(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO comments (id, user_id, commentable_id, commentable_type, comment, created_at, updated_at) VALUES (1, 1, 1, 'Account', 'Great company!', ?, ?)", n, n)
	db.Exec("INSERT INTO comments (id, user_id, commentable_id, commentable_type, comment, created_at, updated_at) VALUES (2, 1, 1, 'Account', 'Follow up needed', ?, ?)", n, n)
	db.Exec("INSERT INTO comments (id, user_id, commentable_id, commentable_type, comment, created_at, updated_at) VALUES (3, 1, 2, 'Account', 'Different account', ?, ?)", n, n)

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/comments", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var comments []model.Comment
	if err := json.NewDecoder(rec.Body).Decode(&comments); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(comments) != 2 {
		t.Errorf("expected 2 comments for account 1, got %d", len(comments))
	}
}

func TestListAddresses(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO addresses (id, street1, city, state, country, address_type, addressable_id, addressable_type, created_at, updated_at) VALUES (1, '123 Main St', 'Anytown', 'CA', 'US', 'billing', 1, 'Account', ?, ?)", n, n)
	db.Exec("INSERT INTO addresses (id, street1, city, state, country, address_type, addressable_id, addressable_type, created_at, updated_at, deleted_at) VALUES (2, '456 Elm St', 'Other', 'NY', 'US', 'shipping', 1, 'Account', ?, ?, ?)", n, n, n)

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/addresses", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var addresses []model.Address
	if err := json.NewDecoder(rec.Body).Decode(&addresses); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Only 1 (soft-deleted excluded)
	if len(addresses) != 1 {
		t.Errorf("expected 1 address (soft-deleted excluded), got %d", len(addresses))
	}
}

func TestListEntityTags(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO tags (id, name, taggings_count) VALUES (1, 'vip', 1)")
	db.Exec("INSERT INTO tags (id, name, taggings_count) VALUES (2, 'prospect', 1)")
	db.Exec("INSERT INTO taggings (id, tag_id, taggable_id, taggable_type, context, created_at) VALUES (1, 1, 1, 'Account', 'tags', ?)", n)
	db.Exec("INSERT INTO taggings (id, tag_id, taggable_id, taggable_type, context, created_at) VALUES (2, 2, 2, 'Account', 'tags', ?)", n)

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/tags", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var tags []model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tags); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag for account 1, got %d", len(tags))
	}
	if len(tags) > 0 && tags[0].Name != "vip" {
		t.Errorf("expected tag 'vip', got %q", tags[0].Name)
	}
}

func TestListAllTags(t *testing.T) {
	db := testDB(t)
	db.Exec("INSERT INTO tags (id, name, taggings_count) VALUES (1, 'beta', 3)")
	db.Exec("INSERT INTO tags (id, name, taggings_count) VALUES (2, 'alpha', 5)")

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "user", false)
	req := httptest.NewRequest("GET", "/api/v1/tags", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var tags []model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tags); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	// Should be ordered by name: alpha, beta
	if len(tags) == 2 && tags[0].Name != "alpha" {
		t.Errorf("expected first tag 'alpha', got %q", tags[0].Name)
	}
}

func TestListEntityVersions(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, whodunnit, created_at) VALUES (1, 'Account', 1, 'create', '1', ?)", n)
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, whodunnit, created_at) VALUES (2, 'Account', 1, 'update', '1', ?)", n)
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, whodunnit, created_at) VALUES (3, 'Account', 2, 'create', '1', ?)", n)

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/versions", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var versions []model.Version
	if err := json.NewDecoder(rec.Body).Decode(&versions); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions for account 1, got %d", len(versions))
	}
}

func TestListRecentActivity(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, created_at) VALUES (1, 'Account', 1, 'create', ?)", n)
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, created_at) VALUES (2, 'Contact', 1, 'update', ?)", n)
	db.Exec("INSERT INTO versions (id, item_type, item_id, event, created_at) VALUES (3, 'Comment', 1, 'create', ?)", n) // Not a tracked asset

	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "user", false)
	req := httptest.NewRequest("GET", "/api/v1/activity", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var versions []model.Version
	if err := json.NewDecoder(rec.Body).Decode(&versions); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Only Account and Contact are tracked assets, not Comment
	if len(versions) != 2 {
		t.Errorf("expected 2 activity entries (tracked assets only), got %d", len(versions))
	}
}

func TestListUsers_AdminOnly(t *testing.T) {
	db := testDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at) VALUES (1, 'admin', 'admin@test.com', 'x', 'y', true, 'Admin', 'User', ?, ?)", n, n)
	db.Exec("INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at) VALUES (2, 'demo', 'demo@test.com', 'x', 'y', false, 'Demo', 'User', ?, ?)", n, n)

	mux, jwtSvc := supportingRouter(t, db)

	// Admin can list users
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("admin: expected 200, got %d", rec.Code)
	}

	var users []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&users); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
	// Verify sensitive fields are stripped
	for _, u := range users {
		if _, ok := u["encrypted_password"]; ok {
			t.Error("encrypted_password should not be in response")
		}
		if _, ok := u["password_salt"]; ok {
			t.Error("password_salt should not be in response")
		}
	}

	// Non-admin gets 403
	tok2, _ := jwtSvc.GenerateToken(2, "demo", false)
	req2 := httptest.NewRequest("GET", "/api/v1/users", nil)
	req2.Header.Set("Authorization", "Bearer "+tok2)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != 403 {
		t.Errorf("non-admin: expected 403, got %d", rec2.Code)
	}
}

func TestInvalidEntity_Returns400(t *testing.T) {
	db := testDB(t)
	mux, jwtSvc := supportingRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/invalid/1/comments", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid entity, got %d", rec.Code)
	}
}
