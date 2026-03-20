package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
)

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	token, _ := jwtSvc.GenerateToken(1, "admin", true)

	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims := GetClaims(r)
		if claims == nil {
			t.Error("expected claims in context")
			return
		}
		if claims.UserID != 1 {
			t.Errorf("expected user_id 1, got %d", claims.UserID)
		}
		if claims.Username != "admin" {
			t.Errorf("expected username admin, got %s", claims.Username)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := JWTAuth(jwtSvc)(inner)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected inner handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := JWTAuth(jwtSvc)(inner)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := JWTAuth(jwtSvc)(inner)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer some-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	jwtSvc := auth.NewJWTService("test-secret", -time.Hour)
	token, _ := jwtSvc.GenerateToken(1, "user", false)

	validSvc := auth.NewJWTService("test-secret", time.Hour)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	handler := JWTAuth(validSvc)(inner)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetClaims_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := GetClaims(req)
	if claims != nil {
		t.Error("expected nil claims when no context set")
	}
}
