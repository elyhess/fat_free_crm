package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

func authFlowsRouter(t *testing.T) *http.ServeMux {
	t.Helper()
	db := testDB(t)

	// Seed a confirmed user
	salt := "testsalt"
	encrypted := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	now := time.Now().UTC()
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, confirmed_at, created_at, updated_at)
		VALUES (1, 'testuser', 'test@example.com', ?, ?, false, 'Test', 'User', ?, ?, ?)`,
		encrypted, salt, now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	// Create a no-op email service (won't actually send)
	emailSvc := service.NewEmailService(service.EmailConfig{
		Host: "localhost",
		Port: "0", // won't connect
		From: "test@test.com",
	})

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)

	// Override the auth flows handler with our test DB
	authFlows := NewAuthFlowsHandler(db, emailSvc, "http://localhost:3000")
	_ = authFlows // routes already registered via NewRouter

	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux
}

// --- Password Reset Tests ---

func TestForgotPassword(t *testing.T) {
	mux := authFlowsRouter(t)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Should always return success (no email enumeration)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] == "" {
		t.Error("expected status message")
	}
}

func TestForgotPassword_UnknownEmail(t *testing.T) {
	mux := authFlowsRouter(t)

	body := `{"email":"nobody@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Should still return 200 (no email enumeration)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResetPassword(t *testing.T) {
	db := testDB(t)

	// Seed user with a known reset token
	salt := "testsalt"
	encrypted := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	now := time.Now().UTC()
	rawToken := "test-reset-token-123"
	hashedToken := hashToken(rawToken)

	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, confirmed_at, reset_password_token, reset_password_sent_at, created_at, updated_at)
		VALUES (1, 'testuser', 'test@example.com', ?, ?, false, 'Test', 'User', ?, ?, ?, ?, ?)`,
		encrypted, salt, now.Format("2006-01-02 15:04:05"), hashedToken, now.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	emailSvc := service.NewEmailService(service.EmailConfig{Host: "localhost", Port: "0", From: "test@test.com"})
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	_ = NewAuthFlowsHandler(db, emailSvc, "http://localhost:3000")

	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"token":"` + rawToken + `","new_password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	mux := authFlowsRouter(t)

	body := `{"token":"invalid-token","new_password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	db := testDB(t)

	salt := "testsalt"
	encrypted := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	now := time.Now().UTC()
	rawToken := "test-reset-token-456"
	hashedToken := hashToken(rawToken)

	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, confirmed_at, reset_password_token, reset_password_sent_at, created_at, updated_at)
		VALUES (1, 'testuser', 'test@example.com', ?, ?, false, 'Test', 'User', ?, ?, ?, ?, ?)`,
		encrypted, salt, now.Format("2006-01-02 15:04:05"), hashedToken, now.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"token":"` + rawToken + `","new_password":"weak"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Registration Tests ---

func TestRegister_NotAllowed(t *testing.T) {
	mux := authFlowsRouter(t)

	// Default: no user_signup setting = not_allowed
	body := `{"username":"newuser","email":"new@example.com","password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRegister_Allowed(t *testing.T) {
	db := testDB(t)

	// Set signup to allowed
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO settings (name, value, created_at, updated_at) VALUES ('user_signup', '--- :allowed\n', ?, ?)", now, now)

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"username":"newuser","email":"new@example.com","password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp["status"], "confirm") {
		t.Errorf("expected confirmation message, got: %s", resp["status"])
	}
}

func TestRegister_NeedsApproval(t *testing.T) {
	db := testDB(t)

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO settings (name, value, created_at, updated_at) VALUES ('user_signup', '--- :needs_approval\n', ?, ?)", now, now)

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"username":"newuser","email":"new@example.com","password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp["status"], "approval") {
		t.Errorf("expected approval message, got: %s", resp["status"])
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	db := testDB(t)

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO settings (name, value, created_at, updated_at) VALUES ('user_signup', '--- :allowed\n', ?, ?)", now, now)
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, created_at, updated_at)
		VALUES (1, 'existing', 'existing@example.com', 'x', 'y', false, 'E', 'U', ?, ?)`, now, now)

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"username":"existing","email":"other@example.com","password":"NewP@ssword123!"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 409 {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Confirmation Tests ---

func TestConfirmEmail(t *testing.T) {
	db := testDB(t)

	now := time.Now().UTC()
	rawToken := "confirm-token-abc"
	hashedToken := hashToken(rawToken)

	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, password_salt, admin, first_name, last_name, confirmation_token, confirmation_sent_at, created_at, updated_at)
		VALUES (1, 'unconfirmed', 'unconfirmed@example.com', 'x', 'y', false, 'U', 'C', ?, ?, ?, ?)`,
		hashedToken, now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	body := `{"token":"` + rawToken + `"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "email confirmed" {
		t.Errorf("expected 'email confirmed', got: %s", resp["status"])
	}
}

func TestConfirmEmail_InvalidToken(t *testing.T) {
	mux := authFlowsRouter(t)

	body := `{"token":"bad-token"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/confirm", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResendConfirmation(t *testing.T) {
	mux := authFlowsRouter(t)

	// User is already confirmed, so this is a no-op but should still return 200
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/resend-confirmation", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
