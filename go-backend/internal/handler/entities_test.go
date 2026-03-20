package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func setupEntitiesDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Permission{}); err != nil {
		t.Fatalf("migrate permissions: %v", err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS groups_users (user_id INTEGER, group_id INTEGER)")

	// Create entity tables
	db.Exec(`CREATE TABLE accounts (
		id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER,
		name TEXT, access TEXT DEFAULT 'Public', rating INTEGER DEFAULT 0,
		category TEXT, email TEXT, website TEXT, phone TEXT,
		toll_free_phone TEXT, fax TEXT, background_info TEXT,
		contacts_count INTEGER DEFAULT 0, opportunities_count INTEGER DEFAULT 0,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	db.Exec(`CREATE TABLE contacts (
		id INTEGER PRIMARY KEY, user_id INTEGER, lead_id INTEGER,
		assigned_to INTEGER, reports_to INTEGER,
		first_name TEXT, last_name TEXT, access TEXT DEFAULT 'Public',
		title TEXT, department TEXT, email TEXT, alt_email TEXT,
		phone TEXT, mobile TEXT, fax TEXT, blog TEXT,
		linkedin TEXT, facebook TEXT, twitter TEXT,
		born_on DATE, do_not_call INTEGER DEFAULT 0, background_info TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	db.Exec(`CREATE TABLE leads (
		id INTEGER PRIMARY KEY, user_id INTEGER, campaign_id INTEGER,
		assigned_to INTEGER, first_name TEXT, last_name TEXT,
		access TEXT DEFAULT 'Public', company TEXT, title TEXT,
		source TEXT, status TEXT, referred_by TEXT,
		email TEXT, alt_email TEXT, phone TEXT, mobile TEXT,
		blog TEXT, linkedin TEXT, facebook TEXT, twitter TEXT,
		rating INTEGER DEFAULT 0, do_not_call INTEGER DEFAULT 0,
		background_info TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	db.Exec(`CREATE TABLE opportunities (
		id INTEGER PRIMARY KEY, user_id INTEGER, campaign_id INTEGER,
		assigned_to INTEGER, name TEXT, access TEXT DEFAULT 'Public',
		source TEXT, stage TEXT, probability INTEGER,
		amount DECIMAL(12,2), discount DECIMAL(12,2),
		closes_on DATE, background_info TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	db.Exec(`CREATE TABLE campaigns (
		id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER,
		name TEXT, access TEXT DEFAULT 'Public', status TEXT,
		budget DECIMAL(12,2), target_leads INTEGER, target_conversion FLOAT,
		target_revenue DECIMAL(12,2), leads_count INTEGER DEFAULT 0,
		opportunities_count INTEGER DEFAULT 0, revenue DECIMAL(12,2),
		starts_on DATE, ends_on DATE, objectives TEXT, background_info TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	db.Exec(`CREATE TABLE tasks (
		id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER,
		completed_by INTEGER, name TEXT, asset_id INTEGER, asset_type TEXT,
		priority TEXT, category TEXT, bucket TEXT,
		due_at DATETIME, completed_at DATETIME, background_info TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
	)`)
	return db
}

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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
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
	db := setupEntitiesDB(t)
	mux, _ := entitiesRouter(t, db)

	req := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
