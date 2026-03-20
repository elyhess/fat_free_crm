package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_HealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRouter_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := NewRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
