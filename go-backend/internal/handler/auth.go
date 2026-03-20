package handler

import (
	"encoding/json"
	"net/http"

	"github.com/elyhess/fat-free-crm-backend/internal/auth"
	"github.com/elyhess/fat-free-crm-backend/internal/repository"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	jwt      *auth.JWTService
}

func NewAuthHandler(userRepo *repository.UserRepository, jwt *auth.JWTService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwt: jwt}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string    `json:"token"`
	User  loginUser `json:"user"`
}

type loginUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Admin     bool   `json:"admin"`
}

// Login authenticates a user and returns a JWT.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.Login == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "login and password are required"})
		return
	}

	user, err := h.userRepo.FindByLogin(req.Login)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	if !auth.VerifyPassword(req.Password, user.EncryptedPassword, user.PasswordSalt, auth.DefaultStretches) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	if !user.IsActive() {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "account is not active"})
		return
	}

	token, err := h.jwt.GenerateToken(user.ID, user.Username, user.Admin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to generate token"})
		return
	}

	// Update sign-in tracking
	_ = h.userRepo.UpdateSignInTracking(user, r.RemoteAddr)

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User: loginUser{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Admin:     user.Admin,
		},
	})
}
