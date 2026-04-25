package handler

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func exportRouter(t *testing.T) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	// Seed test data
	now := time.Now()
	email := "test@acme.com"
	phone := "555-1234"
	db.Create(&model.Account{
		ID: 1, UserID: 1, AssignedTo: 1, Name: "Acme Corp",
		Access: "Public", Email: &email, Phone: &phone, CreatedAt: now, UpdatedAt: now,
	})

	firstName, lastName := "John", "Doe"
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, email, created_at, updated_at) VALUES (1, 1, 1, ?, ?, 'Public', ?, ?, ?)",
		firstName, lastName, email, now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	// Advance sequences past explicitly-inserted IDs so auto-increment doesn't collide
	db.Exec("SELECT setval('accounts_id_seq', (SELECT COALESCE(MAX(id),0) FROM accounts))")
	db.Exec("SELECT setval('contacts_id_seq', (SELECT COALESCE(MAX(id),0) FROM contacts))")

	return mux, jwtSvc
}

func TestExportAccounts(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/export", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "accounts.csv") {
		t.Errorf("expected Content-Disposition with accounts.csv, got %q", cd)
	}

	reader := csv.NewReader(rec.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected at least 2 rows (header + data), got %d", len(records))
	}
	if records[0][0] != "ID" {
		t.Errorf("expected first header 'ID', got %q", records[0][0])
	}
	if records[1][1] != "Acme Corp" {
		t.Errorf("expected name 'Acme Corp', got %q", records[1][1])
	}
}

func TestExportContacts(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/contacts/export", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	reader := csv.NewReader(rec.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) < 2 {
		t.Fatalf("expected at least 2 rows, got %d", len(records))
	}
	// Header should have "First Name"
	if records[0][1] != "First Name" {
		t.Errorf("expected header 'First Name', got %q", records[0][1])
	}
}

func TestExportAccounts_NoAuth(t *testing.T) {
	mux, _ := exportRouter(t)
	req := httptest.NewRequest("GET", "/api/v1/accounts/export", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- Import Tests ---

func TestImportAccounts(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	csvData := "Name,Email,Phone,Category\nImported Inc,import@test.com,555-9999,customer\nAnother Corp,another@test.com,,partner\n"
	req := httptest.NewRequest("POST", "/api/v1/accounts/import", strings.NewReader(csvData))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "text/csv")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"imported":2`) {
		t.Errorf("expected 2 imported, got: %s", body)
	}
}

func TestImportAccounts_MissingName(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	csvData := "Name,Email\n,bad@test.com\nGood Inc,good@test.com\n"
	req := httptest.NewRequest("POST", "/api/v1/accounts/import", strings.NewReader(csvData))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "text/csv")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"imported":1`) {
		t.Errorf("expected 1 imported (1 skipped), got: %s", body)
	}
	if !strings.Contains(body, "name is required") {
		t.Errorf("expected error about name, got: %s", body)
	}
}

func TestImportContacts(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	csvData := "First Name,Last Name,Email,Phone\nJane,Smith,jane@test.com,555-1111\n"
	req := httptest.NewRequest("POST", "/api/v1/contacts/import", strings.NewReader(csvData))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "text/csv")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"imported":1`) {
		t.Errorf("expected 1 imported, got: %s", body)
	}
}

func TestImportTemplate(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/accounts/import/template", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %q", ct)
	}

	reader := csv.NewReader(rec.Body)
	records, _ := reader.ReadAll()
	if len(records) != 1 {
		t.Errorf("expected 1 row (headers only), got %d", len(records))
	}
}

func TestVCardExport(t *testing.T) {
	mux, jwtSvc := exportRouter(t)
	tok, _ := jwtSvc.GenerateToken(1, "admin", true)

	req := httptest.NewRequest("GET", "/api/v1/contacts/export/vcard", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/vcard" {
		t.Errorf("expected text/vcard, got %q", ct)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "BEGIN:VCARD") {
		t.Error("expected vCard output")
	}
	if !strings.Contains(body, "FN:John Doe") {
		t.Errorf("expected FN:John Doe in vCard, got: %s", body)
	}
}

