package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	authpkg "github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

func setupFieldsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testDB(t)

	groups := []model.FieldGroup{
		{ID: 1, Name: "general", Label: "General", Position: 1, KlassName: "Account"},
	}
	for _, g := range groups {
		db.Create(&g)
	}
	fields := []model.Field{
		{ID: 1, Type: "CustomField", FieldGroupID: 1, Position: 1, Name: "cf_size", Label: "Size", As: "integer"},
	}
	for _, f := range fields {
		db.Create(&f)
	}

	return db
}

func setupFieldsHandler(t *testing.T) *FieldsHandler {
	t.Helper()
	db := setupFieldsTestDB(t)
	repo := repository.NewFieldGroupRepository(db)
	svc := service.NewCustomFieldService(repo)
	return NewFieldsHandler(svc)
}

func TestListFieldGroups_OK(t *testing.T) {
	h := setupFieldsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups?entity=Account", nil)
	w := httptest.NewRecorder()

	h.ListFieldGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if resp["entity_type"] != "Account" {
		t.Errorf("expected entity_type Account, got %v", resp["entity_type"])
	}

	groups, ok := resp["field_groups"].([]interface{})
	if !ok {
		t.Fatal("expected field_groups to be an array")
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
}

func TestListFieldGroups_MissingEntity(t *testing.T) {
	h := setupFieldsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups", nil)
	w := httptest.NewRecorder()

	h.ListFieldGroups(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListFieldGroups_InvalidEntity(t *testing.T) {
	h := setupFieldsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups?entity=Bogus", nil)
	w := httptest.NewRecorder()

	h.ListFieldGroups(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListFieldGroups_EmptyResult(t *testing.T) {
	h := setupFieldsHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups?entity=Lead", nil)
	w := httptest.NewRecorder()

	h.ListFieldGroups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	groups := resp["field_groups"].([]interface{})
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestListFieldGroups_ViaRouter(t *testing.T) {
	db := setupFieldsTestDB(t)
	jwtSecret := "test-secret"
	router := NewRouter(RouterConfig{DB: db, JWTSecret: jwtSecret, JWTExpiryHours: 1})

	jwtSvc := authpkg.NewJWTService(jwtSecret, time.Hour)
	token, err := jwtSvc.GenerateToken(1, "testuser", false)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups?entity=Account", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 via router, got %d", w.Code)
	}
}

func TestProtectedRoute_NoToken(t *testing.T) {
	db := setupFieldsTestDB(t)
	router := NewRouter(RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/field_groups?entity=Account", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}
