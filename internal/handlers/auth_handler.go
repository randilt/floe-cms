// internal/handlers/auth_handler.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authManager *auth.Manager
	db          *db.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authManager *auth.Manager, db *db.DB) *AuthHandler {
	return &AuthHandler{
		authManager: authManager,
		db:          db,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.Email == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Attempt login
	accessToken, refreshToken, err := h.authManager.Login(req.Email, req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Return tokens
	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// RefreshToken handles refresh token requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.RefreshToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Refresh the token
	accessToken, err := h.authManager.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Return the new access token
	response := RefreshTokenResponse{
		AccessToken: accessToken,
	}

	utils.RespondWithSuccess(w, http.StatusOK, response)
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.RefreshToken == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	// Revoke the refresh token
	if err := h.authManager.RevokeToken(req.RefreshToken); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}