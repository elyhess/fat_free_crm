package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
)

// SettingsHandler provides admin endpoints for application settings.
type SettingsHandler struct {
	db *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// setting represents a row in the settings table.
type setting struct {
	ID    int64  `gorm:"primaryKey" json:"id"`
	Name  string `gorm:"column:name" json:"name"`
	Value string `gorm:"column:value" json:"value"`
}

func (setting) TableName() string { return "settings" }

// requireAdmin extracts claims and checks admin flag.
func (h *SettingsHandler) requireAdmin(w http.ResponseWriter, r *http.Request) *auth.Claims {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return nil
	}
	if !claims.Admin {
		writeError(w, http.StatusForbidden, "admin access required")
		return nil
	}
	return claims
}

// GetSettings returns all settings as a JSON object.
// GET /admin/settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var rows []setting
	if err := h.db.Order("name").Find(&rows).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	result := make(map[string]interface{})
	for _, row := range rows {
		result[row.Name] = deserializeYAMLValue(row.Value)
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateSettings bulk-updates settings from a JSON object.
// PUT /admin/settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var incoming map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	tx := h.db.Begin()
	for name, value := range incoming {
		if name == "" {
			continue
		}
		serialized, err := serializeToYAML(value)
		if err != nil {
			tx.Rollback()
			writeError(w, http.StatusBadRequest, fmt.Sprintf("cannot serialize setting %q", name))
			return
		}

		var existing setting
		result := tx.Where("name = ?", name).First(&existing)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "database error")
			return
		}

		if result.Error == gorm.ErrRecordNotFound {
			// Insert new setting
			s := setting{Name: name, Value: serialized}
			if err := tx.Create(&s).Error; err != nil {
				tx.Rollback()
				writeError(w, http.StatusInternalServerError, "failed to create setting")
				return
			}
		} else {
			// Update existing
			if err := tx.Model(&existing).Update("value", serialized).Error; err != nil {
				tx.Rollback()
				writeError(w, http.StatusInternalServerError, "failed to update setting")
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit settings")
		return
	}

	// Return updated settings
	h.GetSettings(w, r)
}

// deserializeYAMLValue converts a Rails YAML-serialized value column into a Go value.
// Rails `serialize :value` stores arbitrary Ruby objects as YAML.
func deserializeYAMLValue(raw string) interface{} {
	if raw == "" {
		return nil
	}

	var val interface{}
	if err := yaml.Unmarshal([]byte(raw), &val); err != nil {
		// If YAML parsing fails, return as raw string
		return raw
	}

	return normalizeYAMLValue(val)
}

// normalizeYAMLValue converts yaml.v3 types to JSON-friendly Go types.
// Strips Ruby symbol colons from map keys and string values.
func normalizeYAMLValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[stripSymbolColon(k)] = normalizeYAMLValue(v)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v := range val {
			result[stripSymbolColon(fmt.Sprintf("%v", k))] = normalizeYAMLValue(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = normalizeYAMLValue(v)
		}
		return result
	case string:
		// Keep Ruby symbol strings (e.g. ":not_allowed") as-is for enum values
		return val
	default:
		return val
	}
}

// stripSymbolColon removes leading ":" from Ruby symbol keys in YAML maps.
func stripSymbolColon(s string) string {
	if strings.HasPrefix(s, ":") {
		return s[1:]
	}
	return s
}

// serializeToYAML converts a Go value to YAML string for storage.
// Produces YAML compatible with Rails `serialize :value`.
func serializeToYAML(v interface{}) (string, error) {
	// Handle scalar string values with Rails document marker format
	if s, ok := v.(string); ok {
		if strings.HasPrefix(s, ":") {
			return fmt.Sprintf("--- %s\n", s), nil
		}
		// Use yaml.Marshal for proper quoting, then prepend document marker
		out, err := yaml.Marshal(s)
		if err != nil {
			return "", err
		}
		return "--- " + string(out), nil
	}

	// Handle booleans and numbers with document marker
	switch v.(type) {
	case bool, int, int64, float64:
		out, err := yaml.Marshal(v)
		if err != nil {
			return "", err
		}
		return "--- " + string(out), nil
	}

	// Arrays and maps: yaml.Marshal produces valid standalone YAML
	out, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return "---\n" + string(out), nil
}
