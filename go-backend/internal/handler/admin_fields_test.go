package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

func setupCustomFieldsDB(t *testing.T) *gorm.DB {
	db := testDB(t)

	// Drop any cf_* columns left by previous test runs before seeding
	dropCfColumns(db)

	// Seed a field group
	if err := db.Exec(`INSERT INTO field_groups (id, name, label, klass_name, position) VALUES (1, 'custom_fields', 'Custom Fields', 'Account', 0)`).Error; err != nil {
		t.Fatalf("seed field_groups: %v", err)
	}

	// Seed an account
	if err := db.Exec(`INSERT INTO accounts (id, user_id, name, access) VALUES (1, 1, 'Test Account', 'Public')`).Error; err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	t.Cleanup(func() { dropCfColumns(db) })

	return db
}

func dropCfColumns(db *gorm.DB) {
	var cols []struct{ ColumnName string }
	db.Raw(`SELECT column_name FROM information_schema.columns WHERE table_name = 'accounts' AND column_name LIKE 'cf_%'`).Scan(&cols)
	for _, col := range cols {
		db.Exec(fmt.Sprintf("ALTER TABLE accounts DROP COLUMN IF EXISTS %s", col.ColumnName))
	}
}

func customFieldsHandler(db *gorm.DB) (*AdminFieldsHandler, *service.CustomFieldService) {
	repo := repository.NewFieldGroupRepository(db)
	cfSvc := service.NewCustomFieldService(repo)
	return NewAdminFieldsHandler(db, cfSvc), cfSvc
}

func adminFieldsRequest(t *testing.T, r *chi.Mux, method, path string, body interface{}, claims *auth.Claims) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if claims != nil {
		ctx := context.WithValue(req.Context(), middleware.UserClaimsKey, claims)
		req = req.WithContext(ctx)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func setupFieldsRouter(db *gorm.DB) *chi.Mux {
	h, _ := customFieldsHandler(db)
	r := chi.NewRouter()
	r.Post("/admin/fields", h.CreateField)
	r.Put("/admin/fields/{id}", h.UpdateField)
	r.Delete("/admin/fields/{id}", h.DeleteField)
	r.Post("/admin/fields/sort", h.SortFields)
	r.Get("/{entity}/{id}/custom_fields", h.GetEntityCustomFields)
	r.Put("/{entity}/{id}/custom_fields", h.UpdateEntityCustomFields)
	return r
}

func TestCreateField(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Company Size",
		"as":             "integer",
	}

	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var field map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &field)

	if field["name"] != "cf_company_size" {
		t.Errorf("expected name=cf_company_size, got %v", field["name"])
	}
	if field["label"] != "Company Size" {
		t.Errorf("expected label=Company Size, got %v", field["label"])
	}
	if field["as"] != "integer" {
		t.Errorf("expected as=integer, got %v", field["as"])
	}
	if field["type"] != "CustomField" {
		t.Errorf("expected type=CustomField, got %v", field["type"])
	}

	// Verify column was added to accounts table
	var count int
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'accounts' AND column_name = 'cf_company_size'").Scan(&count)
	if count != 1 {
		t.Error("expected cf_company_size column to exist on accounts table")
	}
}

func TestCreateField_NonAdmin(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	user := &auth.Claims{UserID: 2, Admin: false}

	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Test",
		"as":             "string",
	}

	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, user)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestCreateField_InvalidType(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Test",
		"as":             "invalid_type",
	}

	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestCreateField_CollisionAvoidance(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create first field
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Revenue",
		"as":             "decimal",
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create failed: %d %s", rr.Code, rr.Body.String())
	}

	// Create second field with same label
	rr = adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("second create failed: %d %s", rr.Code, rr.Body.String())
	}

	var field map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &field)

	if field["name"] != "cf_revenue_2" {
		t.Errorf("expected collision avoidance name cf_revenue_2, got %v", field["name"])
	}
}

func TestUpdateField(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create a field first
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Notes",
		"as":             "string",
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	fieldID := int(created["id"].(float64))

	// Update label and required
	updateBody := map[string]interface{}{
		"label":    "Important Notes",
		"required": true,
	}
	rr = adminFieldsRequest(t, r, "PUT", fmt.Sprintf("/admin/fields/%d", fieldID), updateBody, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", rr.Code, rr.Body.String())
	}

	var updated map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &updated)
	if updated["label"] != "Important Notes" {
		t.Errorf("expected label=Important Notes, got %v", updated["label"])
	}
	if updated["required"] != true {
		t.Errorf("expected required=true, got %v", updated["required"])
	}
}

func TestUpdateField_CoreFieldBlocked(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Insert a core field
	sqlDB, _ := db.DB()
	sqlDB.Exec(`INSERT INTO fields (id, type, field_group_id, name, label, "as") VALUES (99, 'CoreField', 1, 'name', 'Name', 'string')`)

	updateBody := map[string]interface{}{"label": "New Name"}
	rr := adminFieldsRequest(t, r, "PUT", "/admin/fields/99", updateBody, admin)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDeleteField(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create a field
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "To Delete",
		"as":             "string",
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", rr.Code, rr.Body.String())
	}
	var created map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &created)
	fieldID := int(created["id"].(float64))

	// Delete the field
	rr = adminFieldsRequest(t, r, "DELETE", fmt.Sprintf("/admin/fields/%d", fieldID), nil, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("delete failed: %d %s", rr.Code, rr.Body.String())
	}

	// Verify field is gone
	var count int64
	db.Raw("SELECT COUNT(*) FROM fields WHERE id = ?", fieldID).Scan(&count)
	if count != 0 {
		t.Error("expected field to be deleted")
	}
}

