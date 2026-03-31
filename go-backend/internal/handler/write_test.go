package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func writeRouter(t *testing.T) (*http.ServeMux, *auth.JWTService, func(string) string) {
	t.Helper()
	db := testDB(t)
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)

	makeToken := func(role string) string {
		isAdmin := role == "admin"
		userID := int64(1)
		if !isAdmin {
			userID = 5
		}
		tok, _ := jwtSvc.GenerateToken(userID, role, isAdmin)
		return tok
	}
	return mux, jwtSvc, makeToken
}

func doRequest(mux *http.ServeMux, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// --- Task Tests ---

func TestCreateTask(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{
		"name": "Call Bob", "priority": "high", "bucket": "due_today",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var task model.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Name != "Call Bob" {
		t.Errorf("expected 'Call Bob', got %q", task.Name)
	}
	if task.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", task.UserID)
	}
}

func TestCreateTask_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/tasks", makeToken("admin"), map[string]interface{}{})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestUpdateTask(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Create
	doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{"name": "Original"})

	// Update
	rec := doRequest(mux, "PUT", "/api/v1/tasks/1", tok, map[string]interface{}{"name": "Updated"})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var task model.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Name != "Updated" {
		t.Errorf("expected 'Updated', got %q", task.Name)
	}
}

func TestDeleteTask(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{"name": "To Delete"})
	rec := doRequest(mux, "DELETE", "/api/v1/tasks/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify deleted
	rec2 := doRequest(mux, "GET", "/api/v1/tasks/1", tok, nil)
	if rec2.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestCompleteTask(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{"name": "Complete Me"})
	rec := doRequest(mux, "PUT", "/api/v1/tasks/1/complete", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var task model.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
	if task.CompletedBy == nil || *task.CompletedBy != 1 {
		t.Error("expected completed_by to be 1")
	}
}

func TestUncompleteTask(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{"name": "Toggle"})
	doRequest(mux, "PUT", "/api/v1/tasks/1/complete", tok, nil)
	rec := doRequest(mux, "PUT", "/api/v1/tasks/1/uncomplete", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var task model.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.CompletedAt != nil {
		t.Error("expected completed_at to be nil after uncomplete")
	}
}

func TestTaskForbiddenForOtherUser(t *testing.T) {
	mux, jwtSvc, makeToken := writeRouter(t)

	// Create as admin (user 1)
	doRequest(mux, "POST", "/api/v1/tasks", makeToken("admin"), map[string]interface{}{"name": "Admin Task"})

	// Try update as user 5 (not owner/assignee)
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	rec := doRequest(mux, "PUT", "/api/v1/tasks/1", tok5, map[string]interface{}{"name": "Hacked"})
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// --- Account Tests ---

func TestCreateAccount(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/accounts", makeToken("admin"), map[string]interface{}{
		"name": "Acme Corp", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateAccount(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")
	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Old Name"})

	rec := doRequest(mux, "PUT", "/api/v1/accounts/1", tok, map[string]interface{}{"name": "New Name"})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var acct model.Account
	if err := json.NewDecoder(rec.Body).Decode(&acct); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if acct.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", acct.Name)
	}
}

func TestDeleteAccount(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")
	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Delete Me"})
	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// --- Lead Tests ---

func TestCreateLead(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/leads", makeToken("admin"), map[string]interface{}{
		"first_name": "John", "last_name": "Doe", "company": "Acme",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRejectLead(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")
	doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Doe",
	})

	rec := doRequest(mux, "PUT", "/api/v1/leads/1/reject", tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var lead model.Lead
	if err := json.NewDecoder(rec.Body).Decode(&lead); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if lead.Status == nil || *lead.Status != "rejected" {
		t.Errorf("expected status 'rejected', got %v", lead.Status)
	}
}

// --- Campaign Tests ---

func TestCreateCampaign(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/campaigns", makeToken("admin"), map[string]interface{}{
		"name": "Q1 Push", "status": "planned",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Contact Tests ---

func TestCreateContact(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/contacts", makeToken("admin"), map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Opportunity Tests ---

func TestCreateOpportunity(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/opportunities", makeToken("admin"), map[string]interface{}{
		"name": "Big Deal", "stage": "prospecting", "amount": 50000, "probability": 25,
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateOpportunity_Stage(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")
	doRequest(mux, "POST", "/api/v1/opportunities", tok, map[string]interface{}{
		"name": "Deal", "stage": "prospecting",
	})

	rec := doRequest(mux, "PUT", "/api/v1/opportunities/1", tok, map[string]interface{}{
		"stage": "negotiation", "probability": 75,
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var opp model.Opportunity
	if err := json.NewDecoder(rec.Body).Decode(&opp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if opp.Stage == nil || *opp.Stage != "negotiation" {
		t.Errorf("expected stage 'negotiation', got %v", opp.Stage)
	}
}

// --- Comment Tests ---

func TestCreateComment(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	// Create an account first
	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{
		"comment": "Great company!", "private": false,
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var comment model.Comment
	if err := json.NewDecoder(rec.Body).Decode(&comment); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if comment.Comment != "Great company!" {
		t.Errorf("expected 'Great company!', got %q", comment.Comment)
	}
	if comment.CommentableType != "Account" {
		t.Errorf("expected commentable_type 'Account', got %q", comment.CommentableType)
	}
}

func TestCreateComment_EmptyBody(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/accounts/1/comments", makeToken("admin"), map[string]interface{}{
		"comment": "",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestDeleteComment(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{"comment": "Delete me"})

	rec := doRequest(mux, "DELETE", "/api/v1/comments/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteComment_ForbiddenForOtherUser(t *testing.T) {
	mux, jwtSvc, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/comments", tok, map[string]interface{}{"comment": "Admin comment"})

	// Try to delete as non-admin user 5
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	rec := doRequest(mux, "DELETE", "/api/v1/comments/1", tok5, nil)
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// --- Tag Tests ---

func TestAddTag(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var tag model.Tag
	if err := json.NewDecoder(rec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tag.Name != "vip" {
		t.Errorf("expected tag 'vip', got %q", tag.Name)
	}
}

func TestAddTag_Duplicate(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})

	// Adding same tag again should return 200 (idempotent)
	rec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "vip"})
	if rec.Code != 200 {
		t.Errorf("expected 200 for duplicate tag, got %d", rec.Code)
	}
}

func TestRemoveTag(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	createRec := doRequest(mux, "POST", "/api/v1/accounts/1/tags", tok, map[string]interface{}{"name": "removeme"})

	var tag model.Tag
	if err := json.NewDecoder(createRec.Body).Decode(&tag); err != nil {
		t.Fatalf("decode: %v", err)
	}

	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1/tags/"+strconv.FormatInt(tag.ID, 10), tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Address Tests ---

func TestCreateAddress(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})

	rec := doRequest(mux, "POST", "/api/v1/accounts/1/addresses", tok, map[string]interface{}{
		"street1": "123 Main St", "city": "Anytown", "state": "CA", "country": "US",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var addr model.Address
	if err := json.NewDecoder(rec.Body).Decode(&addr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if addr.AddressableType != "Account" {
		t.Errorf("expected addressable_type 'Account', got %q", addr.AddressableType)
	}
}

func TestDeleteAddress(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Acme"})
	doRequest(mux, "POST", "/api/v1/accounts/1/addresses", tok, map[string]interface{}{
		"street1": "Delete me",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/addresses/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateComment_InvalidEntity(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/invalid/1/comments", makeToken("admin"), map[string]interface{}{
		"comment": "test",
	})
	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid entity, got %d", rec.Code)
	}
}

// --- Auth Tests ---

func TestWriteEndpoints_NoAuth(t *testing.T) {
	mux, _, _ := writeRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/tasks"},
		{"PUT", "/api/v1/tasks/1"},
		{"DELETE", "/api/v1/tasks/1"},
		{"POST", "/api/v1/accounts"},
		{"POST", "/api/v1/leads"},
		{"POST", "/api/v1/campaigns"},
		{"POST", "/api/v1/contacts"},
		{"POST", "/api/v1/opportunities"},
		{"POST", "/api/v1/accounts/1/comments"},
		{"DELETE", "/api/v1/comments/1"},
		{"POST", "/api/v1/accounts/1/tags"},
		{"DELETE", "/api/v1/accounts/1/tags/1"},
		{"POST", "/api/v1/accounts/1/addresses"},
		{"DELETE", "/api/v1/addresses/1"},
	}

	for _, ep := range endpoints {
		rec := doRequest(mux, ep.method, ep.path, "", map[string]interface{}{"name": "test"})
		if rec.Code != 401 {
			t.Errorf("%s %s: expected 401, got %d", ep.method, ep.path, rec.Code)
		}
	}
}
