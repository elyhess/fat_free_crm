package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSubscribe(t *testing.T) {
	db := setupEntitiesDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("POST", "/api/v1/accounts/1/subscribe", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["subscribed"] != true {
		t.Error("expected subscribed = true")
	}
	users := resp["subscribed_users"].([]interface{})
	if len(users) != 1 || int64(users[0].(float64)) != 1 {
		t.Errorf("expected [1], got %v", users)
	}
}

func TestSubscribe_Idempotent(t *testing.T) {
	db := setupEntitiesDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	token := adminToken(t, jwtSvc)

	// Subscribe twice
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/accounts/1/subscribe", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Fatalf("attempt %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Verify only subscribed once
	var resp map[string]interface{}
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/subscription", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	json.NewDecoder(rec.Body).Decode(&resp)

	users := resp["subscribed_users"].([]interface{})
	if len(users) != 1 {
		t.Errorf("expected 1 subscriber after double subscribe, got %d", len(users))
	}
}

func TestUnsubscribe(t *testing.T) {
	db := setupEntitiesDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, subscribed_users, created_at, updated_at) VALUES (1, 1, 0, 'Acme', 'Public', ?, ?, ?)",
		"---\n- 1\n", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("POST", "/api/v1/accounts/1/unsubscribe", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["subscribed"] != false {
		t.Error("expected subscribed = false")
	}
	users := resp["subscribed_users"].([]interface{})
	if len(users) != 0 {
		t.Errorf("expected empty subscribers, got %v", users)
	}
}

func TestGetSubscription(t *testing.T) {
	db := setupEntitiesDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, subscribed_users, created_at, updated_at) VALUES (1, 1, 0, 'Acme', 'Public', ?, ?, ?)",
		"---\n- 1\n- 5\n", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// Admin (user 1) is subscribed
	req := httptest.NewRequest("GET", "/api/v1/accounts/1/subscription", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["subscribed"] != true {
		t.Error("expected subscribed = true for user 1")
	}

	// User 3 is NOT subscribed
	req2 := httptest.NewRequest("GET", "/api/v1/accounts/1/subscription", nil)
	req2.Header.Set("Authorization", "Bearer "+userToken(t, jwtSvc, 3))
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	var resp2 map[string]interface{}
	json.NewDecoder(rec2.Body).Decode(&resp2)
	if resp2["subscribed"] != false {
		t.Error("expected subscribed = false for user 3")
	}
}

func TestSubscribe_NotFound(t *testing.T) {
	db := setupEntitiesDB(t)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("POST", "/api/v1/accounts/999/subscribe", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestSubscribe_NoAuth(t *testing.T) {
	db := setupEntitiesDB(t)
	mux, _ := entitiesRouter(t, db)

	req := httptest.NewRequest("POST", "/api/v1/accounts/1/subscribe", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestSubscribe_MultipleEntities(t *testing.T) {
	db := setupEntitiesDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (1, 1, 0, 'John', 'Doe', 'Public', 'new', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("POST", "/api/v1/leads/1/subscribe", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200 for leads subscribe, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestParseYAMLIntArray(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{"", nil},
		{"--- []\n", nil},
		{"---\n- 1\n", []int64{1}},
		{"---\n- 1\n- 2\n- 3\n", []int64{1, 2, 3}},
		{"---\n- 42\n- 7\n", []int64{42, 7}},
	}

	for _, tc := range tests {
		got := parseYAMLIntArray(tc.input)
		if len(got) == 0 && len(tc.expected) == 0 {
			continue
		}
		if len(got) != len(tc.expected) {
			t.Errorf("parseYAMLIntArray(%q): expected %v, got %v", tc.input, tc.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("parseYAMLIntArray(%q)[%d]: expected %d, got %d", tc.input, i, tc.expected[i], got[i])
			}
		}
	}
}

func TestSerializeYAMLIntArray(t *testing.T) {
	tests := []struct {
		input    []int64
		expected string
	}{
		{[]int64{}, "--- []\n"},
		{[]int64{1}, "---\n- 1\n"},
		{[]int64{1, 2, 3}, "---\n- 1\n- 2\n- 3\n"},
	}

	for _, tc := range tests {
		got := serializeYAMLIntArray(tc.input)
		if got != tc.expected {
			t.Errorf("serializeYAMLIntArray(%v): expected %q, got %q", tc.input, tc.expected, got)
		}
	}
}
