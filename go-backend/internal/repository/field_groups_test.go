package repository

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.FieldGroup{}, &model.Field{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedFieldGroups(t *testing.T, db *gorm.DB) {
	t.Helper()
	groups := []model.FieldGroup{
		{ID: 1, Name: "general_information", Label: "General Information", Position: 1, KlassName: "Account"},
		{ID: 2, Name: "contact_information", Label: "Contact Information", Position: 2, KlassName: "Account"},
		{ID: 3, Name: "general_information", Label: "General Information", Position: 1, KlassName: "Contact"},
	}
	for _, g := range groups {
		if err := db.Create(&g).Error; err != nil {
			t.Fatalf("failed to seed group: %v", err)
		}
	}

	fields := []model.Field{
		{ID: 1, Type: "CustomField", FieldGroupID: 1, Position: 1, Name: "cf_industry", Label: "Industry", As: "select", Required: true},
		{ID: 2, Type: "CustomField", FieldGroupID: 1, Position: 2, Name: "cf_employees", Label: "Employees", As: "integer"},
		{ID: 3, Type: "CoreField", FieldGroupID: 2, Position: 1, Name: "phone", Label: "Phone", As: "tel"},
		{ID: 4, Type: "CustomField", FieldGroupID: 3, Position: 1, Name: "cf_nickname", Label: "Nickname", As: "string"},
	}
	for _, f := range fields {
		if err := db.Create(&f).Error; err != nil {
			t.Fatalf("failed to seed field: %v", err)
		}
	}
}

func TestFindByEntity(t *testing.T) {
	db := setupTestDB(t)
	seedFieldGroups(t, db)
	repo := NewFieldGroupRepository(db)

	groups, err := repo.FindByEntity("Account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// Should be ordered by position
	if groups[0].Label != "General Information" {
		t.Errorf("expected first group 'General Information', got %q", groups[0].Label)
	}
	if groups[1].Label != "Contact Information" {
		t.Errorf("expected second group 'Contact Information', got %q", groups[1].Label)
	}

	// First group should have 2 fields
	if len(groups[0].Fields) != 2 {
		t.Errorf("expected 2 fields in first group, got %d", len(groups[0].Fields))
	}

	// Fields should be ordered by position
	if groups[0].Fields[0].Name != "cf_industry" {
		t.Errorf("expected first field 'cf_industry', got %q", groups[0].Fields[0].Name)
	}
}

func TestFindByEntity_Contact(t *testing.T) {
	db := setupTestDB(t)
	seedFieldGroups(t, db)
	repo := NewFieldGroupRepository(db)

	groups, err := repo.FindByEntity("Contact")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group for Contact, got %d", len(groups))
	}
	if len(groups[0].Fields) != 1 {
		t.Errorf("expected 1 field in Contact group, got %d", len(groups[0].Fields))
	}
}

func TestFindByEntity_InvalidType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFieldGroupRepository(db)

	_, err := repo.FindByEntity("Invalid")
	if err == nil {
		t.Error("expected error for invalid entity type")
	}
}

func TestFindByEntity_NoResults(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFieldGroupRepository(db)

	groups, err := repo.FindByEntity("Lead")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for Lead, got %d", len(groups))
	}
}

func TestFindCustomFieldsByEntity(t *testing.T) {
	db := setupTestDB(t)
	seedFieldGroups(t, db)
	repo := NewFieldGroupRepository(db)

	fields, err := repo.FindCustomFieldsByEntity("Account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return 2 CustomFields, excluding the CoreField
	if len(fields) != 2 {
		t.Fatalf("expected 2 custom fields, got %d", len(fields))
	}
	for _, f := range fields {
		if f.Type == "CoreField" {
			t.Error("should not include CoreField records")
		}
	}
}

func TestFindFieldsByEntity(t *testing.T) {
	db := setupTestDB(t)
	seedFieldGroups(t, db)
	repo := NewFieldGroupRepository(db)

	fields, err := repo.FindFieldsByEntity("Account")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return all 3 Account fields (2 custom + 1 core)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
}
