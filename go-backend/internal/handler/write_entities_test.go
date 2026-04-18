package handler

import (
	"encoding/json"
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// ========== Account CRUD ==========

func TestCreateAccount_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{
		"name": "Acme Corp", "access": "Public",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var acct model.Account
	if err := json.NewDecoder(rec.Body).Decode(&acct); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if acct.Name != "Acme Corp" {
		t.Errorf("expected name 'Acme Corp', got %q", acct.Name)
	}
	if acct.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", acct.UserID)
	}
	if acct.Access != "Public" {
		t.Errorf("expected access 'Public', got %q", acct.Access)
	}
	if acct.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateAccount_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/accounts", makeToken("admin"), map[string]interface{}{
		"access": "Public",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateAccount_Unauthorized(t *testing.T) {
	mux, _, _ := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/accounts", "", map[string]interface{}{
		"name": "Acme Corp",
	})
	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUpdateAccount_Success(t *testing.T) {
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

func TestUpdateAccount_NotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "PUT", "/api/v1/accounts/9999", makeToken("admin"), map[string]interface{}{
		"name": "Updated",
	})
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateAccount_Forbidden(t *testing.T) {
	mux, jwtSvc, makeToken := writeRouter(t)

	// Create as admin (user 1)
	doRequest(mux, "POST", "/api/v1/accounts", makeToken("admin"), map[string]interface{}{"name": "Admin Account"})

	// Try to update as non-admin user 5
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	rec := doRequest(mux, "PUT", "/api/v1/accounts/1", tok5, map[string]interface{}{"name": "Hacked"})
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteAccount_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/accounts", tok, map[string]interface{}{"name": "Delete Me"})

	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify deleted
	rec2 := doRequest(mux, "GET", "/api/v1/accounts/1", tok, nil)
	if rec2.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestDeleteAccount_NotFound(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "DELETE", "/api/v1/accounts/9999", makeToken("admin"), nil)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteAccount_Forbidden(t *testing.T) {
	mux, jwtSvc, makeToken := writeRouter(t)

	// Create as admin (user 1)
	doRequest(mux, "POST", "/api/v1/accounts", makeToken("admin"), map[string]interface{}{"name": "Admin Account"})

	// Try to delete as non-admin user 5
	tok5, _ := jwtSvc.GenerateToken(5, "other", false)
	rec := doRequest(mux, "DELETE", "/api/v1/accounts/1", tok5, nil)
	if rec.Code != 403 {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Campaign CRUD ==========

func TestCreateCampaign_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/campaigns", tok, map[string]interface{}{
		"name": "Q1 Push", "status": "planned",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var campaign model.Campaign
	if err := json.NewDecoder(rec.Body).Decode(&campaign); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if campaign.Name != "Q1 Push" {
		t.Errorf("expected name 'Q1 Push', got %q", campaign.Name)
	}
	if campaign.Status == nil || *campaign.Status != "planned" {
		t.Errorf("expected status 'planned', got %v", campaign.Status)
	}
}

func TestCreateCampaign_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/campaigns", makeToken("admin"), map[string]interface{}{
		"status": "planned",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateCampaign_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/campaigns", tok, map[string]interface{}{
		"name": "Old Campaign", "status": "planned",
	})

	rec := doRequest(mux, "PUT", "/api/v1/campaigns/1", tok, map[string]interface{}{
		"name": "Updated Campaign", "status": "started",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var campaign model.Campaign
	if err := json.NewDecoder(rec.Body).Decode(&campaign); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if campaign.Name != "Updated Campaign" {
		t.Errorf("expected 'Updated Campaign', got %q", campaign.Name)
	}
	if campaign.Status == nil || *campaign.Status != "started" {
		t.Errorf("expected status 'started', got %v", campaign.Status)
	}
}

func TestDeleteCampaign_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/campaigns", tok, map[string]interface{}{"name": "Delete Me"})

	rec := doRequest(mux, "DELETE", "/api/v1/campaigns/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Lead CRUD ==========

func TestCreateLead_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "John", "last_name": "Doe", "company": "Acme",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var lead model.Lead
	if err := json.NewDecoder(rec.Body).Decode(&lead); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if lead.FirstName != "John" {
		t.Errorf("expected first_name 'John', got %q", lead.FirstName)
	}
	if lead.LastName != "Doe" {
		t.Errorf("expected last_name 'Doe', got %q", lead.LastName)
	}
	if lead.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", lead.UserID)
	}
}

