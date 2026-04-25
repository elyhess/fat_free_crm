package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// AdminHandler provides admin-only management endpoints.
type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// requireAdmin extracts claims and checks admin flag. Returns claims or writes error.
func (h *AdminHandler) requireAdmin(w http.ResponseWriter, r *http.Request) *auth.Claims {
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

// --- User Management ---

type createUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Admin    bool    `json:"admin"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	Phone     string `json:"phone"`
	Mobile    string `json:"mobile"`
}

type updateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"`
	Admin     *bool   `json:"admin,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Title     *string `json:"title,omitempty"`
	Company   *string `json:"company,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Mobile    *string `json:"mobile,omitempty"`
}

// CreateUser creates a new user.
// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Username == "" || req.Email == "" {
		writeError(w, http.StatusUnprocessableEntity, "username and email are required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusUnprocessableEntity, "password is required")
		return
	}

	// Validate password complexity
	if complexErr := auth.ValidatePasswordComplexity(req.Password); complexErr.HasErrors() {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"error":    "password does not meet complexity requirements",
			"details":  complexErr.Messages(),
		})
		return
	}

	// Check uniqueness
	var count int64
	h.db.Model(&model.User{}).Where("lower(username) = lower(?) OR lower(email) = lower(?)", req.Username, req.Email).Count(&count)
	if count > 0 {
		writeError(w, http.StatusConflict, "username or email already exists")
		return
	}

	// Generate salt and hash password
	salt := generateSalt()
	encrypted := auth.DigestPassword(req.Password, salt, auth.DefaultStretches)

	now := time.Now()
	user := model.User{
		Username:          req.Username,
		Email:             req.Email,
		EncryptedPassword: encrypted,
		PasswordSalt:      salt,
		Admin:             req.Admin,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Title:             req.Title,
		Company:           req.Company,
		Phone:             req.Phone,
		Mobile:            req.Mobile,
		ConfirmedAt:       &now, // Admin-created users are auto-confirmed
	}

	if err := h.db.Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// UpdateUser updates user attributes.
// PUT /api/v1/admin/users/{id}
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var user model.User
	if err := h.db.Where("deleted_at IS NULL").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Admin != nil {
		updates["admin"] = *req.Admin
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Mobile != nil {
		updates["mobile"] = *req.Mobile
	}

	// Handle password change
	if req.Password != nil && *req.Password != "" {
		if complexErr := auth.ValidatePasswordComplexity(*req.Password); complexErr.HasErrors() {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
				"error":   "password does not meet complexity requirements",
				"details": complexErr.Messages(),
			})
			return
		}
		salt := generateSalt()
		updates["password_salt"] = salt
		updates["encrypted_password"] = auth.DigestPassword(*req.Password, salt, auth.DefaultStretches)
	}

	if len(updates) == 0 {
		writeJSON(w, http.StatusOK, user)
		return
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// Reload
	h.db.First(&user, id)
	writeJSON(w, http.StatusOK, user)
}

// DeleteUser deletes a user.
// DELETE /api/v1/admin/users/{id}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAdmin(w, r)
	if claims == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if id == claims.UserID {
		writeError(w, http.StatusUnprocessableEntity, "cannot delete yourself")
		return
	}

	var user model.User
	if err := h.db.Where("deleted_at IS NULL").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Check if user has related assets (matches Rails destroyable? check)
	assetTables := []struct {
		table string
		col   string
	}{
		{"accounts", "user_id"},
		{"campaigns", "user_id"},
		{"leads", "user_id"},
		{"contacts", "user_id"},
		{"opportunities", "user_id"},
		{"comments", "user_id"},
		{"tasks", "user_id"},
	}
	for _, at := range assetTables {
		var c int64
		h.db.Table(at.table).Where(at.col+" = ?", id).Count(&c)
		if c > 0 {
			writeError(w, http.StatusUnprocessableEntity, "cannot delete user with related "+at.table)
			return
		}
	}

	h.db.Delete(&user)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// SuspendUser suspends a user.
// PUT /api/v1/admin/users/{id}/suspend
func (h *AdminHandler) SuspendUser(w http.ResponseWriter, r *http.Request) {
	claims := h.requireAdmin(w, r)
	if claims == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if id == claims.UserID {
		writeError(w, http.StatusUnprocessableEntity, "cannot suspend yourself")
		return
	}

	var user model.User
	if err := h.db.Where("deleted_at IS NULL").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	now := time.Now()
	h.db.Model(&user).Update("suspended_at", &now)
	user.SuspendedAt = &now
	writeJSON(w, http.StatusOK, user)
}

