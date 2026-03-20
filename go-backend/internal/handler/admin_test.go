package handler

import (
	"encoding/json"
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// adminRouter reuses writeRouter (same DB setup + JWT + mux).

// --- User Management Tests ---

func TestAdminCreateUser(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "newuser", "email": "new@test.com",
		"password": "SecureP@ssw0rd!!", "first_name": "New", "last_name": "User",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.Username != "newuser" {
		t.Errorf("expected username 'newuser', got %q", user.Username)
	}
	if user.ConfirmedAt == nil {
		t.Error("expected confirmed_at to be set for admin-created user")
	}
}

func TestAdminCreateUser_WeakPassword(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/admin/users", makeToken("admin"), map[string]interface{}{
		"username": "weak", "email": "weak@test.com", "password": "short",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422 for weak password, got %d", rec.Code)
	}
}

func TestAdminCreateUser_NonAdmin(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/admin/users", makeToken("user"), map[string]interface{}{
		"username": "x", "email": "x@test.com", "password": "SecureP@ssw0rd!!",
	})
	if rec.Code != 403 {
		t.Errorf("expected 403 for non-admin, got %d", rec.Code)
	}
}

func TestAdminUpdateUser(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Create user first
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "editme", "email": "edit@test.com", "password": "SecureP@ssw0rd!!",
	})

	// Update - need to find the ID. The admin user is ID 1, so new user gets next ID.
	// In SQLite with fresh DB, admin user doesn't exist in the DB; the first created user gets ID 1.
	rec := doRequest(mux, "PUT", "/api/v1/admin/users/1", tok, map[string]interface{}{
		"first_name": "Updated",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.FirstName != "Updated" {
		t.Errorf("expected first_name 'Updated', got %q", user.FirstName)
	}
}

func TestAdminSuspendUser(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin") // admin is user ID 1

	// Create two users so we can suspend user 2 (not self)
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "first", "email": "first@test.com", "password": "SecureP@ssw0rd!!",
	})
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "suspendme", "email": "suspend@test.com", "password": "SecureP@ssw0rd!!",
	})

	rec := doRequest(mux, "PUT", "/api/v1/admin/users/2/suspend", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.SuspendedAt == nil {
		t.Error("expected suspended_at to be set")
	}
}

func TestAdminReactivateUser(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin") // admin is user ID 1

	// Create two users, suspend user 2, then reactivate
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "first", "email": "first@test.com", "password": "SecureP@ssw0rd!!",
	})
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "reactivateme", "email": "react@test.com", "password": "SecureP@ssw0rd!!",
	})
	doRequest(mux, "PUT", "/api/v1/admin/users/2/suspend", tok, nil)

	rec := doRequest(mux, "PUT", "/api/v1/admin/users/2/reactivate", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var user model.User
	if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if user.SuspendedAt != nil {
		t.Error("expected suspended_at to be nil after reactivation")
	}
}

func TestAdminDeleteUser_CannotDeleteSelf(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin") // admin is user ID 1

	// Create a user so ID 1 exists
	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "admin", "email": "admin@test.com", "password": "SecureP@ssw0rd!!",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/admin/users/1", tok, nil)
	if rec.Code != 422 {
		t.Errorf("expected 422 for self-delete, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSuspend_CannotSuspendSelf(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin") // admin is user ID 1

	doRequest(mux, "POST", "/api/v1/admin/users", tok, map[string]interface{}{
		"username": "admin", "email": "admin@test.com", "password": "SecureP@ssw0rd!!",
	})

	rec := doRequest(mux, "PUT", "/api/v1/admin/users/1/suspend", tok, nil)
	if rec.Code != 422 {
		t.Errorf("expected 422 for self-suspend, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Group Management Tests ---

func TestAdminCreateGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/admin/groups", tok, map[string]interface{}{
		"name": "Sales Team",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var group model.Group
	if err := json.NewDecoder(rec.Body).Decode(&group); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if group.Name != "Sales Team" {
		t.Errorf("expected 'Sales Team', got %q", group.Name)
	}
}

func TestAdminListGroups(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/admin/groups", tok, map[string]interface{}{"name": "Beta"})
	doRequest(mux, "POST", "/api/v1/admin/groups", tok, map[string]interface{}{"name": "Alpha"})

	rec := doRequest(mux, "GET", "/api/v1/admin/groups", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var groups []model.Group
	if err := json.NewDecoder(rec.Body).Decode(&groups); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].Name != "Alpha" {
		t.Errorf("expected first group 'Alpha' (sorted), got %q", groups[0].Name)
	}
}

func TestAdminUpdateGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/admin/groups", tok, map[string]interface{}{"name": "Old Name"})

	rec := doRequest(mux, "PUT", "/api/v1/admin/groups/1", tok, map[string]interface{}{"name": "New Name"})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var group model.Group
	if err := json.NewDecoder(rec.Body).Decode(&group); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if group.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", group.Name)
	}
}

func TestAdminDeleteGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/admin/groups", tok, map[string]interface{}{"name": "Delete Me"})

	rec := doRequest(mux, "DELETE", "/api/v1/admin/groups/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- Field Group Management Tests ---

func TestAdminCreateFieldGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/admin/field_groups", tok, map[string]interface{}{
		"klass_name": "Account", "label": "Custom Info", "name": "custom_info",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminUpdateFieldGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/admin/field_groups", tok, map[string]interface{}{
		"klass_name": "Account", "label": "Old Label", "name": "old_label",
	})

	rec := doRequest(mux, "PUT", "/api/v1/admin/field_groups/1", tok, map[string]interface{}{
		"label": "New Label",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminDeleteFieldGroup(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/admin/field_groups", tok, map[string]interface{}{
		"klass_name": "Account", "label": "Temp", "name": "temp",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/admin/field_groups/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- Admin Auth Tests ---

func TestAdminEndpoints_NonAdmin(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("user") // non-admin

	endpoints := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/admin/users"},
		{"PUT", "/api/v1/admin/users/1"},
		{"DELETE", "/api/v1/admin/users/1"},
		{"PUT", "/api/v1/admin/users/1/suspend"},
		{"PUT", "/api/v1/admin/users/1/reactivate"},
		{"GET", "/api/v1/admin/groups"},
		{"POST", "/api/v1/admin/groups"},
		{"PUT", "/api/v1/admin/groups/1"},
		{"DELETE", "/api/v1/admin/groups/1"},
		{"POST", "/api/v1/admin/field_groups"},
		{"PUT", "/api/v1/admin/field_groups/1"},
		{"DELETE", "/api/v1/admin/field_groups/1"},
	}

	for _, ep := range endpoints {
		rec := doRequest(mux, ep.method, ep.path, tok, map[string]interface{}{"name": "test"})
		if rec.Code != 403 {
			t.Errorf("%s %s: expected 403, got %d", ep.method, ep.path, rec.Code)
		}
	}
}
