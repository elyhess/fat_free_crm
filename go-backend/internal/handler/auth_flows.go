package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/model"
	"github.com/elyhess/fat-free-crm-backend/internal/service"
)

const resetTokenExpiry = 6 * time.Hour

// AuthFlowsHandler handles password reset, registration, and email confirmation.
type AuthFlowsHandler struct {
	db       *gorm.DB
	emailSvc *service.EmailService
	baseURL  string // frontend base URL for links in emails
}

func NewAuthFlowsHandler(db *gorm.DB, emailSvc *service.EmailService, baseURL string) *AuthFlowsHandler {
	return &AuthFlowsHandler{db: db, emailSvc: emailSvc, baseURL: baseURL}
}

// --- Password Reset ---

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPassword generates a reset token and sends an email.
// POST /auth/forgot-password
func (h *AuthFlowsHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Always return success to prevent email enumeration
	defer writeJSON(w, http.StatusOK, map[string]string{
		"status": "If that email exists, a reset link has been sent.",
	})

	if req.Email == "" {
		return
	}

	var user model.User
	if err := h.db.Where("email = ? AND deleted_at IS NULL", req.Email).First(&user).Error; err != nil {
		return
	}

	// Generate token
	rawToken := generateSecureToken()
	hashedToken := hashToken(rawToken)
	now := time.Now().UTC()

	h.db.Model(&user).Updates(map[string]interface{}{
		"reset_password_token":   hashedToken,
		"reset_password_sent_at": now,
	})

	// Send email with raw token (user receives unhashed version)
	resetURL := h.baseURL + "/reset-password?token=" + rawToken
	if err := h.emailSvc.SendPasswordReset(user.Email, resetURL); err != nil {
		slog.Error("failed to send password reset email", "error", err, "user_id", user.ID)
	}
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ResetPassword validates the token and sets a new password.
// POST /auth/reset-password
func (h *AuthFlowsHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "token and new_password are required")
		return
	}

	// Hash the provided token to match DB
	hashedToken := hashToken(req.Token)

	var user model.User
	if err := h.db.Where("reset_password_token = ? AND deleted_at IS NULL", hashedToken).First(&user).Error; err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid or expired reset token")
		return
	}

	// Check expiry
	if user.ResetPasswordSentAt == nil || time.Since(*user.ResetPasswordSentAt) > resetTokenExpiry {
		writeError(w, http.StatusUnprocessableEntity, "invalid or expired reset token")
		return
	}

	// Validate password complexity
	if complexErr := auth.ValidatePasswordComplexity(req.NewPassword); complexErr.HasErrors() {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: complexErr.Error()})
		return
	}

	// Set new password
	salt := generateSalt()
	encrypted := auth.DigestPassword(req.NewPassword, salt, auth.DefaultStretches)

	h.db.Model(&user).Updates(map[string]interface{}{
		"encrypted_password":     encrypted,
		"password_salt":          salt,
		"reset_password_token":   nil,
		"reset_password_sent_at": nil,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "password has been reset"})
}

