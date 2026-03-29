package service

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// CustomFieldValueService reads and writes dynamic cf_* column values.
type CustomFieldValueService struct {
	db   *gorm.DB
	repo *CustomFieldService
}

func NewCustomFieldValueService(db *gorm.DB, cfSvc *CustomFieldService) *CustomFieldValueService {
	return &CustomFieldValueService{db: db, repo: cfSvc}
}

// ReadCustomFieldValues returns a map of cf_* column names to their values for a specific entity record.
func (s *CustomFieldValueService) ReadCustomFieldValues(entityType string, entityID int64) (map[string]interface{}, error) {
	tableName, ok := model.ValidEntityTypes[entityType]
	if !ok {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	fields, err := s.repo.GetCustomFields(entityType)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return map[string]interface{}{}, nil
	}

	// Build SELECT for just the cf_* columns
	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = f.Name
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(cols, ", "), tableName)

	row := s.db.Raw(query, entityID).Row()
	if row == nil {
		return nil, fmt.Errorf("record not found")
	}

	// Scan into interface slice
	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	if err := row.Scan(ptrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		result[col] = normalizeDBValue(values[i])
	}

	return result, nil
}

// ReadCustomFieldValuesForList returns custom field values for multiple entity records.
// Returns a map of entityID -> { cf_name: value }.
func (s *CustomFieldValueService) ReadCustomFieldValuesForList(entityType string, entityIDs []int64) (map[int64]map[string]interface{}, error) {
	tableName, ok := model.ValidEntityTypes[entityType]
	if !ok {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	fields, err := s.repo.GetCustomFields(entityType)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 || len(entityIDs) == 0 {
		return map[int64]map[string]interface{}{}, nil
	}

	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = f.Name
	}

	// Build placeholders for IDs
	placeholders := make([]string, len(entityIDs))
	args := make([]interface{}, len(entityIDs))
	for i, id := range entityIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id, %s FROM %s WHERE id IN (%s)",
		strings.Join(cols, ", "), tableName, strings.Join(placeholders, ","))

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]map[string]interface{})
	for rows.Next() {
		values := make([]interface{}, len(cols)+1)
		ptrs := make([]interface{}, len(cols)+1)
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		id, _ := toInt64(values[0])
		cfValues := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			cfValues[col] = normalizeDBValue(values[i+1])
		}
		result[id] = cfValues
	}

	return result, nil
}

// WriteCustomFieldValues writes cf_* column values for an entity record via raw SQL UPDATE.
func (s *CustomFieldValueService) WriteCustomFieldValues(entityType string, entityID int64, values map[string]interface{}) error {
	tableName, ok := model.ValidEntityTypes[entityType]
	if !ok {
		return fmt.Errorf("invalid entity type: %s", entityType)
	}

	if len(values) == 0 {
		return nil
	}

	// Validate all keys are cf_* columns
	fields, err := s.repo.GetCustomFields(entityType)
	if err != nil {
		return err
	}

	validCols := make(map[string]bool, len(fields))
	for _, f := range fields {
		validCols[f.Name] = true
	}

	setClauses := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))
	for col, val := range values {
		if !validCols[col] {
			return fmt.Errorf("unknown custom field: %s", col)
		}
		if !IsValidColumnName(col) {
			return fmt.Errorf("invalid column name: %s", col)
		}
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	args = append(args, entityID)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(setClauses, ", "))

	return s.db.Exec(query, args...).Error
}

// normalizeDBValue converts database driver types to JSON-friendly Go types.
func normalizeDBValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	// Convert []byte to string
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case int:
		return int64(val), true
	case float64:
		return int64(val), true
	default:
		return 0, false
	}
}

// IsValidColumnName ensures a column name is safe for SQL interpolation.
var validColumnRe = regexp.MustCompile(`^cf_[a-z0-9_]+$`)

func IsValidColumnName(name string) bool {
	return validColumnRe.MatchString(name)
}

// GenerateColumnName creates a cf_* column name from a label, avoiding collisions.
func GenerateColumnName(label string, existingNames map[string]bool) string {
	// Lowercase, replace non-alphanumeric with underscore
	name := strings.ToLower(label)
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	name = "cf_" + name

	if !existingNames[name] {
		return name
	}

	// Add numeric suffix for collisions
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", name, i)
		if !existingNames[candidate] {
			return candidate
		}
	}
}
