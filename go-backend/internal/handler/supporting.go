package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
)

// SupportingHandler provides endpoints for comments, addresses, tags, versions, and users.
type SupportingHandler struct {
	repo *repository.SupportingRepository
}

func NewSupportingHandler(repo *repository.SupportingRepository) *SupportingHandler {
	return &SupportingHandler{repo: repo}
}

// validPolymorphicTypes maps URL entity names to Rails class names.
var validPolymorphicTypes = map[string]string{
	"accounts":      "Account",
	"contacts":      "Contact",
	"leads":         "Lead",
	"opportunities": "Opportunity",
	"campaigns":     "Campaign",
	"tasks":         "Task",
}

func parseEntityParams(r *http.Request) (string, int64, bool) {
	entitySlug := r.PathValue("entity")
	className, ok := validPolymorphicTypes[entitySlug]
	if !ok {
		return "", 0, false
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return "", 0, false
	}
	return className, id, true
}

// ListComments returns comments for a specific entity.
// GET /api/v1/{entity}/{id}/comments
func (h *SupportingHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	className, id, ok := parseEntityParams(r)
	if !ok {
		http.Error(w, "invalid entity or id", http.StatusBadRequest)
		return
	}

	comments, err := h.repo.ListComments(className, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comments)
}

// ListAddresses returns addresses for a specific entity.
// GET /api/v1/{entity}/{id}/addresses
func (h *SupportingHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	className, id, ok := parseEntityParams(r)
	if !ok {
		http.Error(w, "invalid entity or id", http.StatusBadRequest)
		return
	}

	addresses, err := h.repo.ListAddresses(className, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(addresses)
}

// ListEntityTags returns tags for a specific entity.
// GET /api/v1/{entity}/{id}/tags
func (h *SupportingHandler) ListEntityTags(w http.ResponseWriter, r *http.Request) {
	className, id, ok := parseEntityParams(r)
	if !ok {
		http.Error(w, "invalid entity or id", http.StatusBadRequest)
		return
	}

	tags, err := h.repo.ListTagsForEntity(className, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tags)
}

// ListAllTags returns all tags.
// GET /api/v1/tags
func (h *SupportingHandler) ListAllTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.repo.ListAllTags()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tags)
}

// ListEntityVersions returns audit history for a specific entity.
// GET /api/v1/{entity}/{id}/versions
func (h *SupportingHandler) ListEntityVersions(w http.ResponseWriter, r *http.Request) {
	className, id, ok := parseEntityParams(r)
	if !ok {
		http.Error(w, "invalid entity or id", http.StatusBadRequest)
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	versions, err := h.repo.ListVersions(className, id, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(versions)
}

// ListRecentActivity returns recent activity across all entities.
// GET /api/v1/activity
func (h *SupportingHandler) ListRecentActivity(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	versions, err := h.repo.ListRecentActivity(limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(versions)
}

// ListUsers returns all users (admin only in future, currently all authenticated users).
// GET /api/v1/users
func (h *SupportingHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil || !claims.Admin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	users, err := h.repo.ListUsers()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Strip sensitive fields before returning
	type safeUser struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Title     string `json:"title,omitempty"`
		Company   string `json:"company,omitempty"`
		Phone     string `json:"phone,omitempty"`
		Mobile    string `json:"mobile,omitempty"`
		Admin     bool   `json:"admin"`
	}

	safe := make([]safeUser, len(users))
	for i, u := range users {
		safe[i] = safeUser{
			ID: u.ID, Username: u.Username, Email: u.Email,
			FirstName: u.FirstName, LastName: u.LastName,
			Title: u.Title, Company: u.Company,
			Phone: u.Phone, Mobile: u.Mobile,
			Admin: u.Admin,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(safe)
}
