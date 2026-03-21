package service

import (
	"encoding/json"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func setupVersionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })

	db.Exec(`CREATE TABLE versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		item_type TEXT NOT NULL,
		item_id INTEGER NOT NULL,
		event TEXT NOT NULL,
		whodunnit TEXT,
		object TEXT,
		object_changes TEXT,
		related_id INTEGER,
		related_type TEXT,
		transaction_id INTEGER,
		created_at DATETIME
	)`)
	return db
}

func TestRecordCreate(t *testing.T) {
	db := setupVersionDB(t)
	rec := NewVersionRecorder(db)

	type TestEntity struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	entity := TestEntity{ID: 42, Name: "Acme Corp"}
	rec.RecordCreate("Account", 42, 1, entity)

	var ver model.Version
	if err := db.First(&ver).Error; err != nil {
		t.Fatalf("query version: %v", err)
	}

	if ver.ItemType != "Account" {
		t.Errorf("item_type: got %q, want Account", ver.ItemType)
	}
	if ver.ItemID != 42 {
		t.Errorf("item_id: got %d, want 42", ver.ItemID)
	}
	if ver.Event != "create" {
		t.Errorf("event: got %q, want create", ver.Event)
	}
	if ver.Whodunnit == nil || *ver.Whodunnit != "1" {
		t.Errorf("whodunnit: got %v, want '1'", ver.Whodunnit)
	}
	if ver.Object == nil {
		t.Fatal("object should not be nil")
	}

	var decoded TestEntity
	if err := json.Unmarshal([]byte(*ver.Object), &decoded); err != nil {
		t.Fatalf("unmarshal object: %v", err)
	}
	if decoded.Name != "Acme Corp" {
		t.Errorf("object.name: got %q, want Acme Corp", decoded.Name)
	}
	if ver.ObjectChanges != nil {
		t.Errorf("object_changes should be nil for create, got %v", *ver.ObjectChanges)
	}
}

func TestRecordUpdate(t *testing.T) {
	db := setupVersionDB(t)
	rec := NewVersionRecorder(db)

	type TestEntity struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	old := TestEntity{ID: 10, Name: "Old Name"}
	changes := map[string]interface{}{"name": "New Name"}
	rec.RecordUpdate("Contact", 10, 5, old, changes)

	var ver model.Version
	if err := db.First(&ver).Error; err != nil {
		t.Fatalf("query version: %v", err)
	}

	if ver.Event != "update" {
		t.Errorf("event: got %q, want update", ver.Event)
	}
	if ver.Whodunnit == nil || *ver.Whodunnit != "5" {
		t.Errorf("whodunnit: got %v, want '5'", ver.Whodunnit)
	}

	// Check object has old state
	var decoded TestEntity
	if err := json.Unmarshal([]byte(*ver.Object), &decoded); err != nil {
		t.Fatalf("unmarshal object: %v", err)
	}
	if decoded.Name != "Old Name" {
		t.Errorf("object should have old state, got %q", decoded.Name)
	}

	// Check changes
	if ver.ObjectChanges == nil {
		t.Fatal("object_changes should not be nil for update")
	}
	var decodedChanges map[string]interface{}
	if err := json.Unmarshal([]byte(*ver.ObjectChanges), &decodedChanges); err != nil {
		t.Fatalf("unmarshal changes: %v", err)
	}
	if decodedChanges["name"] != "New Name" {
		t.Errorf("changes.name: got %v, want 'New Name'", decodedChanges["name"])
	}
}

func TestRecordDestroy(t *testing.T) {
	db := setupVersionDB(t)
	rec := NewVersionRecorder(db)

	type TestEntity struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	entity := TestEntity{ID: 7, Name: "To Delete"}
	rec.RecordDestroy("Lead", 7, 3, entity)

	var ver model.Version
	if err := db.First(&ver).Error; err != nil {
		t.Fatalf("query version: %v", err)
	}

	if ver.ItemType != "Lead" {
		t.Errorf("item_type: got %q, want Lead", ver.ItemType)
	}
	if ver.Event != "destroy" {
		t.Errorf("event: got %q, want destroy", ver.Event)
	}
	if ver.Object == nil {
		t.Fatal("object should not be nil for destroy")
	}
}

func TestRecordCreate_Timestamp(t *testing.T) {
	db := setupVersionDB(t)
	rec := NewVersionRecorder(db)

	before := time.Now().Add(-time.Second)
	rec.RecordCreate("Campaign", 1, 1, map[string]string{"name": "test"})
	after := time.Now().Add(time.Second)

	var ver model.Version
	db.First(&ver)

	if ver.CreatedAt.Before(before) || ver.CreatedAt.After(after) {
		t.Errorf("created_at %v not between %v and %v", ver.CreatedAt, before, after)
	}
}

func TestMultipleVersions(t *testing.T) {
	db := setupVersionDB(t)
	rec := NewVersionRecorder(db)

	rec.RecordCreate("Account", 1, 1, map[string]string{"name": "v1"})
	rec.RecordUpdate("Account", 1, 1, map[string]string{"name": "v1"}, map[string]interface{}{"name": "v2"})
	rec.RecordDestroy("Account", 1, 1, map[string]string{"name": "v2"})

	var count int64
	db.Model(&model.Version{}).Count(&count)
	if count != 3 {
		t.Errorf("expected 3 versions, got %d", count)
	}

	var versions []model.Version
	db.Order("id").Find(&versions)
	events := []string{versions[0].Event, versions[1].Event, versions[2].Event}
	if events[0] != "create" || events[1] != "update" || events[2] != "destroy" {
		t.Errorf("events: got %v, want [create update destroy]", events)
	}
}
