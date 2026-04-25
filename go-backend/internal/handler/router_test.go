package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRouterConfig(t *testing.T) RouterConfig {
	t.Helper()
	return RouterConfig{
		DB:             testDB(t),
		JWTSecret:      "test-secret",
		JWTExpiryHours: 1,
	}
}

func TestRouter_HealthEndpoint(t *testing.T) {
	router := NewRouter(testRouterConfig(t))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRouter_NotFound(t *testing.T) {
	router := NewRouter(testRouterConfig(t))

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
