package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// --- Comments ---

type commentRequest struct {
	Comment string `json:"comment"`
	Private bool   `json:"private"`
}

// CreateComment adds a comment to an entity.
// POST /api/v1/{entity}/{id}/comments
func (h *WriteHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	className, ok := validPolymorphicTypes[entitySlug]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	entityID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req commentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Comment == "" {
		writeError(w, http.StatusUnprocessableEntity, "comment is required")
		return
	}

	comment := model.Comment{
		UserID:          claims.UserID,
		CommentableID:   entityID,
		CommentableType: className,
		Comment:         req.Comment,
		Private:         req.Private,
		State:           "Expanded",
	}

	if err := h.db.Create(&comment).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create comment")
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

// DeleteComment removes a comment.
// DELETE /api/v1/comments/{id}
func (h *WriteHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var comment model.Comment
	if err := h.db.First(&comment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if !claims.Admin && comment.UserID != claims.UserID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	h.db.Delete(&comment)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Tags ---

type tagRequest struct {
	Name string `json:"name"`
}

// AddTag adds a tag to an entity.
// POST /api/v1/{entity}/{id}/tags
func (h *WriteHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	className, ok := validPolymorphicTypes[entitySlug]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	entityID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	// Find or create tag
	var tag model.Tag
	result := h.db.Where("name = ?", req.Name).First(&tag)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tag = model.Tag{Name: req.Name}
		if err := h.db.Create(&tag).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create tag")
			return
		}
	}

	// Check if tagging already exists
	var count int64
	h.db.Model(&model.Tagging{}).
		Where("tag_id = ? AND taggable_id = ? AND taggable_type = ? AND context = 'tags'",
			tag.ID, entityID, className).
		Count(&count)
	if count > 0 {
		writeJSON(w, http.StatusOK, tag)
		return
	}

	// Create tagging
	tagging := model.Tagging{
		TagID:        tag.ID,
		TaggableID:   entityID,
		TaggableType: className,
		TaggerID:     &claims.UserID,
		Context:      "tags",
	}
	if err := h.db.Create(&tagging).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add tag")
		return
	}

	// Increment taggings_count
	h.db.Model(&tag).Update("taggings_count", gorm.Expr("taggings_count + 1"))

	writeJSON(w, http.StatusCreated, tag)
}

// RemoveTag removes a tag from an entity.
// DELETE /api/v1/{entity}/{id}/tags/{tag_id}
func (h *WriteHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	className, ok := validPolymorphicTypes[entitySlug]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	entityID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tagID, err := strconv.ParseInt(r.PathValue("tag_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tag_id")
		return
	}

	result := h.db.
		Where("tag_id = ? AND taggable_id = ? AND taggable_type = ?", tagID, entityID, className).
		Delete(&model.Tagging{})
	if result.RowsAffected > 0 {
		h.db.Model(&model.Tag{}).Where("id = ?", tagID).
			Update("taggings_count", gorm.Expr("CASE WHEN taggings_count > 0 THEN taggings_count - 1 ELSE 0 END"))
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Addresses ---

type addressRequest struct {
	Street1     *string `json:"street1,omitempty"`
	Street2     *string `json:"street2,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	Zipcode     *string `json:"zipcode,omitempty"`
	Country     *string `json:"country,omitempty"`
	AddressType *string `json:"address_type,omitempty"`
}

// CreateAddress adds an address to an entity.
// POST /api/v1/{entity}/{id}/addresses
func (h *WriteHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	entitySlug := r.PathValue("entity")
	className, ok := validPolymorphicTypes[entitySlug]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return
	}

	entityID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req addressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	addr := model.Address{
		Street1:         req.Street1,
		Street2:         req.Street2,
		City:            req.City,
		State:           req.State,
		Zipcode:         req.Zipcode,
		Country:         req.Country,
		AddressType:     req.AddressType,
		AddressableID:   entityID,
		AddressableType: className,
	}

	if err := h.db.Create(&addr).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create address")
		return
	}
	writeJSON(w, http.StatusCreated, addr)
}

// DeleteAddress removes an address.
// DELETE /api/v1/addresses/{id}
func (h *WriteHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var addr model.Address
	if err := h.db.Where("id = ? AND deleted_at IS NULL", id).First(&addr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.db.Delete(&addr)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
