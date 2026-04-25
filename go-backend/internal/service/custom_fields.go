package service

import (
	"fmt"
	"strings"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
)

// FieldTypeInfo describes how a field type maps to a DB column type.
type FieldTypeInfo struct {
	ColumnType string
	Precision  int
	Scale      int
}

// FieldTypeRegistry maps the `as` value to its DB column type,
// matching Rails BASE_FIELD_TYPES.
var FieldTypeRegistry = map[string]FieldTypeInfo{
	"string":        {ColumnType: "string"},
	"text":          {ColumnType: "text"},
	"email":         {ColumnType: "string"},
	"url":           {ColumnType: "string"},
	"tel":           {ColumnType: "string"},
	"select":        {ColumnType: "string"},
	"radio_buttons": {ColumnType: "string"},
	"check_boxes":   {ColumnType: "text"},
	"boolean":       {ColumnType: "boolean"},
	"date":          {ColumnType: "date"},
	"datetime":      {ColumnType: "timestamp"},
	"decimal":       {ColumnType: "decimal", Precision: 15, Scale: 2},
	"integer":       {ColumnType: "integer"},
	"float":         {ColumnType: "float"},
	"date_pair":     {ColumnType: "date"},
	"datetime_pair": {ColumnType: "timestamp"},
}

// ValidationError represents a field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// CustomFieldService handles custom field operations.
type CustomFieldService struct {
	repo *repository.FieldGroupRepository
}

func NewCustomFieldService(repo *repository.FieldGroupRepository) *CustomFieldService {
	return &CustomFieldService{repo: repo}
}

// GetFieldGroups returns field groups with fields for an entity type.
func (s *CustomFieldService) GetFieldGroups(entityType string) ([]model.FieldGroup, error) {
	return s.repo.FindByEntity(entityType)
}

// GetCustomFields returns only custom fields (cf_* columns) for an entity type.
func (s *CustomFieldService) GetCustomFields(entityType string) ([]model.Field, error) {
	return s.repo.FindCustomFieldsByEntity(entityType)
}

// ValidateFieldValue validates a single value against its field definition.
func ValidateFieldValue(field model.Field, value interface{}) []ValidationError {
	var errs []ValidationError

	strVal := fmt.Sprintf("%v", value)
	isBlank := value == nil || strings.TrimSpace(strVal) == ""

	if field.Required && isBlank {
		errs = append(errs, ValidationError{
			Field:   field.Name,
			Message: "is required",
		})
		return errs
	}

	if isBlank {
		return errs
	}

	if field.Minlength > 0 && len(strVal) < field.Minlength {
		errs = append(errs, ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("is too short (minimum %d characters)", field.Minlength),
		})
	}

	if field.Maxlength != nil && *field.Maxlength > 0 && len(strVal) > *field.Maxlength {
		errs = append(errs, ValidationError{
			Field:   field.Name,
			Message: fmt.Sprintf("is too long (maximum %d characters)", *field.Maxlength),
		})
	}

	return errs
}

// ValidateCustomFields validates a map of custom field values against their definitions.
func (s *CustomFieldService) ValidateCustomFields(entityType string, values map[string]interface{}) ([]ValidationError, error) {
	fields, err := s.repo.FindCustomFieldsByEntity(entityType)
	if err != nil {
		return nil, err
	}

	var allErrors []ValidationError
	for _, field := range fields {
		val := values[field.Name]
		if fieldErrors := ValidateFieldValue(field, val); len(fieldErrors) > 0 {
			allErrors = append(allErrors, fieldErrors...)
		}
	}

	return allErrors, nil
}

// IsValidFieldType checks if an `as` value is a known field type.
func IsValidFieldType(as string) bool {
	_, ok := FieldTypeRegistry[as]
	return ok
}
