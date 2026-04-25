package handler

import (
	"encoding/json"
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func TestGetProfile(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	// Create a user with ID 1
	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		FirstName:         "Test",
		LastName:          "Admin",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "GET", "/api/v1/profile", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp profileResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Username != "admin" {
		t.Errorf("username: got %q, want admin", resp.Username)
	}
	if resp.Email != "admin@test.com" {
		t.Errorf("email: got %q, want admin@test.com", resp.Email)
	}
	if !resp.Admin {
		t.Error("expected admin=true")
	}
}

func TestGetProfile_NoAuth(t *testing.T) {
	mux, _, _ := writeRouterWithDB(t)
	rec := doRequest(mux, "GET", "/api/v1/profile", "", nil)
	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUpdateProfile(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		FirstName:         "Old",
		LastName:          "Name",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "PUT", "/api/v1/profile", tok, map[string]interface{}{
		"first_name": "New",
		"last_name":  "Person",
		"phone":      "555-1234",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp profileResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.FirstName != "New" {
		t.Errorf("first_name: got %q, want New", resp.FirstName)
	}
	if resp.LastName != "Person" {
		t.Errorf("last_name: got %q, want Person", resp.LastName)
	}
	if resp.Phone != "555-1234" {
		t.Errorf("phone: got %q, want 555-1234", resp.Phone)
	}
}

func TestChangePassword_Success(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "PUT", "/api/v1/profile/password", tok, map[string]interface{}{
		"current_password": "Dem0P@ssword!!",
		"new_password":     "N3wS3cureP@ss!!",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify new password works
	var user model.User
	db.First(&user, 1)
	if !auth.VerifyPassword("N3wS3cureP@ss!!", user.EncryptedPassword, user.PasswordSalt, auth.DefaultStretches) {
		t.Error("new password should verify after change")
	}
	if auth.VerifyPassword("Dem0P@ssword!!", user.EncryptedPassword, user.PasswordSalt, auth.DefaultStretches) {
		t.Error("old password should no longer verify")
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "PUT", "/api/v1/profile/password", tok, map[string]interface{}{
		"current_password": "WrongPassword!!1",
		"new_password":     "N3wS3cureP@ss!!",
	})
	if rec.Code != 403 {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChangePassword_WeakNew(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "PUT", "/api/v1/profile/password", tok, map[string]interface{}{
		"current_password": "Dem0P@ssword!!",
		"new_password":     "weak",
	})
	if rec.Code != 422 {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestChangePassword_MissingFields(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	salt := generateSalt()
	hash := auth.DigestPassword("Dem0P@ssword!!", salt, auth.DefaultStretches)
	db.Create(&model.User{
		Username:          "admin",
		Email:             "admin@test.com",
		Admin:             true,
		EncryptedPassword: hash,
		PasswordSalt:      salt,
	})

	rec := doRequest(mux, "PUT", "/api/v1/profile/password", tok, map[string]interface{}{})
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
