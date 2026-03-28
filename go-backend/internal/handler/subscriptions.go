package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
)

// SubscriptionHandler manages subscribe/unsubscribe on CRM entities.
type SubscriptionHandler struct {
	db *gorm.DB
}

func NewSubscriptionHandler(db *gorm.DB) *SubscriptionHandler {
	return &SubscriptionHandler{db: db}
}

// validSubscriptionEntities maps URL slugs to table names.
var validSubscriptionEntities = map[string]string{
	"accounts":      "accounts",
	"contacts":      "contacts",
	"leads":         "leads",
	"opportunities": "opportunities",
	"campaigns":     "campaigns",
	"tasks":         "tasks",
}

// Subscribe adds the current user to the entity's subscribed_users list.
// POST /api/v1/{entity}/{id}/subscribe
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	table, id, ok := h.parseParams(w, r)
	if !ok {
		return
	}

	userIDs, err := h.getSubscribedUsers(table, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Add user if not already subscribed
	if !containsInt(userIDs, claims.UserID) {
		userIDs = append(userIDs, claims.UserID)
		if err := h.setSubscribedUsers(table, id, userIDs); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update subscription")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscribed":       true,
		"subscribed_users": userIDs,
	})
}

// Unsubscribe removes the current user from the entity's subscribed_users list.
// POST /api/v1/{entity}/{id}/unsubscribe
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	table, id, ok := h.parseParams(w, r)
	if !ok {
		return
	}

	userIDs, err := h.getSubscribedUsers(table, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Remove user
	userIDs = removeInt(userIDs, claims.UserID)
	if err := h.setSubscribedUsers(table, id, userIDs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update subscription")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscribed":       false,
		"subscribed_users": userIDs,
	})
}

// GetSubscription returns the current subscription state for the authenticated user.
// GET /api/v1/{entity}/{id}/subscription
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	table, id, ok := h.parseParams(w, r)
	if !ok {
		return
	}

	userIDs, err := h.getSubscribedUsers(table, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscribed":       containsInt(userIDs, claims.UserID),
		"subscribed_users": userIDs,
	})
}

func (h *SubscriptionHandler) parseParams(w http.ResponseWriter, r *http.Request) (string, int64, bool) {
	entitySlug := r.PathValue("entity")
	table, ok := validSubscriptionEntities[entitySlug]
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entity type")
		return "", 0, false
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return "", 0, false
	}
	return table, id, true
}

// getSubscribedUsers reads the subscribed_users YAML text column and parses it.
func (h *SubscriptionHandler) getSubscribedUsers(table string, id int64) ([]int64, error) {
	var raw *string
	query := fmt.Sprintf("SELECT subscribed_users FROM %s WHERE id = ? AND deleted_at IS NULL", table)
	if err := h.db.Raw(query, id).Scan(&raw).Error; err != nil {
		return nil, err
	}
	// Check if the record was found — Scan on a nil result won't error but raw stays nil
	var exists int64
	h.db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ? AND deleted_at IS NULL", table), id).Scan(&exists)
	if exists == 0 {
		return nil, fmt.Errorf("not found")
	}
	if raw == nil {
		return []int64{}, nil
	}
	return parseYAMLIntArray(*raw), nil
}

// setSubscribedUsers writes the subscribed_users column as YAML.
func (h *SubscriptionHandler) setSubscribedUsers(table string, id int64, userIDs []int64) error {
	yaml := serializeYAMLIntArray(userIDs)
	query := fmt.Sprintf("UPDATE %s SET subscribed_users = ? WHERE id = ? AND deleted_at IS NULL", table)
	return h.db.Exec(query, yaml, id).Error
}

// parseYAMLIntArray parses a YAML-serialized integer array like "---\n- 1\n- 2\n".
func parseYAMLIntArray(s string) []int64 {
	if s == "" || s == "--- []\n" || s == "---\n[]" {
		return []int64{}
	}
	var result []int64
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "---" || line == "" || line == "[]" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			numStr := strings.TrimPrefix(line, "- ")
			if n, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64); err == nil {
				result = append(result, n)
			}
		}
	}
	return result
}

// serializeYAMLIntArray serializes an int64 slice to Rails-compatible YAML format.
func serializeYAMLIntArray(ids []int64) string {
	if len(ids) == 0 {
		return "--- []\n"
	}
	var b strings.Builder
	b.WriteString("---\n")
	for _, id := range ids {
		fmt.Fprintf(&b, "- %d\n", id)
	}
	return b.String()
}

func containsInt(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func removeInt(slice []int64, val int64) []int64 {
	result := make([]int64, 0, len(slice))
	for _, v := range slice {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}
