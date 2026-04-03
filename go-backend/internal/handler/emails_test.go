package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func emailRouter(db *gorm.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := &auth.Claims{UserID: 1, Username: "admin", Admin: true}
			ctx := context.WithValue(r.Context(), middleware.UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	h := NewEmailHandler(db)
	r.Get("/{entity}/{id}/emails", h.ListEmails)
	r.Delete("/emails/{id}", h.DeleteEmail)
	return r
}

func seedEmailUser(db *gorm.DB) {
	db.Exec(`INSERT INTO users (id, username, email, encrypted_password, admin, created_at, updated_at)
		VALUES (1, 'admin', 'admin@test.com', 'x', true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`)
}

func TestListEmails(t *testing.T) {
	db := testDB(t)
	seedEmailUser(db)
	router := emailRouter(db)

	// Create an account
	db.Exec(`INSERT INTO accounts (id, user_id, name, created_at, updated_at) VALUES (1, 1, 'Acme', NOW(), NOW())`)

	// Create emails attached to the account
	now := time.Now().UTC()
	for i := 1; i <= 3; i++ {
		email := model.Email{
			ImapMessageID: fmt.Sprintf("msg-%d", i),
			UserID:        ptr(int64(1)),
			MediatorType:  "Account",
			MediatorID:    1,
			SentFrom:      "sender@example.com",
			SentTo:        "admin@test.com",
			Subject:       fmt.Sprintf("Test Email %d", i),
			Body:          "body",
			State:         "Expanded",
			CreatedAt:     &now,
			UpdatedAt:     &now,
		}
		db.Create(&email)
	}

	// Create an email for a different entity — should NOT appear
	otherEmail := model.Email{
		ImapMessageID: "msg-other",
		UserID:        ptr(int64(1)),
		MediatorType:  "Contact",
		MediatorID:    99,
		SentFrom:      "other@example.com",
		SentTo:        "admin@test.com",
		Subject:       "Other Entity Email",
		Body:          "body",
		State:         "Expanded",
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	db.Create(&otherEmail)

	req := httptest.NewRequest("GET", "/accounts/1/emails", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	body := rr.Body.String()
	for i := 1; i <= 3; i++ {
		want := fmt.Sprintf("Test Email %d", i)
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain %q", want)
		}
	}
	if strings.Contains(body, "Other Entity Email") {
		t.Error("should not include emails for other entities")
	}
}

func TestListEmails_InvalidEntity(t *testing.T) {
	db := testDB(t)
	router := emailRouter(db)

	req := httptest.NewRequest("GET", "/widgets/1/emails", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDeleteEmail(t *testing.T) {
	db := testDB(t)
	seedEmailUser(db)
	router := emailRouter(db)

	now := time.Now().UTC()
	email := model.Email{
		ImapMessageID: "msg-del",
		UserID:        ptr(int64(1)),
		MediatorType:  "Account",
		MediatorID:    1,
		SentFrom:      "sender@example.com",
		SentTo:        "admin@test.com",
		Subject:       "Delete Me",
		Body:          "body",
		State:         "Expanded",
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	db.Create(&email)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/emails/%d", email.ID), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify soft-deleted
	var found model.Email
	err := db.Where("id = ? AND deleted_at IS NULL", email.ID).First(&found).Error
	if err == nil {
		t.Error("email should be soft-deleted")
	}
}

func TestDeleteEmail_NotFound(t *testing.T) {
	db := testDB(t)
	router := emailRouter(db)

	req := httptest.NewRequest("DELETE", "/emails/99999", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func ptr(v int64) *int64 { return &v }
