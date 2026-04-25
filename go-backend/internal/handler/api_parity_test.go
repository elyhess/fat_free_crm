package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Helpers
// ============================================================================

// decodeJSON decodes the response body into a map using json.Number for numeric
// precision, which lets us distinguish integers from floats.
func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	dec := json.NewDecoder(rec.Body)
	dec.UseNumber()
	var m map[string]interface{}
	if err := dec.Decode(&m); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, rec.Body.String())
	}
	return m
}

// decodeJSONList decodes a paginated list response, returning the envelope map
// and the data array of maps.
func decodeJSONList(t *testing.T, rec *httptest.ResponseRecorder) (map[string]interface{}, []map[string]interface{}) {
	t.Helper()
	dec := json.NewDecoder(rec.Body)
	dec.UseNumber()
	var envelope map[string]interface{}
	if err := dec.Decode(&envelope); err != nil {
		t.Fatalf("decode JSON list: %v\nbody: %s", err, rec.Body.String())
	}

	rawData, ok := envelope["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' to be an array, got %T", envelope["data"])
	}
	var items []map[string]interface{}
	for _, v := range rawData {
		item, ok := v.(map[string]interface{})
		if !ok {
			t.Fatalf("expected data item to be a map, got %T", v)
		}
		items = append(items, item)
	}
	return envelope, items
}

// assertFieldPresent checks that a field exists in the map (even if null).
func assertFieldPresent(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	if _, ok := m[field]; !ok {
		t.Errorf("%s: expected field %q to be present in response, but it was missing", entity, field)
	}
}

// assertFieldAbsent checks that a field does NOT exist in the map.
func assertFieldAbsent(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	if _, ok := m[field]; ok {
		t.Errorf("%s: unexpected field %q found in response", entity, field)
	}
}

// assertJSONNumber checks that a field is a json.Number.
func assertJSONNumber(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		t.Errorf("%s: field %q missing (expected json.Number)", entity, field)
		return
	}
	if v == nil {
		return // null is acceptable for nullable numeric fields
	}
	if _, ok := v.(json.Number); !ok {
		t.Errorf("%s: field %q expected json.Number, got %T (%v)", entity, field, v, v)
	}
}

// assertJSONBool checks that a field is a boolean.
func assertJSONBool(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		t.Errorf("%s: field %q missing (expected bool)", entity, field)
		return
	}
	if _, ok := v.(bool); !ok {
		t.Errorf("%s: field %q expected bool, got %T (%v)", entity, field, v, v)
	}
}

// assertJSONString checks that a field is a string.
func assertJSONString(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		t.Errorf("%s: field %q missing (expected string)", entity, field)
		return
	}
	if v == nil {
		return // null is acceptable for nullable string fields
	}
	if _, ok := v.(string); !ok {
		t.Errorf("%s: field %q expected string, got %T (%v)", entity, field, v, v)
	}
}

// assertJSONNull checks that a field is present and null.
func assertJSONNull(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		// Field may be omitted via omitempty — that's also a form of "null"
		return
	}
	if v != nil {
		t.Errorf("%s: field %q expected null, got %T (%v)", entity, field, v, v)
	}
}

// assertTimestampFormat checks a string field matches RFC3339/ISO 8601 format.
func assertTimestampFormat(t *testing.T, m map[string]interface{}, entity, field string) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		t.Errorf("%s: field %q missing (expected timestamp string)", entity, field)
		return
	}
	s, ok := v.(string)
	if !ok {
		t.Errorf("%s: field %q expected string timestamp, got %T", entity, field, v)
		return
	}
	// Try parsing as RFC3339 (ISO 8601 with timezone)
	if _, err := time.Parse(time.RFC3339, s); err != nil {
		// Also try RFC3339Nano
		if _, err2 := time.Parse(time.RFC3339Nano, s); err2 != nil {
			t.Errorf("%s: field %q not valid RFC3339: %q (err: %v)", entity, field, s, err)
		}
	}
}

func doGet(t *testing.T, mux *http.ServeMux, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// ============================================================================
// 1. Response Field Completeness & 2. Data Type Verification
// ============================================================================

func TestAPIParity_AccountFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts
		(id, user_id, assigned_to, name, access, website, toll_free_phone, phone,
		 fax, email, background_info, rating, category,
		 contacts_count, opportunities_count, created_at, updated_at)
		VALUES (1, 1, 2, 'Acme Corp', 'Public', 'https://acme.com', '1-800-ACME',
		        '555-0100', '555-0101', 'acme@example.com', 'A large company', 5,
		        'customer', 10, 3, ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	// Required fields that must always be present
	requiredFields := []string{
		"id", "user_id", "assigned_to", "name", "access",
		"rating", "contacts_count", "opportunities_count",
		"created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Account", f)
	}

	// Optional fields that should be present when populated
	optionalFields := []string{
		"website", "toll_free_phone", "phone", "fax", "email",
		"background_info", "category",
	}
	for _, f := range optionalFields {
		assertFieldPresent(t, m, "Account", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Account", "id")
	assertJSONNumber(t, m, "Account", "user_id")
	assertJSONNumber(t, m, "Account", "assigned_to")
	assertJSONString(t, m, "Account", "name")
	assertJSONString(t, m, "Account", "access")
	assertJSONNumber(t, m, "Account", "rating")
	assertJSONNumber(t, m, "Account", "contacts_count")
	assertJSONNumber(t, m, "Account", "opportunities_count")
	assertJSONString(t, m, "Account", "website")
	assertJSONString(t, m, "Account", "toll_free_phone")
	assertJSONString(t, m, "Account", "phone")
	assertJSONString(t, m, "Account", "fax")
	assertJSONString(t, m, "Account", "email")
	assertJSONString(t, m, "Account", "background_info")
	assertJSONString(t, m, "Account", "category")
	assertTimestampFormat(t, m, "Account", "created_at")
	assertTimestampFormat(t, m, "Account", "updated_at")

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "assigned_to": true, "name": true,
		"access": true, "website": true, "toll_free_phone": true, "phone": true,
		"fax": true, "email": true, "background_info": true, "rating": true,
		"category": true, "contacts_count": true, "opportunities_count": true,
		"created_at": true, "updated_at": true, "deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Account: unexpected field %q in response", key)
		}
	}
}

