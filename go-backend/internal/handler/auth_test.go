package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testDB(t)
}

func seedTestUser(t *testing.T, db *gorm.DB) {
	t.Helper()
	// Hash "Dem0P@ssword!!" with salt "testsalt12345678" at 20 stretches
	password := "Dem0P@ssword!!"
	salt := "testsalt12345678"
	hash := auth.DigestPassword(password, salt, auth.DefaultStretches)

	now := time.Now()
	user := model.User{
		ID:                1,
		Username:          "admin",
		Email:             "admin@example.com",
		FirstName:         "Admin",
		LastName:          "User",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
		ConfirmedAt:       &now,
		SignInCount:        5,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
}

func setupAuthHandler(t *testing.T) (*AuthHandler, *gorm.DB) {
	t.Helper()
	db := setupAuthTestDB(t)
	seedTestUser(t, db)
	userRepo := repository.NewUserRepository(db)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	return NewAuthHandler(userRepo, jwtSvc), db
}

func TestLogin_Success_ByUsername(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "admin", Password: "Dem0P@ssword!!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp loginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "admin" {
		t.Errorf("expected username admin, got %s", resp.User.Username)
	}
	if !resp.User.Admin {
		t.Error("expected admin to be true")
	}
}

func TestLogin_Success_ByEmail(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "admin@example.com", Password: "Dem0P@ssword!!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLogin_Success_CaseInsensitive(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "ADMIN", Password: "Dem0P@ssword!!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for case-insensitive login, got %d", w.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "admin", Password: "wrongpassword"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "nobody", Password: "anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_SuspendedUser(t *testing.T) {
	db := setupAuthTestDB(t)
	salt := "testsalt12345678"
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	now := time.Now()
	user := model.User{
		ID: 2, Username: "suspended", Email: "suspended@example.com",
		EncryptedPassword: hash, PasswordSalt: salt,
		ConfirmedAt: &now, SuspendedAt: &now, SignInCount: 0,
	}
	db.Create(&user)

	userRepo := repository.NewUserRepository(db)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	h := NewAuthHandler(userRepo, jwtSvc)

	body, _ := json.Marshal(loginRequest{Login: "suspended", Password: "Dem0P@ssword!!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for suspended user, got %d", w.Code)
	}
}

func TestLogin_UnconfirmedUser(t *testing.T) {
	db := setupAuthTestDB(t)
	salt := "testsalt12345678"
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	user := model.User{
		ID: 3, Username: "unconfirmed", Email: "unconfirmed@example.com",
		EncryptedPassword: hash, PasswordSalt: salt,
		ConfirmedAt: nil, SignInCount: 0,
	}
	db.Create(&user)

	userRepo := repository.NewUserRepository(db)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	h := NewAuthHandler(userRepo, jwtSvc)

	body, _ := json.Marshal(loginRequest{Login: "unconfirmed", Password: "Dem0P@ssword!!"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unconfirmed user, got %d", w.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	h, _ := setupAuthHandler(t)

	body, _ := json.Marshal(loginRequest{Login: "", Password: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	h, _ := setupAuthHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
