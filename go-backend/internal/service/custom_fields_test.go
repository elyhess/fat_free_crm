package service

import (
	"testing"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

func TestIsValidFieldType(t *testing.T) {
	valid := []string{"string", "text", "email", "url", "tel", "select", "radio_buttons",
		"check_boxes", "boolean", "date", "datetime", "decimal", "integer", "float",
		"date_pair", "datetime_pair"}

	for _, ft := range valid {
		if !IsValidFieldType(ft) {
			t.Errorf("expected %q to be valid", ft)
		}
	}

	invalid := []string{"", "unknown", "blob", "json"}
	for _, ft := range invalid {
		if IsValidFieldType(ft) {
			t.Errorf("expected %q to be invalid", ft)
		}
	}
}

func TestFieldTypeRegistry_MatchesRails(t *testing.T) {
	// Verify key type mappings match Rails BASE_FIELD_TYPES
	tests := []struct {
		as         string
		columnType string
	}{
		{"string", "string"},
		{"text", "text"},
		{"email", "string"},
		{"select", "string"},
		{"check_boxes", "text"},
		{"boolean", "boolean"},
		{"date", "date"},
		{"datetime", "timestamp"},
		{"integer", "integer"},
		{"float", "float"},
		{"decimal", "decimal"},
		{"date_pair", "date"},
		{"datetime_pair", "timestamp"},
	}

	for _, tt := range tests {
		info, ok := FieldTypeRegistry[tt.as]
		if !ok {
			t.Errorf("missing field type %q", tt.as)
			continue
		}
		if info.ColumnType != tt.columnType {
			t.Errorf("field type %q: expected column type %q, got %q", tt.as, tt.columnType, info.ColumnType)
		}
	}
}

func TestFieldTypeRegistry_DecimalPrecision(t *testing.T) {
	info := FieldTypeRegistry["decimal"]
	if info.Precision != 15 {
		t.Errorf("expected decimal precision 15, got %d", info.Precision)
	}
	if info.Scale != 2 {
		t.Errorf("expected decimal scale 2, got %d", info.Scale)
	}
}

func TestValidateFieldValue_Required(t *testing.T) {
	field := model.Field{Name: "cf_company", Required: true}

	// Blank value should fail
	errs := ValidateFieldValue(field, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Message != "is required" {
		t.Errorf("expected 'is required', got %q", errs[0].Message)
	}

	// Empty string should fail
	errs = ValidateFieldValue(field, "")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for empty string, got %d", len(errs))
	}

	// Whitespace-only should fail
	errs = ValidateFieldValue(field, "   ")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for whitespace, got %d", len(errs))
	}

	// Non-blank value should pass
	errs = ValidateFieldValue(field, "Acme Corp")
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestValidateFieldValue_NotRequired_Blank(t *testing.T) {
	field := model.Field{Name: "cf_optional", Required: false}

	errs := ValidateFieldValue(field, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors for optional blank field, got %d", len(errs))
	}

	errs = ValidateFieldValue(field, "")
	if len(errs) != 0 {
		t.Errorf("expected no errors for optional empty string, got %d", len(errs))
	}
}

func TestValidateFieldValue_Minlength(t *testing.T) {
	field := model.Field{Name: "cf_code", Minlength: 3}

	errs := ValidateFieldValue(field, "ab")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Field != "cf_code" {
		t.Errorf("expected field cf_code, got %q", errs[0].Field)
	}

	errs = ValidateFieldValue(field, "abc")
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for valid length, got %d", len(errs))
	}
}

func TestValidateFieldValue_Maxlength(t *testing.T) {
	max := 5
	field := model.Field{Name: "cf_short", Maxlength: &max}

	errs := ValidateFieldValue(field, "toolong")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	errs = ValidateFieldValue(field, "ok")
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}

	errs = ValidateFieldValue(field, "exact")
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for exact length, got %d", len(errs))
	}
}

func TestValidateFieldValue_MinAndMaxlength(t *testing.T) {
	max := 10
	field := model.Field{Name: "cf_range", Minlength: 3, Maxlength: &max}

	// Too short
	errs := ValidateFieldValue(field, "ab")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for too short, got %d", len(errs))
	}

	// Too long
	errs = ValidateFieldValue(field, "this is way too long")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for too long, got %d", len(errs))
	}

	// Just right
	errs = ValidateFieldValue(field, "perfect")
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestValidateFieldValue_RequiredSkipsLengthChecks(t *testing.T) {
	field := model.Field{Name: "cf_req", Required: true, Minlength: 3}

	// Required + blank should only give required error, not minlength
	errs := ValidateFieldValue(field, "")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Message != "is required" {
		t.Errorf("expected 'is required', got %q", errs[0].Message)
	}
}

func TestValidationError_ErrorString(t *testing.T) {
	err := ValidationError{Field: "cf_name", Message: "is required"}
	if err.Error() != "cf_name: is required" {
		t.Errorf("unexpected error string: %q", err.Error())
	}
}
