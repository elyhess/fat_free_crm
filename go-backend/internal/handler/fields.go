package handler

import (
	"encoding/json"
	"net/http"

	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

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

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