func TestAPIParity_AccountFieldsNullOptionals(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Insert with only required fields — optional fields should be null/omitted
	db.Exec(`INSERT INTO accounts
		(id, user_id, assigned_to, name, access, rating, contacts_count,
		 opportunities_count, created_at, updated_at)
		VALUES (1, 1, 0, 'Minimal Corp', 'Public', 0, 0, 0, ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	// Nullable fields should be null or omitted (omitempty), NOT empty strings
	nullableFields := []string{"website", "toll_free_phone", "phone", "fax", "email", "background_info", "category"}
	for _, f := range nullableFields {
		if v, ok := m[f]; ok && v != nil {
			if s, isStr := v.(string); isStr && s == "" {
				t.Errorf("Account: nullable field %q should be null or omitted, not empty string", f)
			}
		}
	}
}

func TestAPIParity_ContactFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	bornOn := "1985-06-15"

	db.Exec(`INSERT INTO contacts
		(id, user_id, lead_id, assigned_to, reports_to, first_name, last_name,
		 access, title, department, email, alt_email, phone, mobile, fax,
		 blog, linkedin, facebook, twitter, born_on, do_not_call,
		 background_info, created_at, updated_at)
		VALUES (1, 1, NULL, 2, 3, 'Jane', 'Doe', 'Public', 'VP Sales',
		        'Sales', 'jane@example.com', 'jane2@example.com', '555-0100',
		        '555-0101', '555-0102', 'https://blog.jane.com',
		        'https://linkedin.com/jane', 'https://facebook.com/jane',
		        '@janedoe', ?, true, 'Important contact', ?, ?)`, bornOn, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/contacts/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	requiredFields := []string{
		"id", "user_id", "assigned_to", "first_name", "last_name",
		"access", "do_not_call", "created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Contact", f)
	}

	populatedOptionals := []string{
		"reports_to", "title", "department", "email", "alt_email",
		"phone", "mobile", "fax", "blog", "linkedin", "facebook",
		"twitter", "born_on", "background_info",
	}
	for _, f := range populatedOptionals {
		assertFieldPresent(t, m, "Contact", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Contact", "id")
	assertJSONNumber(t, m, "Contact", "user_id")
	assertJSONNumber(t, m, "Contact", "assigned_to")
	assertJSONNumber(t, m, "Contact", "reports_to")
	assertJSONString(t, m, "Contact", "first_name")
	assertJSONString(t, m, "Contact", "last_name")
	assertJSONString(t, m, "Contact", "access")
	assertJSONString(t, m, "Contact", "title")
	assertJSONString(t, m, "Contact", "department")
	assertJSONString(t, m, "Contact", "email")
	assertJSONString(t, m, "Contact", "alt_email")
	assertJSONString(t, m, "Contact", "phone")
	assertJSONString(t, m, "Contact", "mobile")
	assertJSONString(t, m, "Contact", "fax")
	assertJSONString(t, m, "Contact", "blog")
	assertJSONString(t, m, "Contact", "linkedin")
	assertJSONString(t, m, "Contact", "facebook")
	assertJSONString(t, m, "Contact", "twitter")
	assertJSONBool(t, m, "Contact", "do_not_call")
	assertJSONString(t, m, "Contact", "background_info")
	assertTimestampFormat(t, m, "Contact", "created_at")
	assertTimestampFormat(t, m, "Contact", "updated_at")

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "lead_id": true, "assigned_to": true,
		"reports_to": true, "first_name": true, "last_name": true, "access": true,
		"title": true, "department": true, "email": true, "alt_email": true,
		"phone": true, "mobile": true, "fax": true, "blog": true, "linkedin": true,
		"facebook": true, "twitter": true, "born_on": true, "do_not_call": true,
		"background_info": true, "created_at": true, "updated_at": true,
		"deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Contact: unexpected field %q in response", key)
		}
	}
}

func TestAPIParity_LeadFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO leads
		(id, user_id, campaign_id, assigned_to, first_name, last_name, access,
		 company, title, source, status, referred_by, email, alt_email,
		 phone, mobile, blog, linkedin, facebook, twitter, rating, do_not_call,
		 background_info, created_at, updated_at)
		VALUES (1, 1, NULL, 2, 'Bob', 'Builder', 'Public',
		        'BuildCo', 'CEO', 'web', 'new', 'Alice', 'bob@example.com',
		        'bob2@example.com', '555-0200', '555-0201', 'https://blog.bob.com',
		        'https://linkedin.com/bob', 'https://facebook.com/bob', '@bob',
		        3, false, 'Potential lead', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/leads/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	requiredFields := []string{
		"id", "user_id", "assigned_to", "first_name", "last_name",
		"access", "rating", "do_not_call", "created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Lead", f)
	}

	populatedOptionals := []string{
		"company", "title", "source", "status", "referred_by",
		"email", "alt_email", "phone", "mobile", "blog",
		"linkedin", "facebook", "twitter", "background_info",
	}
	for _, f := range populatedOptionals {
		assertFieldPresent(t, m, "Lead", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Lead", "id")
	assertJSONNumber(t, m, "Lead", "user_id")
	assertJSONNumber(t, m, "Lead", "assigned_to")
	assertJSONString(t, m, "Lead", "first_name")
	assertJSONString(t, m, "Lead", "last_name")
	assertJSONString(t, m, "Lead", "access")
	assertJSONString(t, m, "Lead", "company")
	assertJSONString(t, m, "Lead", "title")
	assertJSONString(t, m, "Lead", "source")
	assertJSONString(t, m, "Lead", "status")
	assertJSONString(t, m, "Lead", "referred_by")
	assertJSONString(t, m, "Lead", "email")
	assertJSONString(t, m, "Lead", "alt_email")
	assertJSONString(t, m, "Lead", "phone")
	assertJSONString(t, m, "Lead", "mobile")
	assertJSONString(t, m, "Lead", "blog")
	assertJSONString(t, m, "Lead", "linkedin")
	assertJSONString(t, m, "Lead", "facebook")
	assertJSONString(t, m, "Lead", "twitter")
	assertJSONNumber(t, m, "Lead", "rating")
	assertJSONBool(t, m, "Lead", "do_not_call")
	assertJSONString(t, m, "Lead", "background_info")
	assertTimestampFormat(t, m, "Lead", "created_at")
	assertTimestampFormat(t, m, "Lead", "updated_at")

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "campaign_id": true, "assigned_to": true,
		"first_name": true, "last_name": true, "access": true, "company": true,
		"title": true, "source": true, "status": true, "referred_by": true,
		"email": true, "alt_email": true, "phone": true, "mobile": true,
		"blog": true, "linkedin": true, "facebook": true, "twitter": true,
		"rating": true, "do_not_call": true, "background_info": true,
		"created_at": true, "updated_at": true, "deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Lead: unexpected field %q in response", key)
		}
	}
}

func TestAPIParity_OpportunityFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	closesOn := "2026-12-31"

	db.Exec(`INSERT INTO opportunities
		(id, user_id, campaign_id, assigned_to, name, access, source, stage,
		 probability, amount, discount, closes_on, background_info,
		 created_at, updated_at)
		VALUES (1, 1, NULL, 2, 'Big Deal', 'Public', 'web', 'prospecting',
		        75, 100000.50, 5000.25, ?, 'High-value opportunity', ?, ?)`,
		closesOn, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/opportunities/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	requiredFields := []string{
		"id", "user_id", "assigned_to", "name", "access",
		"created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Opportunity", f)
	}

	populatedOptionals := []string{
		"source", "stage", "probability", "amount", "discount",
		"closes_on", "background_info",
	}
	for _, f := range populatedOptionals {
		assertFieldPresent(t, m, "Opportunity", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Opportunity", "id")
	assertJSONNumber(t, m, "Opportunity", "user_id")
	assertJSONNumber(t, m, "Opportunity", "assigned_to")
	assertJSONString(t, m, "Opportunity", "name")
	assertJSONString(t, m, "Opportunity", "access")
	assertJSONString(t, m, "Opportunity", "source")
	assertJSONString(t, m, "Opportunity", "stage")
	assertJSONNumber(t, m, "Opportunity", "probability")
	assertJSONNumber(t, m, "Opportunity", "amount")
	assertJSONNumber(t, m, "Opportunity", "discount")
	assertJSONString(t, m, "Opportunity", "background_info")
	assertTimestampFormat(t, m, "Opportunity", "created_at")
	assertTimestampFormat(t, m, "Opportunity", "updated_at")

	// Verify amount is a proper number (not a string)
	if amt, ok := m["amount"].(json.Number); ok {
		if _, err := amt.Float64(); err != nil {
			t.Errorf("Opportunity: amount %q is not a valid float64", amt)
		}
	}

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "campaign_id": true, "assigned_to": true,
		"name": true, "access": true, "source": true, "stage": true,
		"probability": true, "amount": true, "discount": true,
		"closes_on": true, "background_info": true,
		"created_at": true, "updated_at": true, "deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Opportunity: unexpected field %q in response", key)
		}
	}
}

func TestAPIParity_CampaignFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	startsOn := "2026-01-01"
	endsOn := "2026-06-30"

	db.Exec(`INSERT INTO campaigns
		(id, user_id, assigned_to, name, access, status, budget,
		 target_leads, target_conversion, target_revenue,
		 leads_count, opportunities_count, revenue,
		 starts_on, ends_on, objectives, background_info,
		 created_at, updated_at)
		VALUES (1, 1, 2, 'Q1 Push', 'Public', 'active', 50000.00,
		        100, 15.5, 250000.00,
		        42, 8, 125000.00,
		        ?, ?, 'Increase market share', 'Major campaign',
		        ?, ?)`, startsOn, endsOn, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/campaigns/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	requiredFields := []string{
		"id", "user_id", "assigned_to", "name", "access",
		"leads_count", "opportunities_count",
		"created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Campaign", f)
	}

	populatedOptionals := []string{
		"status", "budget", "target_leads", "target_conversion",
		"target_revenue", "revenue", "starts_on", "ends_on",
		"objectives", "background_info",
	}
	for _, f := range populatedOptionals {
		assertFieldPresent(t, m, "Campaign", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Campaign", "id")
	assertJSONNumber(t, m, "Campaign", "user_id")
	assertJSONNumber(t, m, "Campaign", "assigned_to")
	assertJSONString(t, m, "Campaign", "name")
	assertJSONString(t, m, "Campaign", "access")
	assertJSONString(t, m, "Campaign", "status")
	assertJSONNumber(t, m, "Campaign", "budget")
	assertJSONNumber(t, m, "Campaign", "target_leads")
	assertJSONNumber(t, m, "Campaign", "target_conversion")
	assertJSONNumber(t, m, "Campaign", "target_revenue")
	assertJSONNumber(t, m, "Campaign", "leads_count")
	assertJSONNumber(t, m, "Campaign", "opportunities_count")
	assertJSONNumber(t, m, "Campaign", "revenue")
	assertJSONString(t, m, "Campaign", "objectives")
	assertJSONString(t, m, "Campaign", "background_info")
	assertTimestampFormat(t, m, "Campaign", "created_at")
	assertTimestampFormat(t, m, "Campaign", "updated_at")

	// Verify count fields are integers
	for _, f := range []string{"leads_count", "opportunities_count"} {
		if n, ok := m[f].(json.Number); ok {
			if _, err := n.Int64(); err != nil {
				t.Errorf("Campaign: %q should be an integer, got %q", f, n)
			}
		}
	}

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "assigned_to": true, "name": true,
		"access": true, "status": true, "budget": true, "target_leads": true,
		"target_conversion": true, "target_revenue": true, "leads_count": true,
		"opportunities_count": true, "revenue": true, "starts_on": true,
		"ends_on": true, "objectives": true, "background_info": true,
		"created_at": true, "updated_at": true, "deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Campaign: unexpected field %q in response", key)
		}
	}
}

func TestAPIParity_TaskFields(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	dueAt := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	completedAt := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO tasks
		(id, user_id, assigned_to, completed_by, name, asset_id, asset_type,
		 priority, category, bucket, due_at, completed_at, background_info,
		 created_at, updated_at)
		VALUES (1, 1, 2, 3, 'Call Bob', 10, 'Contact',
		        'high', 'call', 'due_today', ?, ?, 'Follow up on deal',
		        ?, ?)`, dueAt, completedAt, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/tasks/1", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	m := decodeJSON(t, rec)

	requiredFields := []string{
		"id", "user_id", "assigned_to", "name",
		"created_at", "updated_at",
	}
	for _, f := range requiredFields {
		assertFieldPresent(t, m, "Task", f)
	}

	populatedOptionals := []string{
		"completed_by", "asset_id", "asset_type", "priority",
		"category", "bucket", "due_at", "completed_at", "background_info",
	}
	for _, f := range populatedOptionals {
		assertFieldPresent(t, m, "Task", f)
	}

	// Type checks
	assertJSONNumber(t, m, "Task", "id")
	assertJSONNumber(t, m, "Task", "user_id")
	assertJSONNumber(t, m, "Task", "assigned_to")
	assertJSONNumber(t, m, "Task", "completed_by")
	assertJSONString(t, m, "Task", "name")
	assertJSONNumber(t, m, "Task", "asset_id")
	assertJSONString(t, m, "Task", "asset_type")
	assertJSONString(t, m, "Task", "priority")
	assertJSONString(t, m, "Task", "category")
	assertJSONString(t, m, "Task", "bucket")
	assertJSONString(t, m, "Task", "background_info")
	assertTimestampFormat(t, m, "Task", "created_at")
	assertTimestampFormat(t, m, "Task", "updated_at")

	// Verify no unexpected fields
	expectedKeys := map[string]bool{
		"id": true, "user_id": true, "assigned_to": true, "completed_by": true,
		"name": true, "asset_id": true, "asset_type": true, "priority": true,
		"category": true, "bucket": true, "due_at": true, "completed_at": true,
		"background_info": true, "created_at": true, "updated_at": true,
		"deleted_at": true,
	}
	for key := range m {
		if !expectedKeys[key] {
			t.Errorf("Task: unexpected field %q in response", key)
		}
	}
}

// ============================================================================
// 3. Pagination Shape
// ============================================================================

func TestAPIParity_PaginationShape(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Insert 45 accounts to test pagination math
	for i := 1; i <= 45; i++ {
		db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
			VALUES (?, 1, 0, ?, 'Public', ?, ?)`, i, fmt.Sprintf("Company %02d", i), now, now)
	}

	mux, jwtSvc := entitiesRouter(t, db)

	// Test default pagination
	rec := doGet(t, mux, "/api/v1/accounts", adminToken(t, jwtSvc))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	envelope, items := decodeJSONList(t, rec)

	// Verify all pagination fields exist
	for _, f := range []string{"data", "total", "page", "per_page", "total_pages"} {
		assertFieldPresent(t, envelope, "Pagination", f)
	}

	// Verify data is an array
	if _, ok := envelope["data"].([]interface{}); !ok {
		t.Error("Pagination: 'data' should be an array")
	}

	// Verify pagination metadata types
	assertJSONNumber(t, envelope, "Pagination", "total")
	assertJSONNumber(t, envelope, "Pagination", "page")
	assertJSONNumber(t, envelope, "Pagination", "per_page")
	assertJSONNumber(t, envelope, "Pagination", "total_pages")

	// Verify default page is 1
	if page, ok := envelope["page"].(json.Number); ok {
		if p, _ := page.Int64(); p != 1 {
			t.Errorf("Pagination: default page should be 1, got %d", p)
		}
	}

	// Verify per_page is between 20-25 (default)
	if perPage, ok := envelope["per_page"].(json.Number); ok {
		pp, _ := perPage.Int64()
		if pp < 20 || pp > 25 {
			t.Errorf("Pagination: default per_page should be 20-25, got %d", pp)
		}
	}

	// Verify total = 45
	if total, ok := envelope["total"].(json.Number); ok {
		if tot, _ := total.Int64(); tot != 45 {
			t.Errorf("Pagination: total should be 45, got %d", tot)
		}
	}

	// Verify total_pages calculation: ceil(45/20) = 3
	if tp, ok := envelope["total_pages"].(json.Number); ok {
		totalPages, _ := tp.Int64()
		perPage, _ := envelope["per_page"].(json.Number)
		pp, _ := perPage.Int64()
		expected := (int64(45) + pp - 1) / pp
		if totalPages != expected {
			t.Errorf("Pagination: total_pages should be %d (ceil(45/%d)), got %d", expected, pp, totalPages)
		}
	}

	// Verify the number of items matches per_page for a non-last page
	perPage, _ := envelope["per_page"].(json.Number)
	pp, _ := perPage.Int64()
	if int64(len(items)) != pp {
		t.Errorf("Pagination: first page should have %d items, got %d", pp, len(items))
	}
}

func TestAPIParity_PaginationLastPage(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Insert 5 accounts with per_page=2 -> 3 pages, last page has 1 item
	for i := 1; i <= 5; i++ {
		db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
			VALUES (?, 1, 0, ?, 'Public', ?, ?)`, i, fmt.Sprintf("Co %d", i), now, now)
	}

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts?page=3&per_page=2", adminToken(t, jwtSvc))

	envelope, items := decodeJSONList(t, rec)

	if tp, ok := envelope["total_pages"].(json.Number); ok {
		if v, _ := tp.Int64(); v != 3 {
			t.Errorf("expected 3 total_pages, got %d", v)
		}
	}
	if page, ok := envelope["page"].(json.Number); ok {
		if v, _ := page.Int64(); v != 3 {
			t.Errorf("expected page 3, got %d", v)
		}
	}
	if len(items) != 1 {
		t.Errorf("last page should have 1 item, got %d", len(items))
	}
}

