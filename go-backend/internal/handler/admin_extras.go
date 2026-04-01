package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// AdminExtrasHandler provides plugin listing and research tools CRUD.
type AdminExtrasHandler struct {
	db *gorm.DB
}

func NewAdminExtrasHandler(db *gorm.DB) *AdminExtrasHandler {
	return &AdminExtrasHandler{db: db}
}

func (h *AdminExtrasHandler) requireAdmin(w http.ResponseWriter, r *http.Request) *auth.Claims {
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

// --- Plugins ---

// ListPlugins returns an empty list. The Rails feature is a stub ("not implemented").
// GET /admin/plugins
func (h *AdminExtrasHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}
	writeJSON(w, http.StatusOK, []struct{}{})
}

// --- Research Tools ---

type researchToolRequest struct {
	Name        string `json:"name"`
	URLTemplate string `json:"url_template"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// ListResearchTools returns all research tools.
// GET /admin/research_tools
func (h *AdminExtrasHandler) ListResearchTools(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var tools []model.ResearchTool
	if err := h.db.Order("id ASC").Find(&tools).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list research tools")
		return
	}
	writeJSON(w, http.StatusOK, tools)
}

// CreateResearchTool creates a new research tool.
// POST /admin/research_tools
func (h *AdminExtrasHandler) CreateResearchTool(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req researchToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Name == "" || req.URLTemplate == "" {
		writeError(w, http.StatusBadRequest, "name and url_template are required")
		return
	}

	now := time.Now().UTC()
	tool := model.ResearchTool{
		Name:        req.Name,
		URLTemplate: req.URLTemplate,
		Enabled:     req.Enabled != nil && *req.Enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.Create(&tool).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create research tool")
		return
	}

	writeJSON(w, http.StatusCreated, tool)
}

// UpdateResearchTool updates an existing research tool.
// PUT /admin/research_tools/{id}
func (h *AdminExtrasHandler) UpdateResearchTool(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var tool model.ResearchTool
	if err := h.db.First(&tool, id).Error; err != nil {
		writeError(w, http.StatusNotFound, "research tool not found")
		return
	}

	var req researchToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{
		"updated_at": time.Now().UTC(),
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URLTemplate != "" {
		updates["url_template"] = req.URLTemplate
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if err := h.db.Model(&tool).Updates(updates).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update research tool")
		return
	}

	h.db.First(&tool, id)
	writeJSON(w, http.StatusOK, tool)
}

// DeleteResearchTool deletes a research tool.
// DELETE /admin/research_tools/{id}
func (h *AdminExtrasHandler) DeleteResearchTool(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	result := h.db.Delete(&model.ResearchTool{}, id)
	if result.Error != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete research tool")
		return
	}
	if result.RowsAffected == 0 {
		writeError(w, http.StatusNotFound, "research tool not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
