package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// AdminFieldsHandler provides admin CRUD for custom field definitions.
type AdminFieldsHandler struct {
	db    *gorm.DB
	cfSvc *service.CustomFieldService
}

func NewAdminFieldsHandler(db *gorm.DB, cfSvc *service.CustomFieldService) *AdminFieldsHandler {
	return &AdminFieldsHandler{db: db, cfSvc: cfSvc}
}

func (h *AdminFieldsHandler) requireAdmin(w http.ResponseWriter, r *http.Request) *auth.Claims {
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

type createFieldRequest struct {
	FieldGroupID int64   `json:"field_group_id"`
	Label        string  `json:"label"`
	As           string  `json:"as"`
	Hint         string  `json:"hint,omitempty"`
	Placeholder  string  `json:"placeholder,omitempty"`
	Required     bool    `json:"required"`
	Disabled     bool    `json:"disabled"`
	Minlength    int     `json:"minlength"`
	Maxlength    *int    `json:"maxlength,omitempty"`
	Collection   string  `json:"collection,omitempty"`
	Position     *int    `json:"position,omitempty"`
}

type updateFieldRequest struct {
	Label       *string `json:"label,omitempty"`
	As          *string `json:"as,omitempty"`
	Hint        *string `json:"hint,omitempty"`
	Placeholder *string `json:"placeholder,omitempty"`
	Required    *bool   `json:"required,omitempty"`
	Disabled    *bool   `json:"disabled,omitempty"`
	Minlength   *int    `json:"minlength,omitempty"`
	Maxlength   *int    `json:"maxlength,omitempty"`
	Collection  *string `json:"collection,omitempty"`
}

type sortFieldsRequest struct {
	FieldIDs []int64 `json:"field_ids"`
}

// CreateField creates a new custom field definition and the corresponding database column.
// POST /admin/fields
func (h *AdminFieldsHandler) CreateField(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req createFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Label == "" {
		writeError(w, http.StatusUnprocessableEntity, "label is required")
		return
	}
	if !service.IsValidFieldType(req.As) {
		writeError(w, http.StatusUnprocessableEntity, "invalid field type: "+req.As)
		return
	}

	// Find the field group to get entity type
	var fg model.FieldGroup
	if err := h.db.First(&fg, req.FieldGroupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "field group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tableName, ok := model.ValidEntityTypes[fg.KlassName]
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "invalid entity type on field group")
		return
	}

	// Get existing cf_* column names for collision avoidance
	existingNames := h.getExistingColumnNames(fg.KlassName)

	columnName := service.GenerateColumnName(req.Label, existingNames)

	// Determine position
	position := 0
	if req.Position != nil {
		position = *req.Position
	} else {
		// Auto-increment: max position + 1
		var maxPos *int
		h.db.Model(&model.Field{}).Where("field_group_id = ?", req.FieldGroupID).
			Select("MAX(position)").Scan(&maxPos)
		if maxPos != nil {
			position = *maxPos + 1
		}
	}

	// Create the field record
	field := model.Field{
		Type:         "CustomField",
		FieldGroupID: req.FieldGroupID,
		Name:         columnName,
		Label:        req.Label,
		As:           req.As,
		Hint:         req.Hint,
		Placeholder:  req.Placeholder,
		Required:     req.Required,
		Disabled:     req.Disabled,
		Minlength:    req.Minlength,
		Maxlength:    req.Maxlength,
		Collection:   req.Collection,
		Position:     position,
	}

	tx := h.db.Begin()

	if err := tx.Create(&field).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "failed to create field")
		return
	}

	// Add the column to the entity table
	colSQL := buildAddColumnSQL(tableName, columnName, req.As)
	if err := tx.Exec(colSQL).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "failed to add column: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	writeJSON(w, http.StatusCreated, field)
}

