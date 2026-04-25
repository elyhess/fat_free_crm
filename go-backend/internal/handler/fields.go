package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

// entityAliases maps lowercase plural and singular forms to PascalCase entity names.
// Built from validPolymorphicTypes (plural) plus singular variants.
var entityAliases = func() map[string]string {
	m := make(map[string]string)
	// From validPolymorphicTypes: "accounts" -> "Account", etc.
	for plural, pascal := range validPolymorphicTypes {
		m[plural] = pascal
	}
	// Add singular lowercase forms.
	m["account"] = "Account"
	m["contact"] = "Contact"
	m["lead"] = "Lead"
	m["opportunity"] = "Opportunity"
	m["campaign"] = "Campaign"
	m["task"] = "Task"
	return m
}()

type FieldsHandler struct {
	svc *service.CustomFieldService
}

func NewFieldsHandler(svc *service.CustomFieldService) *FieldsHandler {
	return &FieldsHandler{svc: svc}
}

type fieldGroupsResponse struct {
	EntityType  string             `json:"entity_type"`
	FieldGroups []model.FieldGroup `json:"field_groups"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ListFieldGroups returns field groups and their fields for an entity type.
// GET /api/v1/field_groups?entity=Account
func (h *FieldsHandler) ListFieldGroups(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity")
	if entityType == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "entity query parameter is required"})
		return
	}

	// Normalize: accept PascalCase ("Account"), lowercase plural ("accounts"),
	// or lowercase singular ("account").
	entityType = normalizeEntityType(entityType)

	if _, ok := model.ValidEntityTypes[entityType]; !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid entity type: " + entityType})
		return
	}

	groups, err := h.svc.GetFieldGroups(entityType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch field groups"})
		return
	}

	writeJSON(w, http.StatusOK, fieldGroupsResponse{
		EntityType:  entityType,
		FieldGroups: groups,
	})
}

// normalizeEntityType converts lowercase plural/singular entity names to PascalCase.
// If the input is already a valid PascalCase name it is returned as-is.
func normalizeEntityType(raw string) string {
	// Already valid PascalCase?
	if _, ok := model.ValidEntityTypes[raw]; ok {
		return raw
	}
	// Try lowercase lookup (handles both plural and singular).
	if pascal, ok := entityAliases[strings.ToLower(raw)]; ok {
		return pascal
	}
	return raw // return as-is; validation downstream will reject it
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
