package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// entityTypeMap maps URL entity names to the Rails polymorphic type.
var entityTypeMap = map[string]string{
	"accounts":      "Account",
	"contacts":      "Contact",
	"leads":         "Lead",
	"opportunities": "Opportunity",
	"campaigns":     "Campaign",
}

// EmailHandler provides endpoints for emails attached to entities.
type EmailHandler struct {
	db *gorm.DB
}

func NewEmailHandler(db *gorm.DB) *EmailHandler {
	return &EmailHandler{db: db}
}

// ListEmails returns emails for a given entity.
// GET /{entity}/{id}/emails
func (h *EmailHandler) ListEmails(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entityParam := chi.URLParam(r, "entity")
	mediatorType, ok := entityTypeMap[entityParam]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var emails []model.Email
	h.db.Where("mediator_type = ? AND mediator_id = ? AND deleted_at IS NULL", mediatorType, id).
		Order("created_at DESC").
		Find(&emails)

	writeJSON(w, http.StatusOK, emails)
}

// DeleteEmail soft-deletes an email.
// DELETE /emails/{id}
func (h *EmailHandler) DeleteEmail(w http.ResponseWriter, r *http.Request) {
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

	var email model.Email
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&email).Error; err != nil {
		writeError(w, http.StatusNotFound, "email not found")
		return
	}

	now := time.Now().UTC()
	h.db.Model(&email).Update("deleted_at", now)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// mediatorTypeFromString returns the Go mediator type from a Rails-style string.
func mediatorTypeFromString(s string) string {
	s = strings.ToLower(s)
	for _, v := range entityTypeMap {
		if strings.ToLower(v) == s {
			return v
		}
	}
	return s
}