// ReactivateUser reactivates a suspended user.
// PUT /api/v1/admin/users/{id}/reactivate
func (h *AdminHandler) ReactivateUser(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var user model.User
	if err := h.db.Where("deleted_at IS NULL").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.db.Model(&user).Update("suspended_at", nil)
	user.SuspendedAt = nil
	writeJSON(w, http.StatusOK, user)
}

// --- Group Management ---

type groupRequest struct {
	Name    string  `json:"name"`
	UserIDs []int64 `json:"user_ids"`
}

// ListGroups returns all groups.
// GET /api/v1/admin/groups
func (h *AdminHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var groups []model.Group
	h.db.Order("name ASC").Find(&groups)
	writeJSON(w, http.StatusOK, groups)
}

// CreateGroup creates a new group.
// POST /api/v1/admin/groups
func (h *AdminHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req groupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "name is required")
		return
	}

	// Check uniqueness
	var count int64
	h.db.Model(&model.Group{}).Where("name = ?", req.Name).Count(&count)
	if count > 0 {
		writeError(w, http.StatusConflict, "group name already exists")
		return
	}

	group := model.Group{Name: req.Name}
	if err := h.db.Create(&group).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create group")
		return
	}

	// Add users if specified
	if len(req.UserIDs) > 0 {
		for _, uid := range req.UserIDs {
			h.db.Exec("INSERT INTO groups_users (group_id, user_id) VALUES (?, ?)", group.ID, uid)
		}
	}

	writeJSON(w, http.StatusCreated, group)
}

// UpdateGroup updates a group.
// PUT /api/v1/admin/groups/{id}
func (h *AdminHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var group model.Group
	if err := h.db.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req groupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Name != "" {
		group.Name = req.Name
	}
	h.db.Save(&group)

	// Replace user memberships if provided
	if req.UserIDs != nil {
		h.db.Exec("DELETE FROM groups_users WHERE group_id = ?", group.ID)
		for _, uid := range req.UserIDs {
			h.db.Exec("INSERT INTO groups_users (group_id, user_id) VALUES (?, ?)", group.ID, uid)
		}
	}

	writeJSON(w, http.StatusOK, group)
}

// DeleteGroup deletes a group.
// DELETE /api/v1/admin/groups/{id}
func (h *AdminHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var group model.Group
	if err := h.db.First(&group, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Remove group memberships and permissions
	h.db.Exec("DELETE FROM groups_users WHERE group_id = ?", id)
	h.db.Where("group_id = ?", id).Delete(&model.Permission{})
	h.db.Delete(&group)

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Field Group Management ---

type fieldGroupRequest struct {
	KlassName string `json:"klass_name"`
	Label     string `json:"label"`
	Name      string `json:"name"`
	Position  *int   `json:"position,omitempty"`
	TagID     *int64 `json:"tag_id,omitempty"`
}

// CreateFieldGroup creates a new field group.
// POST /api/v1/admin/field_groups
func (h *AdminHandler) CreateFieldGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	var req fieldGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Label == "" || req.KlassName == "" {
		writeError(w, http.StatusUnprocessableEntity, "label and klass_name are required")
		return
	}

	fg := model.FieldGroup{
		KlassName: req.KlassName,
		Label:     req.Label,
		Name:      req.Name,
		TagID:     req.TagID,
	}
	if req.Position != nil {
		fg.Position = *req.Position
	}

	if err := h.db.Create(&fg).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create field group")
		return
	}
	writeJSON(w, http.StatusCreated, fg)
}

// UpdateFieldGroup updates a field group.
// PUT /api/v1/admin/field_groups/{id}
func (h *AdminHandler) UpdateFieldGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var fg model.FieldGroup
	if err := h.db.First(&fg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "field group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req fieldGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Label != "" {
		updates["label"] = req.Label
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.KlassName != "" {
		updates["klass_name"] = req.KlassName
	}
	if req.Position != nil {
		updates["position"] = *req.Position
	}
	if req.TagID != nil {
		updates["tag_id"] = *req.TagID
	}

	if len(updates) > 0 {
		h.db.Model(&fg).Updates(updates)
		h.db.First(&fg, id)
	}

	writeJSON(w, http.StatusOK, fg)
}

// DeleteFieldGroup deletes a field group.
// DELETE /api/v1/admin/field_groups/{id}
func (h *AdminHandler) DeleteFieldGroup(w http.ResponseWriter, r *http.Request) {
	if h.requireAdmin(w, r) == nil {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var fg model.FieldGroup
	if err := h.db.First(&fg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, http.StatusNotFound, "field group not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.db.Delete(&fg)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Helpers ---

// generateSalt creates a random hex-encoded salt for password hashing.
func generateSalt() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate salt: " + err.Error())
	}
	return hex.EncodeToString(b)
}