// ============================================================================
// 4. Sort Order Parity
// ============================================================================

func TestAPIParity_DefaultSortOrder(t *testing.T) {
	db := testDB(t)

	// Insert 3 accounts at different times to test sort order
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Alpha Corp', 'Public', '2026-01-01 10:00:00', '2026-01-01 10:00:00')`)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (2, 1, 0, 'Beta Inc', 'Public', '2026-01-02 10:00:00', '2026-01-02 10:00:00')`)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (3, 1, 0, 'Gamma LLC', 'Public', '2026-01-03 10:00:00', '2026-01-03 10:00:00')`)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts", adminToken(t, jwtSvc))

	_, items := decodeJSONList(t, rec)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Default sort is by id desc (based on DefaultPagination: Sort="id", Order="desc")
	// So the order should be: Gamma (id=3), Beta (id=2), Alpha (id=1)
	expectedOrder := []string{"Gamma LLC", "Beta Inc", "Alpha Corp"}
	for i, expected := range expectedOrder {
		name, _ := items[i]["name"].(string)
		if name != expected {
			t.Errorf("sort order position %d: expected %q, got %q", i, expected, name)
		}
	}
}

func TestAPIParity_ExplicitSortByName(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Charlie', 'Public', ?, ?)`, now, now)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (2, 1, 0, 'Alpha', 'Public', ?, ?)`, now, now)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (3, 1, 0, 'Bravo', 'Public', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts?sort=name&order=asc", adminToken(t, jwtSvc))

	_, items := decodeJSONList(t, rec)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	expectedOrder := []string{"Alpha", "Bravo", "Charlie"}
	for i, expected := range expectedOrder {
		name, _ := items[i]["name"].(string)
		if name != expected {
			t.Errorf("sort by name asc position %d: expected %q, got %q", i, expected, name)
		}
	}
}

