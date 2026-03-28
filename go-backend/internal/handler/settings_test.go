package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
)

func setupSettingsDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.Exec(`CREATE TABLE settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(32) NOT NULL DEFAULT '',
		value TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`)
	return db
}

func seedSettings(db *gorm.DB) {
	// Insert settings with YAML-serialized values matching Rails format
	db.Exec(`INSERT INTO settings (name, value) VALUES ('host', '--- www.example.com
...')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('locale', '--- en-US
...')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('per_user_locale', '--- false
...')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('default_access', '--- Public
...')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('user_signup', '--- :not_allowed
')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('account_category', '---
- :affiliate
- :competitor
- :customer
- :partner
')`)
	db.Exec(`INSERT INTO settings (name, value) VALUES ('smtp', '---
:address: smtp.gmail.com
:port: "587"
:user_name: user@example.com
')`)
}

func settingsRequest(t *testing.T, handler http.HandlerFunc, method, path string, body interface{}, claims *auth.Claims) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if claims != nil {
		ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
		req = req.WithContext(ctx)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func settingsRouterRequest(t *testing.T, r *chi.Mux, method, path string, body interface{}, claims *auth.Claims) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if claims != nil {
		ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
		req = req.WithContext(ctx)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestGetSettings(t *testing.T) {
	db := setupSettingsDB(t)
	seedSettings(db)
	h := NewSettingsHandler(db)
	adminClaims := &auth.Claims{UserID: 1, Admin: true}

	rr := settingsRequest(t, h.GetSettings, "GET", "/admin/settings", nil, adminClaims)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	if result["host"] != "www.example.com" {
		t.Errorf("expected host=www.example.com, got %v", result["host"])
	}
	if result["locale"] != "en-US" {
		t.Errorf("expected locale=en-US, got %v", result["locale"])
	}
	if result["per_user_locale"] != false {
		t.Errorf("expected per_user_locale=false, got %v", result["per_user_locale"])
	}
	if result["default_access"] != "Public" {
		t.Errorf("expected default_access=Public, got %v", result["default_access"])
	}

	// Check array deserialization
	cats, ok := result["account_category"].([]interface{})
	if !ok {
		t.Fatalf("expected account_category to be array, got %T: %v", result["account_category"], result["account_category"])
	}
	if len(cats) != 4 {
		t.Errorf("expected 4 account categories, got %d", len(cats))
	}

	// Check hash deserialization
	smtp, ok := result["smtp"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected smtp to be map, got %T", result["smtp"])
	}
	if smtp["address"] != "smtp.gmail.com" {
		t.Errorf("expected smtp.address=smtp.gmail.com, got %v", smtp["address"])
	}
}

func TestGetSettings_NonAdmin(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)
	userClaims := &auth.Claims{UserID: 2, Admin: false}

	rr := settingsRequest(t, h.GetSettings, "GET", "/admin/settings", nil, userClaims)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestGetSettings_NoAuth(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)

	rr := settingsRequest(t, h.GetSettings, "GET", "/admin/settings", nil, nil)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUpdateSettings(t *testing.T) {
	db := setupSettingsDB(t)
	seedSettings(db)
	h := NewSettingsHandler(db)
	adminClaims := &auth.Claims{UserID: 1, Admin: true}

	r := chi.NewRouter()
	r.Get("/admin/settings", h.GetSettings)
	r.Put("/admin/settings", h.UpdateSettings)

	body := map[string]interface{}{
		"host":            "crm.example.com",
		"per_user_locale": true,
		"new_setting":     "hello",
	}

	rr := settingsRouterRequest(t, r, "PUT", "/admin/settings", body, adminClaims)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	if result["host"] != "crm.example.com" {
		t.Errorf("expected host=crm.example.com, got %v", result["host"])
	}
	if result["per_user_locale"] != true {
		t.Errorf("expected per_user_locale=true, got %v", result["per_user_locale"])
	}
	if result["new_setting"] != "hello" {
		t.Errorf("expected new_setting=hello, got %v", result["new_setting"])
	}
	// Verify unchanged settings persist
	if result["locale"] != "en-US" {
		t.Errorf("expected locale=en-US unchanged, got %v", result["locale"])
	}
}

func TestUpdateSettings_NonAdmin(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)
	userClaims := &auth.Claims{UserID: 2, Admin: false}

	body := map[string]interface{}{"host": "evil.com"}
	rr := settingsRequest(t, h.UpdateSettings, "PUT", "/admin/settings", body, userClaims)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestUpdateSettings_Array(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)
	adminClaims := &auth.Claims{UserID: 1, Admin: true}

	r := chi.NewRouter()
	r.Get("/admin/settings", h.GetSettings)
	r.Put("/admin/settings", h.UpdateSettings)

	body := map[string]interface{}{
		"lead_status": []string{"new", "contacted", "converted", "rejected", "recycled"},
	}

	rr := settingsRouterRequest(t, r, "PUT", "/admin/settings", body, adminClaims)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	statuses, ok := result["lead_status"].([]interface{})
	if !ok {
		t.Fatalf("expected lead_status to be array, got %T", result["lead_status"])
	}
	if len(statuses) != 5 {
		t.Errorf("expected 5 statuses, got %d", len(statuses))
	}
}

func TestUpdateSettings_Hash(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)
	adminClaims := &auth.Claims{UserID: 1, Admin: true}

	r := chi.NewRouter()
	r.Get("/admin/settings", h.GetSettings)
	r.Put("/admin/settings", h.UpdateSettings)

	body := map[string]interface{}{
		"smtp": map[string]interface{}{
			"address":  "smtp.newhost.com",
			"port":     "465",
			"user_name": "admin@newhost.com",
		},
	}

	rr := settingsRouterRequest(t, r, "PUT", "/admin/settings", body, adminClaims)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	smtp, ok := result["smtp"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected smtp to be map, got %T", result["smtp"])
	}
	if smtp["address"] != "smtp.newhost.com" {
		t.Errorf("expected smtp.address=smtp.newhost.com, got %v", smtp["address"])
	}
}

func TestDeserializeYAMLValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"empty string", "", nil},
		{"simple string", "--- hello\n...\n", "hello"},
		{"boolean true", "--- true\n...\n", true},
		{"boolean false", "--- false\n...\n", false},
		{"integer", "--- 42\n", 42},
		{"symbol", "--- :not_allowed\n", ":not_allowed"},
		{"array of symbols", "---\n- :one\n- :two\n- :three\n", []interface{}{":one", ":two", ":three"}},
		{"array of strings", "---\n- hello\n- world\n", []interface{}{"hello", "world"}},
		{"hash", "---\n:address: smtp.gmail.com\n:port: '587'\n", map[string]interface{}{"address": "smtp.gmail.com", "port": "587"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deserializeYAMLValue(tt.input)
			gotJSON, _ := json.Marshal(got)
			expectedJSON, _ := json.Marshal(tt.expected)
			if string(gotJSON) != string(expectedJSON) {
				t.Errorf("deserializeYAMLValue(%q) = %s, want %s", tt.input, gotJSON, expectedJSON)
			}
		})
	}
}

func TestGetSettings_Empty(t *testing.T) {
	db := setupSettingsDB(t)
	h := NewSettingsHandler(db)
	adminClaims := &auth.Claims{UserID: 1, Admin: true}

	rr := settingsRequest(t, h.GetSettings, "GET", "/admin/settings", nil, adminClaims)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Errorf("expected empty settings, got %d entries", len(result))
	}
}
