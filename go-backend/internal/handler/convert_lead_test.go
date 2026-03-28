package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func setupConvertDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupRelationshipsDB(t) // includes entity tables + join tables
	return db
}

func insertLead(db *gorm.DB, id, userID int64, campaignID *int64, status string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	if campaignID != nil {
		db.Exec(`INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, campaign_id, company, title, email, phone, mobile, do_not_call, background_info, created_at, updated_at)
			VALUES (?, ?, 0, 'John', 'Doe', 'Public', ?, ?, 'Acme Inc', 'VP Sales', 'john@example.com', '555-0100', '555-0101', 0, 'Met at conference', ?, ?)`,
			id, userID, status, *campaignID, now, now)
	} else {
		db.Exec(`INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, company, title, email, phone, mobile, do_not_call, background_info, created_at, updated_at)
			VALUES (?, ?, 0, 'John', 'Doe', 'Public', ?, 'Acme Inc', 'VP Sales', 'john@example.com', '555-0100', '555-0101', 0, 'Met at conference', ?, ?)`,
			id, userID, status, now, now)
	}
}

func TestConvertLead_NewAccount(t *testing.T) {
	db := setupConvertDB(t)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{
		"account": {"name": "New Corp"},
		"opportunity": {"name": "Big Deal", "stage": "prospecting", "amount": 50000, "probability": 25},
		"access": "Public"
	}`

	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp convertLeadResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Account created
	if resp.Account.Name != "New Corp" {
		t.Errorf("expected account name 'New Corp', got %q", resp.Account.Name)
	}

	// Contact created from lead data
	if resp.Contact.FirstName != "John" || resp.Contact.LastName != "Doe" {
		t.Errorf("expected contact 'John Doe', got %q %q", resp.Contact.FirstName, resp.Contact.LastName)
	}
	if resp.Contact.LeadID == nil || *resp.Contact.LeadID != 1 {
		t.Error("expected contact.lead_id = 1")
	}
	if resp.Contact.Email == nil || *resp.Contact.Email != "john@example.com" {
		t.Error("expected contact email copied from lead")
	}

	// Opportunity created
	if resp.Opportunity.Name != "Big Deal" {
		t.Errorf("expected opportunity name 'Big Deal', got %q", resp.Opportunity.Name)
	}
	if resp.Opportunity.Amount == nil || *resp.Opportunity.Amount != 50000 {
		t.Error("expected opportunity amount = 50000")
	}

	// Lead status set to converted
	var lead model.Lead
	db.First(&lead, 1)
	if lead.Status == nil || *lead.Status != "converted" {
		t.Errorf("expected lead status 'converted', got %v", lead.Status)
	}

	// Join records created
	var acCount int64
	db.Table("account_contacts").Where("account_id = ? AND contact_id = ?", resp.Account.ID, resp.Contact.ID).Count(&acCount)
	if acCount != 1 {
		t.Error("expected account_contact join record")
	}

	var aoCount int64
	db.Table("account_opportunities").Where("account_id = ? AND opportunity_id = ?", resp.Account.ID, resp.Opportunity.ID).Count(&aoCount)
	if aoCount != 1 {
		t.Error("expected account_opportunity join record")
	}

	var coCount int64
	db.Table("contact_opportunities").Where("contact_id = ? AND opportunity_id = ?", resp.Contact.ID, resp.Opportunity.ID).Count(&coCount)
	if coCount != 1 {
		t.Error("expected contact_opportunity join record")
	}
}

func TestConvertLead_ExistingAccount(t *testing.T) {
	db := setupConvertDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, contacts_count, opportunities_count, created_at, updated_at) VALUES (1, 1, 0, 'Existing Corp', 'Public', 0, 0, ?, ?)", now, now)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{
		"account": {"id": 1},
		"opportunity": {"name": "Existing Deal"},
		"access": "Public"
	}`

	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp convertLeadResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Account.ID != 1 {
		t.Errorf("expected existing account ID 1, got %d", resp.Account.ID)
	}
	if resp.Account.Name != "Existing Corp" {
		t.Errorf("expected 'Existing Corp', got %q", resp.Account.Name)
	}
}

