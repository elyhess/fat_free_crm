package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)


func seedRelationships(db *gorm.DB) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Accounts
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Corp', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 0, 'Globex', 'Public', ?, ?)", now, now)

	// Contacts
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 0, 'Alice', 'Smith', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (2, 1, 0, 'Bob', 'Jones', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (3, 1, 0, 'Charlie', 'Brown', 'Public', ?, ?)", now, now)

	// Opportunities
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, created_at, updated_at) VALUES (1, 1, 0, 'Big Deal', 'Public', 'prospecting', 50000, ?, ?)", now, now)
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, created_at, updated_at) VALUES (2, 1, 0, 'Small Deal', 'Public', 'negotiation', 10000, ?, ?)", now, now)

	// Campaigns
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, created_at, updated_at) VALUES (1, 1, 0, 'Q1 Push', 'Public', 'active', ?, ?)", now, now)

	// Leads belonging to campaign 1
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, campaign_id, status, created_at, updated_at) VALUES (1, 1, 0, 'Lead', 'One', 'Public', 1, 'new', ?, ?)", now, now)
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, campaign_id, status, created_at, updated_at) VALUES (2, 1, 0, 'Lead', 'Two', 'Public', 1, 'new', ?, ?)", now, now)
	// Lead not linked to any campaign
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (3, 1, 0, 'Lead', 'Orphan', 'Public', 'new', ?, ?)", now, now)

	// Opportunities belonging to campaign 1
	db.Exec("UPDATE opportunities SET campaign_id = 1 WHERE id = 1")

	// Join table: account 1 has contacts 1 and 2
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at) VALUES (1, 1, 1, ?, ?)", now, now)
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at) VALUES (2, 1, 2, ?, ?)", now, now)

	// Join table: account 1 has opportunity 1
	db.Exec("INSERT INTO account_opportunities (id, account_id, opportunity_id, created_at, updated_at) VALUES (1, 1, 1, ?, ?)", now, now)

	// Join table: contact 1 has opportunities 1 and 2
	db.Exec("INSERT INTO contact_opportunities (id, contact_id, opportunity_id, role, created_at, updated_at) VALUES (1, 1, 1, 'Decision Maker', ?, ?)", now, now)
	db.Exec("INSERT INTO contact_opportunities (id, contact_id, opportunity_id, created_at, updated_at) VALUES (2, 1, 2, ?, ?)", now, now)
}

// --- Account Contacts ---

func TestAccountContacts(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/1/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 contacts for account 1, got %d", result.Total)
	}
}

func TestAccountContacts_Empty(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/2/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 contacts for account 2, got %d", result.Total)
	}
}

func TestAccountContacts_Pagination(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/1/contacts?per_page=1&page=1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Data) != 1 {
		t.Errorf("expected 1 item per page, got %d", len(result.Data))
	}
	if result.TotalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", result.TotalPages)
	}
}

func TestAccountContacts_AccessControl(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (10, 1, 0, 'Test Acct', 'Public', ?, ?)", now, now)
	// Private contact owned by user 99 — should NOT be visible to user 5
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (10, 99, 0, 'Secret', 'Person', 'Private', ?, ?)", now, now)
	// Public contact — visible to all
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (11, 1, 0, 'Public', 'Person', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at) VALUES (10, 10, 10, ?, ?)", now, now)
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at) VALUES (11, 10, 11, ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts/10/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t, jwtSvc, 5))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Only the public contact should be visible
	if result.Total != 1 {
		t.Errorf("expected 1 accessible contact, got %d", result.Total)
	}
}

// --- Account Opportunities ---

func TestAccountOpportunities(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/1/opportunities", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Opportunity]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 opportunity for account 1, got %d", result.Total)
	}
}

// --- Campaign Leads ---

func TestCampaignLeads(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/campaigns/1/leads", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Lead]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 leads for campaign 1, got %d", result.Total)
	}
}

func TestCampaignLeads_Empty(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	// Campaign 1 exists but use a non-existent campaign ID
	req := httptest.NewRequest("GET", "/api/v1/campaigns/999/leads", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Lead]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 leads, got %d", result.Total)
	}
}

// --- Campaign Opportunities ---

func TestCampaignOpportunities(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/campaigns/1/opportunities", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Opportunity]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 opportunity for campaign 1, got %d", result.Total)
	}
}

// --- Contact Opportunities ---

func TestContactOpportunities(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/contacts/1/opportunities", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Opportunity]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 opportunities for contact 1, got %d", result.Total)
	}
}

func TestContactOpportunities_Empty(t *testing.T) {
	db := testDB(t)
	seedRelationships(db)
	mux, jwtSvc := entitiesRouter(t, db)

	// Contact 3 has no opportunities
	req := httptest.NewRequest("GET", "/api/v1/contacts/3/opportunities", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Opportunity]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 opportunities for contact 3, got %d", result.Total)
	}
}

// --- Auth ---

func TestRelationships_NoToken(t *testing.T) {
	db := testDB(t)
	mux, _ := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/1/contacts", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRelationships_InvalidID(t *testing.T) {
	db := testDB(t)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/abc/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Soft-deleted join records are excluded ---

func TestAccountContacts_SoftDeletedJoinExcluded(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (20, 1, 0, 'Test', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (20, 1, 0, 'Active', 'Contact', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (21, 1, 0, 'Deleted', 'Contact', 'Public', ?, ?)", now, now)
	// Active join
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at) VALUES (20, 20, 20, ?, ?)", now, now)
	// Soft-deleted join
	db.Exec("INSERT INTO account_contacts (id, account_id, contact_id, created_at, updated_at, deleted_at) VALUES (21, 20, 21, ?, ?, ?)", now, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts/20/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 (soft-deleted join excluded), got %d", result.Total)
	}
}
