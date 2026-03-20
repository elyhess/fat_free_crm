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
)

func setupDashboardDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	db.Exec(`CREATE TABLE tasks (
		id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER,
		completed_by INTEGER, name TEXT, asset_id INTEGER, asset_type TEXT,
		priority TEXT, category TEXT, bucket TEXT,
		due_at DATETIME, completed_at DATETIME, background_info TEXT,
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
	// Needed for router init (permissions, groups_users, accounts, contacts, leads, campaigns, field_groups, fields, users)
	db.Exec("CREATE TABLE permissions (id INTEGER PRIMARY KEY, user_id INTEGER, group_id INTEGER, asset_id INTEGER, asset_type TEXT, created_at DATETIME, updated_at DATETIME)")
	db.Exec("CREATE TABLE groups_users (user_id INTEGER, group_id INTEGER)")
	db.Exec("CREATE TABLE accounts (id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER, name TEXT, access TEXT, rating INTEGER, category TEXT, email TEXT, website TEXT, phone TEXT, toll_free_phone TEXT, fax TEXT, background_info TEXT, contacts_count INTEGER, opportunities_count INTEGER, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	db.Exec("CREATE TABLE contacts (id INTEGER PRIMARY KEY, user_id INTEGER, lead_id INTEGER, assigned_to INTEGER, reports_to INTEGER, first_name TEXT, last_name TEXT, access TEXT, title TEXT, department TEXT, email TEXT, alt_email TEXT, phone TEXT, mobile TEXT, fax TEXT, blog TEXT, linkedin TEXT, facebook TEXT, twitter TEXT, born_on DATE, do_not_call INTEGER, background_info TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	db.Exec("CREATE TABLE leads (id INTEGER PRIMARY KEY, user_id INTEGER, campaign_id INTEGER, assigned_to INTEGER, first_name TEXT, last_name TEXT, access TEXT, company TEXT, title TEXT, source TEXT, status TEXT, referred_by TEXT, email TEXT, alt_email TEXT, phone TEXT, mobile TEXT, blog TEXT, linkedin TEXT, facebook TEXT, twitter TEXT, rating INTEGER, do_not_call INTEGER, background_info TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	db.Exec("CREATE TABLE campaigns (id INTEGER PRIMARY KEY, user_id INTEGER, assigned_to INTEGER, name TEXT, access TEXT, status TEXT, budget DECIMAL, target_leads INTEGER, target_conversion FLOAT, target_revenue DECIMAL, leads_count INTEGER, opportunities_count INTEGER, revenue DECIMAL, starts_on DATE, ends_on DATE, objectives TEXT, background_info TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	db.Exec("CREATE TABLE field_groups (id INTEGER PRIMARY KEY, klass_name TEXT, label TEXT, name TEXT, tag_id INTEGER, position INTEGER, created_at DATETIME, updated_at DATETIME)")
	db.Exec("CREATE TABLE fields (id INTEGER PRIMARY KEY, type TEXT, field_group_id INTEGER, position INTEGER, name TEXT, label TEXT, hint TEXT, placeholder TEXT, as_field TEXT, collection TEXT, disabled INTEGER, required INTEGER, maxlength INTEGER, minlength INTEGER, created_at DATETIME, updated_at DATETIME)")
	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT, email TEXT, encrypted_password TEXT, password_salt TEXT, admin INTEGER, sign_in_count INTEGER, current_sign_in_at DATETIME, last_sign_in_at DATETIME, current_sign_in_ip TEXT, last_sign_in_ip TEXT, confirmed_at DATETIME, suspended_at DATETIME, first_name TEXT, last_name TEXT, title TEXT, company TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)")
	return db
}

func dashboardRouter(t *testing.T, db *gorm.DB) (*http.ServeMux, *auth.JWTService) {
	t.Helper()
	jwtSvc := auth.NewJWTService("test-secret", time.Hour)
	cfg := RouterConfig{DB: db, JWTSecret: "test-secret", JWTExpiryHours: 1}
	router := NewRouter(cfg)
	mux := http.NewServeMux()
	mux.Handle("/", router)
	return mux, jwtSvc
}

func TestTaskSummary_Empty(t *testing.T) {
	db := setupDashboardDB(t)
	mux, jwtSvc := dashboardRouter(t, db)

	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp TaskSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalTasks != 0 {
		t.Errorf("expected 0 total tasks, got %d", resp.TotalTasks)
	}
	if len(resp.Buckets) != 7 {
		t.Errorf("expected 7 buckets, got %d", len(resp.Buckets))
	}
}

func TestTaskSummary_WithTasks(t *testing.T) {
	db := setupDashboardDB(t)
	now := time.Now()
	todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := todayNoon.AddDate(0, 0, -1)
	tomorrow := todayNoon.AddDate(0, 0, 1)

	ts := func(t time.Time) string { return t.Format("2006-01-02 15:04:05") }
	n := ts(now)

	// Task owned by user 1, due today
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, due_at, created_at, updated_at) VALUES (1, 1, 0, 'Today task', ?, ?, ?)", ts(todayNoon), n, n)
	// Task owned by user 1, overdue
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, due_at, created_at, updated_at) VALUES (2, 1, 0, 'Overdue task', ?, ?, ?)", ts(yesterday), n, n)
	// Task assigned to user 1, due tomorrow
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, due_at, created_at, updated_at) VALUES (3, 99, 1, 'Tomorrow task', ?, ?, ?)", ts(tomorrow), n, n)
	// Completed task - should not appear
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, due_at, completed_at, created_at, updated_at) VALUES (4, 1, 0, 'Done task', ?, ?, ?, ?)", ts(todayNoon), n, n, n)
	// Task for different user - should not appear
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, due_at, created_at, updated_at) VALUES (5, 99, 88, 'Other task', ?, ?, ?)", ts(todayNoon), n, n)
	// ASAP task (no due_at, bucket=due_asap)
	db.Exec("INSERT INTO tasks (id, user_id, assigned_to, name, bucket, created_at, updated_at) VALUES (6, 1, 0, 'ASAP task', 'due_asap', ?, ?)", n, n)

	mux, jwtSvc := dashboardRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "user", false)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp TaskSummaryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Should see: ASAP(1) + overdue(1) + today(1) + tomorrow(1) = 4
	if resp.TotalTasks != 4 {
		t.Errorf("expected 4 total tasks, got %d", resp.TotalTasks)
	}

	bucketMap := make(map[string]int64)
	for _, b := range resp.Buckets {
		bucketMap[b.Bucket] = b.Count
	}
	if bucketMap["due_asap"] != 1 {
		t.Errorf("expected 1 due_asap, got %d", bucketMap["due_asap"])
	}
	if bucketMap["overdue"] != 1 {
		t.Errorf("expected 1 overdue, got %d", bucketMap["overdue"])
	}
	if bucketMap["due_today"] != 1 {
		t.Errorf("expected 1 due_today, got %d", bucketMap["due_today"])
	}
	if bucketMap["due_tomorrow"] != 1 {
		t.Errorf("expected 1 due_tomorrow, got %d", bucketMap["due_tomorrow"])
	}
}

func TestPipelineSummary_Empty(t *testing.T) {
	db := setupDashboardDB(t)
	mux, jwtSvc := dashboardRouter(t, db)

	tok, _ := jwtSvc.GenerateToken(1, "admin", true)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/pipeline", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp PipelineResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalCount != 0 {
		t.Errorf("expected 0 total, got %d", resp.TotalCount)
	}
}

func TestPipelineSummary_WithOpportunities(t *testing.T) {
	db := setupDashboardDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")

	// Prospecting: $10k, 25% probability -> weighted $2500
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, probability, created_at, updated_at) VALUES (1, 1, 0, 'Deal A', 'prospecting', 10000, 25, ?, ?)", n, n)
	// Negotiation: $50k, 75% probability -> weighted $37500
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, probability, created_at, updated_at) VALUES (2, 1, 0, 'Deal B', 'negotiation', 50000, 75, ?, ?)", n, n)
	// Won: should NOT appear in pipeline
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, probability, created_at, updated_at) VALUES (3, 1, 0, 'Won Deal', 'won', 100000, 100, ?, ?)", n, n)
	// Lost: should NOT appear
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, probability, created_at, updated_at) VALUES (4, 1, 0, 'Lost Deal', 'lost', 5000, 0, ?, ?)", n, n)
	// Different user - should NOT appear
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, probability, created_at, updated_at) VALUES (5, 99, 88, 'Other Deal', 'prospecting', 20000, 50, ?, ?)", n, n)

	mux, jwtSvc := dashboardRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "user", false)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/pipeline", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp PipelineResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("expected 2 pipeline opportunities, got %d", resp.TotalCount)
	}
	if resp.TotalAmount != 60000 {
		t.Errorf("expected total amount 60000, got %.2f", resp.TotalAmount)
	}
	// 10000*25/100 + 50000*75/100 = 2500 + 37500 = 40000
	if resp.TotalWeighted != 40000 {
		t.Errorf("expected weighted 40000, got %.2f", resp.TotalWeighted)
	}
}

func TestPipelineSummary_WithDiscount(t *testing.T) {
	db := setupDashboardDB(t)
	n := time.Now().Format("2006-01-02 15:04:05")

	// $10k amount, $2k discount, 50% probability -> weighted ($10k-$2k)*50% = $4000
	db.Exec("INSERT INTO opportunities (id, user_id, assigned_to, name, stage, amount, discount, probability, created_at, updated_at) VALUES (1, 1, 0, 'Discounted', 'prospecting', 10000, 2000, 50, ?, ?)", n, n)

	mux, jwtSvc := dashboardRouter(t, db)
	tok, _ := jwtSvc.GenerateToken(1, "user", false)
	req := httptest.NewRequest("GET", "/api/v1/dashboard/pipeline", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var resp PipelineResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.TotalAmount != 8000 {
		t.Errorf("expected total amount 8000 (10000-2000), got %.2f", resp.TotalAmount)
	}
	if resp.TotalWeighted != 4000 {
		t.Errorf("expected weighted 4000, got %.2f", resp.TotalWeighted)
	}
}

func TestDashboard_NoAuth(t *testing.T) {
	db := setupDashboardDB(t)
	mux, _ := dashboardRouter(t, db)

	for _, path := range []string{"/api/v1/dashboard/tasks", "/api/v1/dashboard/pipeline"} {
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != 401 {
			t.Errorf("%s: expected 401, got %d", path, rec.Code)
		}
	}
}