// UpdateField updates a custom field definition.
// PUT /admin/fields/{id}
func (h *AdminFieldsHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var field model.Field
	if err := h.db.First(&field, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "field not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if field.Type == "CoreField" {
		writeError(w, http.StatusUnprocessableEntity, "cannot modify core fields")
		return
	}

	var req updateFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Label != nil {
		updates["label"] = *req.Label
	}
	if req.Hint != nil {
		updates["hint"] = *req.Hint
	}
	if req.Placeholder != nil {
		updates["placeholder"] = *req.Placeholder
	}
	if req.Required != nil {
		updates["required"] = *req.Required
	}
	if req.Disabled != nil {
		updates["disabled"] = *req.Disabled
	}
	if req.Minlength != nil {
		updates["minlength"] = *req.Minlength
	}
	if req.Maxlength != nil {
		updates["maxlength"] = *req.Maxlength
	}
	if req.Collection != nil {
		updates["collection"] = *req.Collection
	}

	// Handle type change with safety check
	if req.As != nil && *req.As != field.As {
		if !service.IsValidFieldType(*req.As) {
			writeError(w, http.StatusUnprocessableEntity, "invalid field type: "+*req.As)
			return
		}

		transition := checkTypeTransition(field.As, *req.As)
		if transition == "unsafe" {
			writeError(w, http.StatusUnprocessableEntity,
				fmt.Sprintf("cannot change field type from %s to %s (unsafe transition)", field.As, *req.As))
			return
		}

		if transition == "safe" {
			// Need to ALTER TABLE to change column type
			var fg model.FieldGroup
			h.db.First(&fg, field.FieldGroupID)
			tableName := model.ValidEntityTypes[fg.KlassName]
			colSQL := buildChangeColumnSQL(tableName, field.Name, *req.As)
			if err := h.db.Exec(colSQL).Error; err != nil {
				writeError(w, http.StatusInternalServerError, "failed to change column type: "+err.Error())
				return
			}
		}
		updates["as"] = *req.As
	}

	if len(updates) > 0 {
		h.db.Model(&field).Updates(updates)
		h.db.First(&field, id)
	}

	writeJSON(w, http.StatusOK, field)
}