// ============================================================================
// 5. Soft Delete Behavior
// ============================================================================

func TestAPIParity_SoftDeleteExcludedFromList(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Active Account', 'Public', ?, ?)`, now, now)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at, deleted_at)
		VALUES (2, 1, 0, 'Deleted Account', 'Public', ?, ?, ?)`, now, now, now)
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (3, 1, 0, 'Another Active', 'Public', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts", adminToken(t, jwtSvc))

	envelope, items := decodeJSONList(t, rec)

	// Total should be 2, not 3
	if total, ok := envelope["total"].(json.Number); ok {
		if v, _ := total.Int64(); v != 2 {
			t.Errorf("expected total=2 (soft-deleted excluded), got %d", v)
		}
	}

	// Verify the deleted account name is not in results
	for _, item := range items {
		name, _ := item["name"].(string)
		if name == "Deleted Account" {
			t.Error("soft-deleted record should not appear in list results")
		}
	}
}

func TestAPIParity_SoftDeleteGetReturns404(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at, deleted_at)
		VALUES (1, 1, 0, 'Deleted Account', 'Public', ?, ?, ?)`, now, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts/1", adminToken(t, jwtSvc))

	if rec.Code != 404 {
		t.Errorf("GET on soft-deleted record should return 404, got %d", rec.Code)
	}
}