func TestCreateLead_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/leads", makeToken("admin"), map[string]interface{}{
		"first_name": "John",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateLead_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "John", "last_name": "Doe",
	})

	rec := doRequest(mux, "PUT", "/api/v1/leads/1", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var lead model.Lead
	if err := json.NewDecoder(rec.Body).Decode(&lead); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if lead.FirstName != "Jane" {
		t.Errorf("expected first_name 'Jane', got %q", lead.FirstName)
	}
	if lead.LastName != "Smith" {
		t.Errorf("expected last_name 'Smith', got %q", lead.LastName)
	}
}

func TestDeleteLead_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "John", "last_name": "Doe",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/leads/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRejectLead_Success(t *testing.T) {
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

func TestRejectLead_AlreadyRejected(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/leads", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Doe",
	})

	// Reject once
	doRequest(mux, "PUT", "/api/v1/leads/1/reject", tok, nil)

	// Reject again — should still succeed (idempotent)
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

// ========== Contact CRUD ==========

func TestCreateContact_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith",
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var contact model.Contact
	if err := json.NewDecoder(rec.Body).Decode(&contact); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if contact.FirstName != "Jane" {
		t.Errorf("expected first_name 'Jane', got %q", contact.FirstName)
	}
	if contact.LastName != "Smith" {
		t.Errorf("expected last_name 'Smith', got %q", contact.LastName)
	}
	if contact.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", contact.UserID)
	}
}

func TestCreateContact_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/contacts", makeToken("admin"), map[string]interface{}{
		"first_name": "Jane",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateContact_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith",
	})

	rec := doRequest(mux, "PUT", "/api/v1/contacts/1", tok, map[string]interface{}{
		"first_name": "Janet", "last_name": "Jones",
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var contact model.Contact
	if err := json.NewDecoder(rec.Body).Decode(&contact); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if contact.FirstName != "Janet" {
		t.Errorf("expected first_name 'Janet', got %q", contact.FirstName)
	}
	if contact.LastName != "Jones" {
		t.Errorf("expected last_name 'Jones', got %q", contact.LastName)
	}
}

func TestDeleteContact_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/contacts", tok, map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/contacts/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ========== Opportunity CRUD ==========

func TestCreateOpportunity_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	rec := doRequest(mux, "POST", "/api/v1/opportunities", tok, map[string]interface{}{
		"name": "Big Deal", "stage": "prospecting", "amount": 50000, "probability": 25,
	})
	if rec.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var opp model.Opportunity
	if err := json.NewDecoder(rec.Body).Decode(&opp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if opp.Name != "Big Deal" {
		t.Errorf("expected name 'Big Deal', got %q", opp.Name)
	}
	if opp.Stage == nil || *opp.Stage != "prospecting" {
		t.Errorf("expected stage 'prospecting', got %v", opp.Stage)
	}
	if opp.UserID != 1 {
		t.Errorf("expected user_id 1, got %d", opp.UserID)
	}
}

func TestCreateOpportunity_MissingName(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	rec := doRequest(mux, "POST", "/api/v1/opportunities", makeToken("admin"), map[string]interface{}{
		"stage": "prospecting",
	})
	if rec.Code != 422 {
		t.Errorf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateOpportunity_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/opportunities", tok, map[string]interface{}{
		"name": "Deal", "stage": "prospecting",
	})

	rec := doRequest(mux, "PUT", "/api/v1/opportunities/1", tok, map[string]interface{}{
		"name": "Updated Deal", "stage": "negotiation", "probability": 75,
	})
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var opp model.Opportunity
	if err := json.NewDecoder(rec.Body).Decode(&opp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if opp.Name != "Updated Deal" {
		t.Errorf("expected name 'Updated Deal', got %q", opp.Name)
	}
	if opp.Stage == nil || *opp.Stage != "negotiation" {
		t.Errorf("expected stage 'negotiation', got %v", opp.Stage)
	}
}

func TestDeleteOpportunity_Success(t *testing.T) {
	mux, _, makeToken := writeRouter(t)
	tok := makeToken("admin")

	doRequest(mux, "POST", "/api/v1/opportunities", tok, map[string]interface{}{
		"name": "Delete Me", "stage": "prospecting",
	})

	rec := doRequest(mux, "DELETE", "/api/v1/opportunities/1", tok, nil)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
