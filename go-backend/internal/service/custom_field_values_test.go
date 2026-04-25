package service

import (
	"testing"
)

func TestGenerateColumnName(t *testing.T) {
	tests := []struct {
		label    string
		existing map[string]bool
		expected string
	}{
		{"My Field", map[string]bool{}, "cf_my_field"},
		{"Email", map[string]bool{"cf_email": true}, "cf_email_2"},
		{"Test-Field", map[string]bool{}, "cf_test_field"},
		{"  Leading Spaces  ", map[string]bool{}, "cf_leading_spaces"},
		{"123 Numbers", map[string]bool{}, "cf_123_numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			got := GenerateColumnName(tt.label, tt.existing)
			if got != tt.expected {
				t.Errorf("GenerateColumnName(%q) = %q, want %q", tt.label, got, tt.expected)
			}
		})
	}
}

func TestIsValidColumnName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"cf_test", true},
		{"cf_my_field_2", true},
		{"cf_123", true},
		{"name", false},                 // doesn't start with cf_
		{"cf_", false},                  // nothing after cf_
		{"cf_Test", false},              // uppercase
		{"cf_field; DROP TABLE", false}, // SQL injection
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidColumnName(tt.name)
			if got != tt.valid {
				t.Errorf("IsValidColumnName(%q) = %v, want %v", tt.name, got, tt.valid)
			}
		})
	}
}