func TestAPIParity_SoftDeleteAllEntities(t *testing.T) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Test soft-delete filtering across all entity types
	entities := []struct {
		table  string
		path   string
		insert string
	}{
		{
			table: "contacts",
			path:  "/api/v1/contacts",
			insert: fmt.Sprintf(`INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at, deleted_at)
				VALUES (1, 1, 0, 'Del', 'Contact', 'Public', '%s', '%s', '%s')`, now, now, now),
		},
		{
			table: "leads",
			path:  "/api/v1/leads",
			insert: fmt.Sprintf(`INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at, deleted_at)
				VALUES (1, 1, 0, 'Del', 'Lead', 'Public', 'new', '%s', '%s', '%s')`, now, now, now),
		},
		{
			table: "opportunities",
			path:  "/api/v1/opportunities",
			insert: fmt.Sprintf(`INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at, deleted_at)
				VALUES (1, 1, 0, 'Del Opp', 'Public', 'lost', 0, 0, '%s', '%s', '%s')`, now, now, now),
		},
		{
			table: "campaigns",
			path:  "/api/v1/campaigns",
			insert: fmt.Sprintf(`INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, created_at, updated_at, deleted_at)
				VALUES (1, 1, 0, 'Del Camp', 'Public', 'active', '%s', '%s', '%s')`, now, now, now),
		},
		{
			table: "tasks",
			path:  "/api/v1/tasks",
			insert: fmt.Sprintf(`INSERT INTO tasks (id, user_id, assigned_to, name, bucket, priority, created_at, updated_at, deleted_at)
				VALUES (1, 1, 0, 'Del Task', 'due_today', 'high', '%s', '%s', '%s')`, now, now, now),
		},
	}

	for _, e := range entities {
		t.Run(e.table, func(t *testing.T) {
			db := testDB(t)
			db.Exec(e.insert)

			mux, jwtSvc := entitiesRouter(t, db)

			// List should be empty
			rec := doGet(t, mux, e.path, adminToken(t, jwtSvc))
			envelope, _ := decodeJSONList(t, rec)
			if total, ok := envelope["total"].(json.Number); ok {
				if v, _ := total.Int64(); v != 0 {
					t.Errorf("%s: soft-deleted should be excluded from list, got total=%d", e.table, v)
				}
			}

			// GET by ID should return 404
			rec = doGet(t, mux, e.path+"/1", adminToken(t, jwtSvc))
			if rec.Code != 404 {
				t.Errorf("%s: GET soft-deleted should return 404, got %d", e.table, rec.Code)
			}
		})
	}
}