// DeleteField deletes a custom field definition and drops the database column.
// DELETE /admin/fields/{id}
func (h *AdminFieldsHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var field model.Field
	if err := h.db.First(&field, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "field not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if field.Type == "CoreField" {
		writeError(w, http.StatusUnprocessableEntity, "cannot delete core fields")
		return
	}

	// Find entity table
	var fg model.FieldGroup
	if err := h.db.First(&fg, field.FieldGroupID).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	tableName := model.ValidEntityTypes[fg.KlassName]

	tx := h.db.Begin()

	// Delete the field record
	if err := tx.Delete(&field).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "failed to delete field")
		return
	}

	// Drop the column (no IF EXISTS — PostgreSQL and SQLite 3.35+ both support plain DROP COLUMN)
	dropSQL := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, field.Name)
	if err := tx.Exec(dropSQL).Error; err != nil {
		tx.Rollback()
		writeError(w, http.StatusInternalServerError, "failed to drop column: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// SortFields reorders fields within a group.
// POST /admin/fields/sort
func (h *AdminFieldsHandler) SortFields(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req sortFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	tx := h.db.Begin()
	for i, fieldID := range req.FieldIDs {
		if err := tx.Model(&model.Field{}).Where("id = ?", fieldID).Update("position", i).Error; err != nil {
			tx.Rollback()
			writeError(w, http.StatusInternalServerError, "failed to update position")
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// getExistingColumnNames returns a set of existing cf_* column names for an entity type.
func (h *AdminFieldsHandler) getExistingColumnNames(klassName string) map[string]bool {
	var fields []model.Field
	h.db.Joins("JOIN field_groups ON field_groups.id = fields.field_group_id").
		Where("field_groups.klass_name = ? AND fields.name LIKE 'cf_%'", klassName).
		Find(&fields)

	names := make(map[string]bool, len(fields))
	for _, f := range fields {
		names[f.Name] = true
	}
	return names
}

// buildAddColumnSQL generates ALTER TABLE ADD COLUMN SQL for a custom field.
func buildAddColumnSQL(tableName, columnName, fieldType string) string {
	colType := fieldTypeToSQL(fieldType)
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, colType)
}

// buildChangeColumnSQL generates ALTER TABLE ALTER COLUMN SQL for type changes.
func buildChangeColumnSQL(tableName, columnName, fieldType string) string {
	colType := fieldTypeToSQL(fieldType)
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s",
		tableName, columnName, colType, columnName, colType)
}

// fieldTypeToSQL maps field `as` values to PostgreSQL column types.
func fieldTypeToSQL(as string) string {
	info, ok := service.FieldTypeRegistry[as]
	if !ok {
		return "VARCHAR(255)"
	}
	switch info.ColumnType {
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "boolean":
		return "BOOLEAN"
	case "integer":
		return "INTEGER"
	case "float":
		return "DOUBLE PRECISION"
	case "decimal":
		return fmt.Sprintf("DECIMAL(%d,%d)", info.Precision, info.Scale)
	case "date":
		return "DATE"
	case "timestamp":
		return "TIMESTAMP"
	default:
		return "VARCHAR(255)"
	}
}

// checkTypeTransition determines if a field type change is safe.
// Returns "null" (same type), "safe", or "unsafe".
func checkTypeTransition(from, to string) string {
	fromInfo := service.FieldTypeRegistry[from]
	toInfo := service.FieldTypeRegistry[to]

	if fromInfo.ColumnType == toInfo.ColumnType {
		return "null"
	}

	// Safe bidirectional transitions
	safeGroups := [][]string{
		{"date", "timestamp"},
		{"integer", "float"},
	}
	for _, group := range safeGroups {
		fromIn := contains(group, fromInfo.ColumnType)
		toIn := contains(group, toInfo.ColumnType)
		if fromIn && toIn {
			return "safe"
		}
	}

	// Safe one-way: string -> text
	if fromInfo.ColumnType == "string" && toInfo.ColumnType == "text" {
		return "safe"
	}

	return "unsafe"
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// --- Custom field values on entities ---

// CustomFieldsForEntity returns cf_* values for an entity as a JSON-compatible map.
// Used by entity detail and list endpoints.
// GET /{entities}/{id}/custom_fields
func (h *AdminFieldsHandler) GetEntityCustomFields(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	entityType := slugToEntityType(entitySlug)
	if entityType == "" {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tableName := model.ValidEntityTypes[entityType]
	fields, err := h.cfSvc.GetCustomFields(entityType)
	if err != nil || len(fields) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{})
		return
	}

	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = f.Name
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = ?", strings.Join(cols, ", "), tableName)
	row := h.db.Raw(query, id).Row()
	if row == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	values := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}
	if err := row.Scan(ptrs...); err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	result := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		v := values[i]
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		result[col] = v
	}

	writeJSON(w, http.StatusOK, result)
}

// UpdateEntityCustomFields updates cf_* values for an entity.
// PUT /{entities}/{id}/custom_fields
func (h *AdminFieldsHandler) UpdateEntityCustomFields(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	entityType := slugToEntityType(entitySlug)
	if entityType == "" {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var incoming map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate against field definitions
	valErrors, err := h.cfSvc.ValidateCustomFields(entityType, incoming)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "validation error")
		return
	}
	if len(valErrors) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":  "validation failed",
			"errors": valErrors,
		})
		return
	}

	tableName := model.ValidEntityTypes[entityType]
	fields, err := h.cfSvc.GetCustomFields(entityType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	validCols := make(map[string]bool, len(fields))
	for _, f := range fields {
		validCols[f.Name] = true
	}

	setClauses := make([]string, 0, len(incoming))
	args := make([]interface{}, 0, len(incoming))
	for col, val := range incoming {
		if !validCols[col] {
			writeError(w, http.StatusUnprocessableEntity, "unknown custom field: "+col)
			return
		}
		if !service.IsValidColumnName(col) {
			writeError(w, http.StatusUnprocessableEntity, "invalid column name: "+col)
			return
		}
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	if len(setClauses) > 0 {
		args = append(args, id)
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(setClauses, ", "))
		if err := h.db.Exec(query, args...).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update custom fields")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// slugToEntityType converts URL slugs to entity class names.
func slugToEntityType(slug string) string {
	mapping := map[string]string{
		"accounts":      "Account",
		"contacts":      "Contact",
		"leads":         "Lead",
		"opportunities": "Opportunity",
		"campaigns":     "Campaign",
		"tasks":         "Task",
	}
	return mapping[slug]
}
