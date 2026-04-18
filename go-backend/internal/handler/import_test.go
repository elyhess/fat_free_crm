package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// doCSVRequest sends a raw CSV body request (Content-Type: text/csv).
func doCSVRequest(mux *http.ServeMux, method, path, token, csvData string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(csvData))
	req.Header.Set("Content-Type", "text/csv")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// doMultipartCSVRequest sends a multipart form upload with a CSV file.
func doMultipartCSVRequest(mux *http.ServeMux, method, path, token, csvData string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "import.csv")
	_, _ = part.Write([]byte(csvData))
	_ = writer.Close()

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// ========== Import Template Tests ==========

func TestImportTemplate_Accounts(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "GET", "/api/v1/accounts/import/template", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Errorf("expected Content-Type text/csv, got %q", ct)
	}

	body := rec.Body.String()
	expectedHeaders := []string{"Name", "Email", "Phone", "Website", "Category", "Rating", "Access"}
	for _, h := range expectedHeaders {
		if !strings.Contains(body, h) {
			t.Errorf("expected header %q in CSV, got %q", h, body)
		}
	}
}

func TestImportTemplate_Contacts(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "GET", "/api/v1/contacts/import/template", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	expectedHeaders := []string{"First Name", "Last Name", "Title", "Department", "Email", "Phone", "Mobile", "Access"}
	for _, h := range expectedHeaders {
		if !strings.Contains(body, h) {
			t.Errorf("expected header %q in CSV, got %q", h, body)
		}
	}
}

func TestImportTemplate_Leads(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "GET", "/api/v1/leads/import/template", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	expectedHeaders := []string{"First Name", "Last Name", "Company", "Title", "Email", "Phone", "Source", "Status", "Access"}
	for _, h := range expectedHeaders {
		if !strings.Contains(body, h) {
			t.Errorf("expected header %q in CSV, got %q", h, body)
		}
	}
}

// ========== Import Accounts ==========

func TestImportAccounts_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	csv := "Name,Email,Phone\nAcme Corp,acme@example.com,555-1234\nGlobex,globex@example.com,555-5678\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/accounts/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 2 {
		t.Errorf("expected imported=2, got %v", result["imported"])
	}
	if total, ok := result["total"].(float64); !ok || total != 2 {
		t.Errorf("expected total=2, got %v", result["total"])
	}
}

func TestImportAccounts_MissingRequiredField(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// CSV without "name" column — rows will fail validation
	csv := "Email,Phone\nacme@example.com,555-1234\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/accounts/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 0 {
		t.Errorf("expected imported=0 (missing name), got %v", result["imported"])
	}
	errs, ok := result["errors"].([]interface{})
	if !ok || len(errs) == 0 {
		t.Error("expected errors to be non-empty when name column is missing")
	}
}

func TestImportAccounts_EmptyFile(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Only headers, no data rows
	csv := "Name,Email,Phone\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/accounts/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 0 {
		t.Errorf("expected imported=0 for empty file, got %v", result["imported"])
	}
	if total, ok := result["total"].(float64); !ok || total != 0 {
		t.Errorf("expected total=0 for empty file, got %v", result["total"])
	}
}

func TestImportAccounts_Unauthorized(t *testing.T) {
	mux, _, _ := writeRouter(t)

	csv := "Name\nAcme Corp\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/accounts/import", "", csv)
	if rec.Code != 401 {
		t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Import Contacts ==========

func TestImportContacts_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	csv := "First Name,Last Name,Email,Phone\nJane,Smith,jane@example.com,555-1234\nJohn,Doe,john@example.com,555-5678\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/contacts/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 2 {
		t.Errorf("expected imported=2, got %v", result["imported"])
	}
}

// ========== Import Leads ==========

func TestImportLeads_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	csv := "First Name,Last Name,Company,Email\nAlice,Wonder,Wonderland Inc,alice@example.com\nBob,Builder,BuildCo,bob@example.com\n"

	rec := doMultipartCSVRequest(mux, "POST", "/api/v1/leads/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 2 {
		t.Errorf("expected imported=2, got %v", result["imported"])
	}
}

// ========== VCard Export ==========

func TestVCardExportContacts(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Create some contacts first
	doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith", "email": "jane@example.com", "phone": "555-1234",
	})
	doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "John", "last_name": "Doe", "email": "john@example.com",
	})

	rec := doRequest(mux, "GET", "/api/v1/contacts/export/vcard", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/vcard") {
		t.Errorf("expected Content-Type text/vcard, got %q", ct)
	}

	body := rec.Body.String()

	// Check vCard structure
	if !strings.Contains(body, "BEGIN:VCARD") {
		t.Error("expected BEGIN:VCARD in output")
	}
	if !strings.Contains(body, "END:VCARD") {
		t.Error("expected END:VCARD in output")
	}
	if !strings.Contains(body, "VERSION:3.0") {
		t.Error("expected VERSION:3.0 in output")
	}
	if !strings.Contains(body, "FN:Jane Smith") {
		t.Error("expected FN:Jane Smith in output")
	}
	if !strings.Contains(body, "FN:John Doe") {
		t.Error("expected FN:John Doe in output")
	}
	if !strings.Contains(body, "EMAIL;TYPE=INTERNET:jane@example.com") {
		t.Error("expected jane's email in output")
	}
	if !strings.Contains(body, "TEL;TYPE=WORK:555-1234") {
		t.Error("expected jane's phone in output")
	}

	// Should have exactly 2 vCards
	vcardCount := strings.Count(body, "BEGIN:VCARD")
	if vcardCount != 2 {
		t.Errorf("expected 2 vCards, got %d", vcardCount)
	}
}

// ========== Import with raw CSV body ==========

func TestImportAccounts_RawCSVBody(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	csv := "Name,Email\nRaw Corp,raw@example.com\n"

	rec := doCSVRequest(mux, "POST", "/api/v1/accounts/import", tok, csv)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if imported, ok := result["imported"].(float64); !ok || imported != 1 {
		t.Errorf("expected imported=1, got %v", result["imported"])
	}
}
