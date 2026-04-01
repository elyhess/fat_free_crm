package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// SavedSearchHandler provides CRUD for saved search presets.
type SavedSearchHandler struct {
	db *gorm.DB
}

func NewSavedSearchHandler(db *gorm.DB) *SavedSearchHandler {
	return &SavedSearchHandler{db: db}
}

type savedSearchRequest struct {
	Name    string          `json:"name"`
	Entity  string          `json:"entity"`
	Filters json.RawMessage `json:"filters"`
}

// List returns all saved searches for the current user.
// GET /saved_searches
func (h *SavedSearchHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var searches []model.SavedSearch
	h.db.Where("user_id = ?", claims.UserID).Order("name ASC").Find(&searches)
	writeJSON(w, http.StatusOK, searches)
}

// Create creates a new saved search.
// POST /saved_searches
func (h *SavedSearchHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req savedSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Entity == "" {
		writeError(w, http.StatusBadRequest, "name and entity are required")
		return
	}

	now := time.Now().UTC()
	search := model.SavedSearch{
		UserID:    claims.UserID,
		Name:      req.Name,
		Entity:    req.Entity,
		Filters:   req.Filters,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if search.Filters == nil {
		search.Filters = json.RawMessage("{}")
	}

	if err := h.db.Create(&search).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create saved search")
		return
	}
	writeJSON(w, http.StatusCreated, search)
}

// Update modifies a saved search.
// PUT /saved_searches/{id}
func (h *SavedSearchHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var search model.SavedSearch
	if err := h.db.Where("id = ? AND user_id = ?", id, claims.UserID).First(&search).Error; err != nil {
		writeError(w, http.StatusNotFound, "saved search not found")
		return
	}

	var req savedSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now().UTC()}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Entity != "" {
		updates["entity"] = req.Entity
	}
	if req.Filters != nil {
		updates["filters"] = req.Filters
	}

	h.db.Model(&search).Updates(updates)
	h.db.First(&search, id)
	writeJSON(w, http.StatusOK, search)
}

// Delete removes a saved search.
// DELETE /saved_searches/{id}
func (h *SavedSearchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", id, claims.UserID).Delete(&model.SavedSearch{})
	if result.RowsAffected == 0 {
		writeError(w, http.StatusNotFound, "saved search not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