func TestDeleteField_CoreFieldBlocked(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	sqlDB, _ := db.DB()
	sqlDB.Exec(`INSERT INTO fields (id, type, field_group_id, name, label, "as") VALUES (99, 'CoreField', 1, 'name', 'Name', 'string')`)

	rr := adminFieldsRequest(t, r, "DELETE", "/admin/fields/99", nil, admin)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestSortFields(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	sqlDB, _ := db.DB()
	sqlDB.Exec(`INSERT INTO fields (id, type, field_group_id, name, label, "as", position) VALUES (1, 'CustomField', 1, 'cf_a', 'A', 'string', 0)`)
	sqlDB.Exec(`INSERT INTO fields (id, type, field_group_id, name, label, "as", position) VALUES (2, 'CustomField', 1, 'cf_b', 'B', 'string', 1)`)
	sqlDB.Exec(`INSERT INTO fields (id, type, field_group_id, name, label, "as", position) VALUES (3, 'CustomField', 1, 'cf_c', 'C', 'string', 2)`)

	// Reverse order
	body := map[string]interface{}{
		"field_ids": []int{3, 2, 1},
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields/sort", body, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("sort failed: %d %s", rr.Code, rr.Body.String())
	}

	// Verify positions
	var pos int
	db.Raw("SELECT position FROM fields WHERE id = 3").Scan(&pos)
	if pos != 0 {
		t.Errorf("expected field 3 at position 0, got %d", pos)
	}
	db.Raw("SELECT position FROM fields WHERE id = 1").Scan(&pos)
	if pos != 2 {
		t.Errorf("expected field 1 at position 2, got %d", pos)
	}
}

func TestGetEntityCustomFields(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create a custom field and set a value
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Industry",
		"as":             "string",
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create field failed: %d %s", rr.Code, rr.Body.String())
	}

	// Set value on account
	db.Exec("UPDATE accounts SET cf_industry = ? WHERE id = 1", "Technology")

	// Read custom fields
	rr = adminFieldsRequest(t, r, "GET", "/accounts/1/custom_fields", nil, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("get custom fields failed: %d %s", rr.Code, rr.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &result)

	if result["cf_industry"] != "Technology" {
		t.Errorf("expected cf_industry=Technology, got %v", result["cf_industry"])
	}
}

func TestUpdateEntityCustomFields(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create a custom field
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Employees",
		"as":             "integer",
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create field failed: %d %s", rr.Code, rr.Body.String())
	}

	// Update custom field value
	updateBody := map[string]interface{}{
		"cf_employees": 500,
	}
	rr = adminFieldsRequest(t, r, "PUT", "/accounts/1/custom_fields", updateBody, admin)
	if rr.Code != http.StatusOK {
		t.Fatalf("update custom fields failed: %d %s", rr.Code, rr.Body.String())
	}

	// Verify the value was written
	var val int
	db.Raw("SELECT cf_employees FROM accounts WHERE id = 1").Scan(&val)
	if val != 500 {
		t.Errorf("expected cf_employees=500, got %d", val)
	}
}

func TestUpdateEntityCustomFields_UnknownField(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	body := map[string]interface{}{
		"cf_nonexistent": "value",
	}
	rr := adminFieldsRequest(t, r, "PUT", "/accounts/1/custom_fields", body, admin)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateEntityCustomFields_Validation(t *testing.T) {
	db := setupCustomFieldsDB(t)
	r := setupFieldsRouter(db)
	admin := &auth.Claims{UserID: 1, Admin: true}

	// Create a required field
	body := map[string]interface{}{
		"field_group_id": 1,
		"label":          "Required Field",
		"as":             "string",
		"required":       true,
	}
	rr := adminFieldsRequest(t, r, "POST", "/admin/fields", body, admin)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create field failed: %d %s", rr.Code, rr.Body.String())
	}

	// Try to set required field to empty
	updateBody := map[string]interface{}{
		"cf_required_field": "",
	}
	rr = adminFieldsRequest(t, r, "PUT", "/accounts/1/custom_fields", updateBody, admin)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for validation failure, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGenerateColumnName(t *testing.T) {
	tests := []struct {
		label    string
		existing map[string]bool
		expected string
	}{
		{"Company Size", map[string]bool{}, "cf_company_size"},
		{"Revenue", map[string]bool{"cf_revenue": true}, "cf_revenue_2"},
		{"Revenue", map[string]bool{"cf_revenue": true, "cf_revenue_2": true}, "cf_revenue_3"},
		{"My Field!", map[string]bool{}, "cf_my_field"},
		{"  Spaces  ", map[string]bool{}, "cf_spaces"},
		{"UPPER CASE", map[string]bool{}, "cf_upper_case"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := service.GenerateColumnName(tt.label, tt.existing)
			if got != tt.expected {
				t.Errorf("GenerateColumnName(%q) = %q, want %q", tt.label, got, tt.expected)
			}
		})
	}
}

func TestCheckTypeTransition(t *testing.T) {
	tests := []struct {
		from, to string
		expected string
	}{
		{"string", "string", "null"},
		{"string", "text", "safe"},
		{"date", "datetime", "safe"},
		{"integer", "float", "safe"},
		{"string", "integer", "unsafe"},
		{"boolean", "string", "unsafe"},
		{"text", "string", "unsafe"}, // one-way only
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s->%s", tt.from, tt.to), func(t *testing.T) {
			got := checkTypeTransition(tt.from, tt.to)
			if got != tt.expected {
				t.Errorf("checkTypeTransition(%s, %s) = %q, want %q", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}
