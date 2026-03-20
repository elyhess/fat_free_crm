package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// testEntity implements EntityRecord for testing.
type testEntity struct {
	ID         int64
	Access     string
	UserID     int64
	AssignedTo int64
	AssetType  string
}

func (e testEntity) GetID() int64         { return e.ID }
func (e testEntity) GetAccess() string    { return e.Access }
func (e testEntity) GetUserID() int64     { return e.UserID }
func (e testEntity) GetAssignedTo() int64 { return e.AssignedTo }
func (e testEntity) GetAssetType() string { return e.AssetType }

func setupAuthzDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Permission{}); err != nil {
		t.Fatalf("failed to migrate permissions: %v", err)
	}
	// Create groups_users table manually (join table)
	db.Exec("CREATE TABLE IF NOT EXISTS groups_users (user_id INTEGER, group_id INTEGER)")
	return db
}

func TestCanAccess_AdminBypassesAll(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessPrivate, UserID: 99, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(1, true, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("admin should access any record")
	}
}

func TestCanAccess_PublicRecord(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessPublic, UserID: 99, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("any user should access public records")
	}
}

func TestCanAccess_PrivateRecord_Owner(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessPrivate, UserID: 5, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("owner should access their private record")
	}
}

func TestCanAccess_PrivateRecord_Assignee(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessPrivate, UserID: 99, AssignedTo: 5, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("assignee should access private record assigned to them")
	}
}

func TestCanAccess_PrivateRecord_OtherUser(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessPrivate, UserID: 99, AssignedTo: 88, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("other user should NOT access private record")
	}
}

func TestCanAccess_SharedRecord_Owner(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessShared, UserID: 5, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("owner should access shared record")
	}
}

func TestCanAccess_SharedRecord_WithUserPermission(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	userID := int64(5)
	db.Create(&model.Permission{
		UserID: &userID, AssetID: 1, AssetType: "Account",
	})

	entity := testEntity{ID: 1, Access: model.AccessShared, UserID: 99, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("user with permission should access shared record")
	}
}

func TestCanAccess_SharedRecord_WithGroupPermission(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	// User 5 is in group 10
	db.Exec("INSERT INTO groups_users (user_id, group_id) VALUES (?, ?)", 5, 10)

	groupID := int64(10)
	db.Create(&model.Permission{
		GroupID: &groupID, AssetID: 1, AssetType: "Account",
	})

	entity := testEntity{ID: 1, Access: model.AccessShared, UserID: 99, AssignedTo: 0, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("user in permitted group should access shared record")
	}
}

func TestCanAccess_SharedRecord_NoPermission(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	entity := testEntity{ID: 1, Access: model.AccessShared, UserID: 99, AssignedTo: 88, AssetType: "Account"}
	ok, err := svc.CanAccess(5, false, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("user without permission should NOT access shared record")
	}
}

func TestScopeAccessible_AdminGetsAll(t *testing.T) {
	db := setupAuthzDB(t)
	// Create a simple accounts table for scope testing
	db.Exec("CREATE TABLE accounts (id INTEGER PRIMARY KEY, access TEXT, user_id INTEGER, assigned_to INTEGER)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (1, 'Private', 99, 0)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (2, 'Public', 99, 0)")

	svc := NewAuthorizationService(db)

	var count int64
	db.Table("accounts").Scopes(svc.ScopeAccessible(1, true, "Account")).Count(&count)
	if count != 2 {
		t.Errorf("admin should see all 2 records, got %d", count)
	}
}

func TestScopeAccessible_RegularUserFilters(t *testing.T) {
	db := setupAuthzDB(t)
	db.Exec("CREATE TABLE accounts (id INTEGER PRIMARY KEY, access TEXT, user_id INTEGER, assigned_to INTEGER)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (1, 'Public', 99, 0)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (2, 'Private', 99, 0)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (3, 'Private', 5, 0)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (4, 'Private', 99, 5)")

	svc := NewAuthorizationService(db)

	var count int64
	db.Table("accounts").Scopes(svc.ScopeAccessible(5, false, "Account")).Count(&count)
	// Should see: #1 (public), #3 (owner), #4 (assignee) = 3
	if count != 3 {
		t.Errorf("expected 3 accessible records, got %d", count)
	}
}

func TestScopeAccessible_SharedWithPermission(t *testing.T) {
	db := setupAuthzDB(t)
	db.Exec("CREATE TABLE accounts (id INTEGER PRIMARY KEY, access TEXT, user_id INTEGER, assigned_to INTEGER)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (1, 'Shared', 99, 0)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (2, 'Shared', 99, 0)")

	// User 5 has permission for account 1 only
	userID := int64(5)
	db.Create(&model.Permission{UserID: &userID, AssetID: 1, AssetType: "Account"})

	svc := NewAuthorizationService(db)

	var count int64
	db.Table("accounts").Scopes(svc.ScopeAccessible(5, false, "Account")).Count(&count)
	// Should see only #1 (shared with permission)
	if count != 1 {
		t.Errorf("expected 1 accessible shared record, got %d", count)
	}
}

func TestScopeAccessible_InvalidAssetType(t *testing.T) {
	db := setupAuthzDB(t)
	svc := NewAuthorizationService(db)

	db.Exec("CREATE TABLE accounts (id INTEGER PRIMARY KEY, access TEXT, user_id INTEGER, assigned_to INTEGER)")
	db.Exec("INSERT INTO accounts (id, access, user_id, assigned_to) VALUES (1, 'Public', 1, 0)")

	var count int64
	db.Table("accounts").Scopes(svc.ScopeAccessible(1, false, "Invalid")).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 for invalid asset type, got %d", count)
	}
}
