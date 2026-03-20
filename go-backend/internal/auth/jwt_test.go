package auth

import (
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-key", time.Hour)

	token, err := svc.GenerateToken(42, "admin", true)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected user_id 42, got %d", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username admin, got %s", claims.Username)
	}
	if !claims.Admin {
		t.Error("expected admin to be true")
	}
	if claims.Issuer != "fat-free-crm-api" {
		t.Errorf("expected issuer fat-free-crm-api, got %s", claims.Issuer)
	}
}

func TestJWT_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret-key", -time.Hour) // Already expired

	token, err := svc.GenerateToken(1, "user", false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWT_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-one", time.Hour)
	svc2 := NewJWTService("secret-two", time.Hour)

	token, _ := svc1.GenerateToken(1, "user", false)

	_, err := svc2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestJWT_InvalidTokenString(t *testing.T) {
	svc := NewJWTService("secret", time.Hour)

	_, err := svc.ValidateToken("not-a-valid-jwt")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