func TestConvertLead_ExistingAccountByName(t *testing.T) {
	db := setupConvertDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Inc', 'Public', ?, ?)", now, now)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{
		"account": {"name": "Acme Inc"},
		"opportunity": {"name": "Deal"}
	}`

	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp convertLeadResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	// Should have found the existing account, not created a new one
	if resp.Account.ID != 1 {
		t.Errorf("expected to reuse existing account ID 1, got %d", resp.Account.ID)
	}
}

func TestConvertLead_CounterCaches(t *testing.T) {
	db := setupConvertDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, contacts_count, opportunities_count, created_at, updated_at) VALUES (1, 1, 0, 'Corp', 'Public', 0, 0, ?, ?)", now, now)
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, leads_count, opportunities_count, created_at, updated_at) VALUES (1, 1, 0, 'Q1', 'Public', 1, 0, ?, ?)", now, now)
	campaignID := int64(1)
	insertLead(db, 1, 1, &campaignID, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{
		"account": {"id": 1},
		"opportunity": {"name": "Deal"}
	}`

	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check account counter caches
	var acct model.Account
	db.First(&acct, 1)
	if acct.ContactsCount != 1 {
		t.Errorf("expected account contacts_count = 1, got %d", acct.ContactsCount)
	}
	if acct.OpportunitiesCount != 1 {
		t.Errorf("expected account opportunities_count = 1, got %d", acct.OpportunitiesCount)
	}

	// Check campaign opportunities_count incremented
	var campaign model.Campaign
	db.First(&campaign, 1)
	if campaign.OpportunitiesCount != 1 {
		t.Errorf("expected campaign opportunities_count = 1, got %d", campaign.OpportunitiesCount)
	}
}

func TestConvertLead_CopiesCampaignToOpportunity(t *testing.T) {
	db := setupConvertDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (5, 1, 0, 'Campaign', 'Public', ?, ?)", now, now)
	campaignID := int64(5)
	insertLead(db, 1, 1, &campaignID, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{
		"account": {"name": "Corp"},
		"opportunity": {"name": "Deal"}
	}`

	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp convertLeadResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Opportunity.CampaignID == nil || *resp.Opportunity.CampaignID != 5 {
		t.Error("expected opportunity to inherit lead's campaign_id")
	}
}

func TestConvertLead_AlreadyConverted(t *testing.T) {
	db := setupConvertDB(t)
	insertLead(db, 1, 1, nil, "converted")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {"name": "Corp"}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConvertLead_NotFound(t *testing.T) {
	db := setupConvertDB(t)
	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {"name": "Corp"}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/999/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestConvertLead_Forbidden(t *testing.T) {
	db := setupConvertDB(t)
	// Lead owned by user 99, accessed by non-admin user 5
	insertLead(db, 1, 99, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {"name": "Corp"}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+userToken(t, jwtSvc, 5))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestConvertLead_MissingAccountInfo(t *testing.T) {
	db := setupConvertDB(t)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConvertLead_MissingOpportunityName(t *testing.T) {
	db := setupConvertDB(t)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {"name": "Corp"}, "opportunity": {}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConvertLead_NonexistentAccountID(t *testing.T) {
	db := setupConvertDB(t)
	insertLead(db, 1, 1, nil, "new")

	mux, jwtSvc := entitiesRouter(t, db)

	body := `{"account": {"id": 999}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404 for non-existent account, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConvertLead_NoAuth(t *testing.T) {
	db := setupConvertDB(t)
	mux, _ := entitiesRouter(t, db)

	body := `{"account": {"name": "Corp"}, "opportunity": {"name": "Deal"}}`
	req := httptest.NewRequest("POST", "/api/v1/leads/1/convert", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
