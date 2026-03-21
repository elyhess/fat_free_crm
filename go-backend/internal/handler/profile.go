package handler

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/middleware"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
)

// ProfileHandler provides user self-service endpoints.
type ProfileHandler struct {
	db *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

type profileResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Title     string `json:"title,omitempty"`
	Company   string `json:"company,omitempty"`
	AltEmail  string `json:"alt_email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Mobile    string `json:"mobile,omitempty"`
	Admin     bool   `json:"admin"`
}

func toProfileResponse(u model.User) profileResponse {
	return profileResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Title:     u.Title,
		Company:   u.Company,
		AltEmail:  u.AltEmail,
		Phone:     u.Phone,
		Mobile:    u.Mobile,
		Admin:     u.Admin,
	}
}

// GetProfile returns the current user's profile.
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user model.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, toProfileResponse(user))
}

type updateProfileRequest struct {
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Title     *string `json:"title,omitempty"`
	Company   *string `json:"company,omitempty"`
	AltEmail  *string `json:"alt_email,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Mobile    *string `json:"mobile,omitempty"`
}

// UpdateProfile updates the current user's profile fields.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user model.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	updates := map[string]interface{}{}
	if req.Email != nil {
		updates["email"] = *req.Email
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
	if req.AltEmail != nil {
		updates["alt_email"] = *req.AltEmail
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Mobile != nil {
		updates["mobile"] = *req.Mobile
	}

	if len(updates) > 0 {
		h.db.Model(&user).Updates(updates)
	}

	h.db.First(&user, claims.UserID)
	writeJSON(w, http.StatusOK, toProfileResponse(user))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword changes the current user's password.
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var user model.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "current_password and new_password are required")
		return
	}

	// Verify current password
	if !auth.VerifyPassword(req.CurrentPassword, user.EncryptedPassword, user.PasswordSalt, auth.DefaultStretches) {
		writeError(w, http.StatusForbidden, "current password is incorrect")
		return
	}

	// Validate new password complexity
	if complexErr := auth.ValidatePasswordComplexity(req.NewPassword); complexErr.HasErrors() {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: complexErr.Error()})
		return
	}

	// Hash and save new password
	salt := generateSalt()
	encrypted := auth.DigestPassword(req.NewPassword, salt, auth.DefaultStretches)
	h.db.Model(&user).Updates(map[string]interface{}{
		"encrypted_password": encrypted,
		"password_salt":      salt,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}