// ============================================================================
// 6. Access Control Filtering
// ============================================================================

func TestAPIParity_AccessControlPrivateNotVisibleToOthers(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Private account owned by user 1
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Secret Corp', 'Private', ?, ?)`, now, now)
	// Public account for comparison
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (2, 1, 0, 'Public Corp', 'Public', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// List as user 5 (non-admin) — should NOT see the private account
	rec := doGet(t, mux, "/api/v1/accounts", userToken(t, jwtSvc, 5))
	_, items := decodeJSONList(t, rec)

	for _, item := range items {
		name, _ := item["name"].(string)
		if name == "Secret Corp" {
			t.Error("non-owner should NOT see Private account in list")
		}
	}

	// Verify we do see the public one
	found := false
	for _, item := range items {
		name, _ := item["name"].(string)
		if name == "Public Corp" {
			found = true
		}
	}
	if !found {
		t.Error("non-owner should see Public account in list")
	}
}

func TestAPIParity_AccessControlPrivateVisibleToOwner(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Private account owned by user 1
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Secret Corp', 'Private', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// List as admin (user 1, the owner) — should see it
	rec := doGet(t, mux, "/api/v1/accounts", adminToken(t, jwtSvc))
	_, items := decodeJSONList(t, rec)

	found := false
	for _, item := range items {
		name, _ := item["name"].(string)
		if name == "Secret Corp" {
			found = true
		}
	}
	if !found {
		t.Error("owner/admin should see Private account in list")
	}
}

func TestAPIParity_AccessControlPrivateGetReturns404ForNonOwner(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Secret Corp', 'Private', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// GET as user 5 — should be 404 (filtered by access scope)
	rec := doGet(t, mux, "/api/v1/accounts/1", userToken(t, jwtSvc, 5))
	if rec.Code != 404 {
		t.Errorf("non-owner GET of Private record should return 404, got %d", rec.Code)
	}
}

func TestAPIParity_AccessControlAssignedToCanSeePrivate(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Private account owned by user 1 but assigned to user 5
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 5, 'Assigned Corp', 'Private', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// User 5 is the assignee — should be able to see it
	rec := doGet(t, mux, "/api/v1/accounts", userToken(t, jwtSvc, 5))
	_, items := decodeJSONList(t, rec)

	found := false
	for _, item := range items {
		name, _ := item["name"].(string)
		if name == "Assigned Corp" {
			found = true
		}
	}
	if !found {
		t.Error("assigned user should see Private account they are assigned to")
	}
}

func TestAPIParity_AccessControlAllEntities(t *testing.T) {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Verify access control works for all access-controlled entities
	entities := []struct {
		name   string
		path   string
		insert string
	}{
		{
			name: "contacts",
			path: "/api/v1/contacts",
			insert: fmt.Sprintf(`INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at)
				VALUES (1, 99, 0, 'Private', 'Contact', 'Private', '%s', '%s')`, now, now),
		},
		{
			name: "leads",
			path: "/api/v1/leads",
			insert: fmt.Sprintf(`INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at)
				VALUES (1, 99, 0, 'Private', 'Lead', 'Private', 'new', '%s', '%s')`, now, now),
		},
		{
			name: "opportunities",
			path: "/api/v1/opportunities",
			insert: fmt.Sprintf(`INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at)
				VALUES (1, 99, 0, 'Private Opp', 'Private', 'prospecting', 0, 0, '%s', '%s')`, now, now),
		},
		{
			name: "campaigns",
			path: "/api/v1/campaigns",
			insert: fmt.Sprintf(`INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, created_at, updated_at)
				VALUES (1, 99, 0, 'Private Camp', 'Private', 'active', '%s', '%s')`, now, now),
		},
	}

	for _, e := range entities {
		t.Run(e.name, func(t *testing.T) {
			db := testDB(t)
			db.Exec(e.insert)

			mux, jwtSvc := entitiesRouter(t, db)

			// Non-owner, non-admin should not see private records
			rec := doGet(t, mux, e.path, userToken(t, jwtSvc, 5))
			envelope, _ := decodeJSONList(t, rec)
			if total, ok := envelope["total"].(json.Number); ok {
				if v, _ := total.Int64(); v != 0 {
					t.Errorf("%s: non-owner should not see private records, got total=%d", e.name, v)
				}
			}

			// GET by ID should return 404
			rec = doGet(t, mux, e.path+"/1", userToken(t, jwtSvc, 5))
			if rec.Code != 404 {
				t.Errorf("%s: non-owner GET should return 404, got %d", e.name, rec.Code)
			}
		})
	}
}

// ============================================================================
// 7. Timestamp Format
// ============================================================================

func TestAPIParity_TimestampFormatRFC3339(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'Timestamp Test', 'Public', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	rec := doGet(t, mux, "/api/v1/accounts/1", adminToken(t, jwtSvc))

	m := decodeJSON(t, rec)

	// Verify created_at and updated_at are RFC3339
	assertTimestampFormat(t, m, "Account", "created_at")
	assertTimestampFormat(t, m, "Account", "updated_at")

	// Verify they contain a timezone indicator (Z or +/-offset)
	for _, f := range []string{"created_at", "updated_at"} {
		s, ok := m[f].(string)
		if !ok {
			continue
		}
		hasTimezone := strings.Contains(s, "Z") || strings.Contains(s, "+") || strings.Contains(s, "-")
		if !hasTimezone {
			t.Errorf("timestamp %q (%q) should contain timezone indicator", f, s)
		}
	}
}

func TestAPIParity_TimestampFormatAllEntities(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Insert one record of each type and verify timestamps
	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'TS Acct', 'Public', ?, ?)`, now, now)
	db.Exec(`INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at)
		VALUES (1, 1, 0, 'TS', 'Contact', 'Public', ?, ?)`, now, now)
	db.Exec(`INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at)
		VALUES (1, 1, 0, 'TS', 'Lead', 'Public', 'new', ?, ?)`, now, now)
	db.Exec(`INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at)
		VALUES (1, 1, 0, 'TS Opp', 'Public', 'prospecting', 0, 0, ?, ?)`, now, now)
	db.Exec(`INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, created_at, updated_at)
		VALUES (1, 1, 0, 'TS Camp', 'Public', 'active', ?, ?)`, now, now)
	db.Exec(`INSERT INTO tasks (id, user_id, assigned_to, name, bucket, priority, created_at, updated_at)
		VALUES (1, 1, 0, 'TS Task', 'due_today', 'high', ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	token := adminToken(t, jwtSvc)

	paths := []struct {
		name string
		path string
	}{
		{"Account", "/api/v1/accounts/1"},
		{"Contact", "/api/v1/contacts/1"},
		{"Lead", "/api/v1/leads/1"},
		{"Opportunity", "/api/v1/opportunities/1"},
		{"Campaign", "/api/v1/campaigns/1"},
		{"Task", "/api/v1/tasks/1"},
	}

	for _, p := range paths {
		t.Run(p.name, func(t *testing.T) {
			rec := doGet(t, mux, p.path, token)
			if rec.Code != 200 {
				t.Fatalf("%s: expected 200, got %d: %s", p.name, rec.Code, rec.Body.String())
			}
			m := decodeJSON(t, rec)
			assertTimestampFormat(t, m, p.name, "created_at")
			assertTimestampFormat(t, m, p.name, "updated_at")
		})
	}
}

