package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// writeRouterWithDB is like writeRouter but also returns the underlying DB for assertions.
func writeRouterWithDB(t *testing.T) (*http.ServeMux, func(string) string, *gorm.DB) {
	t.Helper()
	db := setupSupportingDB(t)
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
	return mux, makeToken, db
}

func TestCreateAccount_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{
		"name": "Versioned Corp", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var acct model.Account
	json.NewDecoder(rec.Body).Decode(&acct)

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Account", acct.ID, "create").First(&ver).Error; err != nil {
		t.Fatalf("no version record created: %v", err)
	}
	if ver.Whodunnit == nil || *ver.Whodunnit != "1" {
		t.Errorf("whodunnit: got %v, want '1'", ver.Whodunnit)
	}
}

func TestUpdateAccount_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{
		"name": "Before Update", "access": "Public",
	})
	var acct model.Account
	json.NewDecoder(rec.Body).Decode(&acct)

	rec = doRequest(mux, "PUT", fmt.Sprintf("/api/v1/accounts/%d", acct.ID), tok, map[string]interface{}{
		"name": "After Update",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Account", acct.ID, "update").First(&ver).Error; err != nil {
		t.Fatalf("no update version record: %v", err)
	}

	if ver.Object == nil {
		t.Fatal("object should not be nil")
	}
	if ver.ObjectChanges == nil {
		t.Fatal("object_changes should not be nil")
	}
	var changes map[string]interface{}
	json.Unmarshal([]byte(*ver.ObjectChanges), &changes)
	if changes["name"] != "After Update" {
		t.Errorf("changes.name: got %v, want 'After Update'", changes["name"])
	}
}

func TestDeleteAccount_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{
		"name": "To Delete", "access": "Public",
	})
	var acct model.Account
	json.NewDecoder(rec.Body).Decode(&acct)

	rec = doRequest(mux, "DELETE", fmt.Sprintf("/api/v1/accounts/%d", acct.ID), tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Account", acct.ID, "destroy").First(&ver).Error; err != nil {
		t.Fatalf("no destroy version record: %v", err)
	}
}

func TestCreateContact_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Doe", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var contact model.Contact
	json.NewDecoder(rec.Body).Decode(&contact)

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Contact", contact.ID, "create").First(&ver).Error; err != nil {
		t.Fatalf("no version record: %v", err)
	}
}

func TestCreateOpportunity_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/opportunities", tok, map[string]interface{}{
		"name": "Big Deal", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var opp model.Opportunity
	json.NewDecoder(rec.Body).Decode(&opp)

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Opportunity", opp.ID, "create").First(&ver).Error; err != nil {
		t.Fatalf("no version record: %v", err)
	}
}

func TestCreateCampaign_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/campaigns", tok, map[string]interface{}{
		"name": "Spring Sale", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var campaign model.Campaign
	json.NewDecoder(rec.Body).Decode(&campaign)

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Campaign", campaign.ID, "create").First(&ver).Error; err != nil {
		t.Fatalf("no version record: %v", err)
	}
}

func TestRejectLead_CreatesVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "Bob", "last_name": "Smith", "access": "Public",
	})
	var lead model.Lead
	json.NewDecoder(rec.Body).Decode(&lead)

	rec = doRequest(mux, "PUT", fmt.Sprintf("/api/v1/leads/%d/reject", lead.ID), tok, nil)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var ver model.Version
	if err := db.Where("item_type = ? AND item_id = ? AND event = ?", "Lead", lead.ID, "update").First(&ver).Error; err != nil {
		t.Fatalf("no update version for reject: %v", err)
	}
	if ver.ObjectChanges == nil {
		t.Fatal("object_changes should contain status change")
	}
	var changes map[string]interface{}
	json.Unmarshal([]byte(*ver.ObjectChanges), &changes)
	if changes["status"] != "rejected" {
		t.Errorf("changes.status: got %v, want 'rejected'", changes["status"])
	}
}

func TestCreateTask_NoVersion(t *testing.T) {
	mux, makeToken, db := writeRouterWithDB(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/tasks", tok, map[string]interface{}{
		"name": "No tracking",
	})

	var count int64
	db.Model(&model.Version{}).Where("item_type = ?", "Task").Count(&count)
	if count != 0 {
		t.Errorf("tasks should not create versions, found %d", count)
	}
}
