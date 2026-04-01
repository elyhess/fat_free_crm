package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func entitiesRouter(t *testing.T, db *gorm.DB) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	// Wrap chi router in http.ServeMux for testing
	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux, jwtSvc
}

func adminToken(t *testing.T, jwtSvc *auth.JWTService) string {
	t.Helper()
	tok, err := jwtSvc.GenerateToken(1, "admin", true)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return tok
}

func userToken(t *testing.T, jwtSvc *auth.JWTService, userID int64) string {
	t.Helper()
	tok, err := jwtSvc.GenerateToken(userID, "user", false)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return tok
}

func TestListAccounts_Empty(t *testing.T) {
	db := testDB(t)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Account]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 total, got %d", result.Total)
	}
}

func TestListAccounts_WithData(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Corp', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 0, 'Globex', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 total, got %d", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Data))
	}
}

func TestListAccounts_Pagination(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	for i := 1; i <= 5; i++ {
		db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (?, 1, 0, ?, 'Public', ?, ?)", i, "Account "+string(rune('A'+i-1)), now, now)
	}

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts?page=2&per_page=2", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected 5 total, got %d", result.Total)
	}
	if result.Page != 2 {
		t.Errorf("expected page 2, got %d", result.Page)
	}
	if len(result.Data) != 2 {
		t.Errorf("expected 2 items on page 2, got %d", len(result.Data))
	}
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
}

func TestListAccounts_AccessControl(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	// Public account - visible to all
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 99, 0, 'Public Corp', 'Public', ?, ?)", now, now)
	// Private account owned by user 99 - NOT visible to user 5
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 99, 0, 'Secret Corp', 'Private', ?, ?)", now, now)
	// Private account owned by user 5 - visible
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (3, 5, 0, 'My Corp', 'Private', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t, jwtSvc, 5))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Should see: Public Corp + My Corp = 2
	if result.Total != 2 {
		t.Errorf("expected 2 accessible accounts, got %d", result.Total)
	}
}

func TestGetAccount_Found(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Corp', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var acct model.Account
	if err := json.NewDecoder(rec.Body).Decode(&acct); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if acct.Name != "Acme Corp" {
		t.Errorf("expected 'Acme Corp', got %q", acct.Name)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	db := testDB(t)
	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts/999", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetAccount_AccessDenied(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 99, 0, 'Secret Corp', 'Private', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts/1", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t, jwtSvc, 5))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// Private record owned by user 99, accessed by user 5 -> not found (filtered by scope)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestListTasks(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, bucket, priority, created_at, updated_at) VALUES (1, 1, 0, 'Call Bob', 'due_today', 'high', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Task]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 task, got %d", result.Total)
	}
	if len(result.Data) > 0 && result.Data[0].Name != "Call Bob" {
		t.Errorf("expected 'Call Bob', got %q", result.Data[0].Name)
	}
}

func TestListLeads(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (1, 1, 0, 'John', 'Doe', 'Public', 'new', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/leads", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Lead]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 lead, got %d", result.Total)
	}
}

func TestListOpportunities(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at) VALUES (1, 1, 0, 'Big Deal', 'Public', 'prospecting', 50000, 25, ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/opportunities", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Opportunity]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 opportunity, got %d", result.Total)
	}
}

func TestListCampaigns(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO campaigns (id, user_id, assigned_to, name, access, status, created_at, updated_at) VALUES (1, 1, 0, 'Q1 Push', 'Public', 'active', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/campaigns", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Campaign]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 campaign, got %d", result.Total)
	}
}

func TestListContacts(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO contacts (id, user_id, assigned_to, first_name, last_name, access, created_at, updated_at) VALUES (1, 1, 0, 'Jane', 'Smith', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Contact]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 contact, got %d", result.Total)
	}
}

func TestSoftDeletedRecordsExcluded(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Active', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at, deleted_at) VALUES (2, 1, 0, 'Deleted', 'Public', ?, ?, ?)", now, now, now)

	mux, jwtSvc := entitiesRouter(t, db)
	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 (soft-deleted excluded), got %d", result.Total)
	}
}

func TestNoToken_Returns401(t *testing.T) {
	db := testDB(t)
	mux, _ := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- Filter Tests ---

func TestListAccounts_FilterByName(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Corp', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 0, 'Beta Industries', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (3, 1, 0, 'Acme Labs', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// filter[name] defaults to "cont" (ILIKE) for string columns
	req := httptest.NewRequest("GET", "/api/v1/accounts?filter[name]=Acme", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result model.PaginatedResult[model.Account]
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 2 {
		t.Errorf("expected 2 accounts matching 'Acme', got %d", result.Total)
	}
}

func TestListAccounts_FilterByNameExact(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme Corp', 'Public', ?, ?)", now, now)
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (2, 1, 0, 'Acme Labs', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts?filter[name_eq]=Acme+Corp", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 1 {
		t.Errorf("expected 1 exact match, got %d", result.Total)
	}
}

func TestListLeads_FilterByStatus(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (1, 1, 0, 'John', 'Doe', 'Public', 'new', ?, ?)", now, now)
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (2, 1, 0, 'Jane', 'Smith', 'Public', 'contacted', ?, ?)", now, now)
	db.Exec("INSERT INTO leads (id, user_id, assigned_to, first_name, last_name, access, status, created_at, updated_at) VALUES (3, 1, 0, 'Bob', 'Jones', 'Public', 'new', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/leads?filter[status_eq]=new", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Lead]
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 2 {
		t.Errorf("expected 2 leads with status=new, got %d", result.Total)
	}
}

func TestListOpportunities_FilterByStage(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at) VALUES (1, 1, 0, 'Deal A', 'Public', 'prospecting', 10000, 10, ?, ?)", now, now)
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at) VALUES (2, 1, 0, 'Deal B', 'Public', 'won', 50000, 100, ?, ?)", now, now)
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, access, stage, amount, probability, created_at, updated_at) VALUES (3, 1, 0, 'Deal C', 'Public', 'prospecting', 20000, 20, ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/opportunities?filter[stage_eq]=prospecting", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Opportunity]
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 2 {
		t.Errorf("expected 2 opportunities with stage=prospecting, got %d", result.Total)
	}
}

func TestListAccounts_FilterIgnoresInvalidColumn(t *testing.T) {
	db := testDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	db.Exec("INSERT INTO accounts (id, user_id, assigned_to, name, access, created_at, updated_at) VALUES (1, 1, 0, 'Acme', 'Public', ?, ?)", now, now)

	mux, jwtSvc := entitiesRouter(t, db)

	// Invalid column should be ignored, returning all results
	req := httptest.NewRequest("GET", "/api/v1/accounts?filter[sql_injection]=DROP+TABLE", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, jwtSvc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var result model.PaginatedResult[model.Account]
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 1 {
		t.Errorf("expected 1 (filter should be ignored), got %d", result.Total)
	}
}