// ============================================================================
// Additional: Pagination metadata across entities
// ============================================================================

func TestAPIParity_PaginationAllEntities(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	// Ensure each entity list returns proper pagination envelope even with 0 records
	mux, jwtSvc := entitiesRouter(t, db)
	token := adminToken(t, jwtSvc)

	_ = now // Used in inserts above

	paths := []string{
		"/api/v1/accounts",
		"/api/v1/contacts",
		"/api/v1/leads",
		"/api/v1/opportunities",
		"/api/v1/campaigns",
		"/api/v1/tasks",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rec := doGet(t, mux, path, token)
			if rec.Code != 200 {
				t.Fatalf("%s: expected 200, got %d", path, rec.Code)
			}

			envelope, _ := decodeJSONList(t, rec)

			// All pagination fields must be present
			for _, f := range []string{"data", "total", "page", "per_page", "total_pages"} {
				assertFieldPresent(t, envelope, path, f)
			}

			// total should be 0
			if total, ok := envelope["total"].(json.Number); ok {
				if v, _ := total.Int64(); v != 0 {
					t.Errorf("%s: expected total=0 for empty table, got %d", path, v)
				}
			}

			// total_pages should be 0 for empty result
			if tp, ok := envelope["total_pages"].(json.Number); ok {
				if v, _ := tp.Int64(); v != 0 {
					t.Errorf("%s: expected total_pages=0 for empty table, got %d", path, v)
				}
			}

			// page should still be 1 (requested page)
			if page, ok := envelope["page"].(json.Number); ok {
				if v, _ := page.Int64(); v != 1 {
					t.Errorf("%s: expected page=1, got %d", path, v)
				}
			}
		})
	}
}