// --- Registration ---

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates a new user account.
// POST /auth/register
func (h *AuthFlowsHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Check if signup is allowed
	signupSetting := h.getSignupSetting()
	if signupSetting == "not_allowed" {
		writeError(w, http.StatusForbidden, "user registration is not allowed")
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}

	// Validate password complexity
	if complexErr := auth.ValidatePasswordComplexity(req.Password); complexErr.HasErrors() {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: complexErr.Error()})
		return
	}

	// Check uniqueness
	var count int64
	h.db.Model(&model.User{}).Where("(username = ? OR email = ?) AND deleted_at IS NULL", req.Username, req.Email).Count(&count)
	if count > 0 {
		writeError(w, http.StatusConflict, "username or email already taken")
		return
	}

	// Create user
	salt := generateSalt()
	encrypted := auth.DigestPassword(req.Password, salt, auth.DefaultStretches)
	now := time.Now().UTC()

	confirmToken := generateSecureToken()
	hashedConfirmToken := hashToken(confirmToken)

	user := model.User{
		Username:          req.Username,
		Email:             req.Email,
		EncryptedPassword: encrypted,
		PasswordSalt:      salt,
		ConfirmationToken: &hashedConfirmToken,
		ConfirmationSentAt: &now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// If needs approval, suspend the user
	if signupSetting == "needs_approval" {
		user.SuspendedAt = &now
	}

	if err := h.db.Create(&user).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	// Send confirmation email
	confirmURL := h.baseURL + "/confirm?token=" + confirmToken
	if err := h.emailSvc.SendConfirmation(req.Email, confirmURL); err != nil {
		slog.Error("failed to send confirmation email", "error", err, "user_id", user.ID)
	}

	// Send welcome email
	needsApproval := signupSetting == "needs_approval"
	if err := h.emailSvc.SendWelcome(req.Email, req.Username, needsApproval); err != nil {
		slog.Error("failed to send welcome email", "error", err, "user_id", user.ID)
	}

	status := "account created — please check your email to confirm"
	if needsApproval {
		status = "account created — pending admin approval"
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": status})
}

// --- Email Confirmation ---

type confirmRequest struct {
	Token string `json:"token"`
}

// ConfirmEmail validates the confirmation token and confirms the user.
// POST /auth/confirm
func (h *AuthFlowsHandler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req confirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	hashedToken := hashToken(req.Token)

	var user model.User
	if err := h.db.Where("confirmation_token = ? AND deleted_at IS NULL", hashedToken).First(&user).Error; err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid confirmation token")
		return
	}

	if user.ConfirmedAt != nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "email already confirmed"})
		return
	}

	now := time.Now().UTC()
	h.db.Model(&user).Updates(map[string]interface{}{
		"confirmed_at":       now,
		"confirmation_token": nil,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "email confirmed"})
}

type resendConfirmationRequest struct {
	Email string `json:"email"`
}

// ResendConfirmation resends the confirmation email.
// POST /auth/resend-confirmation
func (h *AuthFlowsHandler) ResendConfirmation(w http.ResponseWriter, r *http.Request) {
	var req resendConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Always return success to prevent email enumeration
	defer writeJSON(w, http.StatusOK, map[string]string{
		"status": "If that email exists and is unconfirmed, a confirmation link has been sent.",
	})

	if req.Email == "" {
		return
	}

	var user model.User
	if err := h.db.Where("email = ? AND confirmed_at IS NULL AND deleted_at IS NULL", req.Email).First(&user).Error; err != nil {
		return
	}

	confirmToken := generateSecureToken()
	hashedToken := hashToken(confirmToken)
	now := time.Now().UTC()

	h.db.Model(&user).Updates(map[string]interface{}{
		"confirmation_token":  hashedToken,
		"confirmation_sent_at": now,
	})

	confirmURL := h.baseURL + "/confirm?token=" + confirmToken
	if err := h.emailSvc.SendConfirmation(user.Email, confirmURL); err != nil {
		slog.Error("failed to send confirmation email", "error", err, "user_id", user.ID)
	}
}

// --- Helpers ---

func (h *AuthFlowsHandler) getSignupSetting() string {
	var setting struct {
		Value string
	}
	// Settings are stored as YAML in the `value` column, keyed by name
	if err := h.db.Table("settings").Where("name = ?", "user_signup").Select("value").Scan(&setting).Error; err != nil {
		return "not_allowed" // default: no signup if setting missing
	}
	v := setting.Value
	// The Rails setting stores YAML like "--- :allowed\n" — extract the symbol
	if strings.Contains(v, "allowed") && !strings.Contains(v, "not_allowed") {
		return "allowed"
	}
	if strings.Contains(v, "needs_approval") {
		return "needs_approval"
	}
	return "not_allowed"
}

// generateSecureToken returns a 32-byte hex-encoded random token.
func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// hashToken returns the SHA-256 hex digest of a token.
// We store hashed tokens in the DB so a DB leak doesn't expose reset links.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