// ============================================================================
// Numeric precision for decimal fields
// ============================================================================

func TestAPIParity_DecimalFieldPrecision(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, discount, probability, created_at, updated_at)
		VALUES (1, 1, 0, 'Precision Deal', 'Public', 'prospecting', 99999.99, 1234.56, 50, ?, ?)`, now, now)
	db.Exec(`INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, budget, target_revenue, revenue, created_at, updated_at)
		VALUES (1, 1, 0, 'Precision Camp', 'Public', 'active', 75000.50, 200000.75, 150000.25, ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	token := adminToken(t, jwtSvc)

	// Test opportunity decimals
	rec := doGet(t, mux, "/api/v1/opportunities/1", token)
	m := decodeJSON(t, rec)

	if amt, ok := m["amount"].(json.Number); ok {
		f, err := amt.Float64()
		if err != nil {
			t.Errorf("amount should be parseable as float64: %v", err)
		}
		if f != 99999.99 {
			t.Errorf("amount: expected 99999.99, got %f", f)
		}
	}
	if disc, ok := m["discount"].(json.Number); ok {
		f, _ := disc.Float64()
		if f != 1234.56 {
			t.Errorf("discount: expected 1234.56, got %f", f)
		}
	}

	// Test campaign decimals
	rec = doGet(t, mux, "/api/v1/campaigns/1", token)
	m = decodeJSON(t, rec)

	if budget, ok := m["budget"].(json.Number); ok {
		f, _ := budget.Float64()
		if f != 75000.50 {
			t.Errorf("budget: expected 75000.50, got %f", f)
		}
	}
}

// ============================================================================
// Integer count fields are proper integers
// ============================================================================

func TestAPIParity_CountFieldsAreIntegers(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")

	db.Exec(`INSERT INTO accounts (id, user_id, assigned_to, name, access, contacts_count, opportunities_count, created_at, updated_at)
		VALUES (1, 1, 0, 'Count Test', 'Public', 15, 7, ?, ?)`, now, now)
	db.Exec(`INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, leads_count, opportunities_count, created_at, updated_at)
		VALUES (1, 1, 0, 'Count Camp', 'Public', 'active', 42, 12, ?, ?)`, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	token := adminToken(t, jwtSvc)

	// Account counts
	rec := doGet(t, mux, "/api/v1/accounts/1", token)
	m := decodeJSON(t, rec)

	for _, f := range []string{"contacts_count", "opportunities_count"} {
		n, ok := m[f].(json.Number)
		if !ok {
			t.Errorf("Account.%s should be a number", f)
			continue
		}
		v, err := n.Int64()
		if err != nil {
			t.Errorf("Account.%s should be an integer: %v", f, err)
			continue
		}
		// Verify actual values
		if f == "contacts_count" && v != 15 {
			t.Errorf("Account.contacts_count: expected 15, got %d", v)
		}
		if f == "opportunities_count" && v != 7 {
			t.Errorf("Account.opportunities_count: expected 7, got %d", v)
		}
	}

	// Campaign counts
	rec = doGet(t, mux, "/api/v1/campaigns/1", token)
	m = decodeJSON(t, rec)

	for _, f := range []string{"leads_count", "opportunities_count"} {
		n, ok := m[f].(json.Number)
		if !ok {
			t.Errorf("Campaign.%s should be a number", f)
			continue
		}
		if _, err := n.Int64(); err != nil {
			t.Errorf("Campaign.%s should be an integer: %v", f, err)
		}
	}
}
