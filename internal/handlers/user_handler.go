// internal/handlers/user_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/middleware"
	"github.com/randilt/floe-cms/internal/models"
	"github.com/randilt/floe-cms/internal/utils"
)

// UserHandler handles user-related requests
type UserHandler struct {
	db *db.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *db.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    uint   `json:"role_id"`
}

// CreateUser handles user creation
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.Email == "" || req.Password == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
		return
	}

	if err := utils.ValidatePassword(req.Password, 8); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if email is already taken
	var count int64
	if err := h.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check email availability")
		return
	}

	if count > 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Email is already taken")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user
	user := models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleID:       req.RoleID,
		Active:       true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	utils.RespondWithSuccess(w, http.StatusCreated, user)
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    uint   `json:"role_id"`
	Active    *bool  `json:"active"`
}

// UpdateUser handles user updates
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get user by ID
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// If email is being changed, check if it's already taken
	if req.Email != "" && req.Email != user.Email {
		if !utils.ValidateEmail(req.Email) {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		}

		var count int64
		if err := h.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check email availability")
			return
		}

		if count > 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is already taken")
			return
		}

		user.Email = req.Email
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.RoleID != 0 {
		user.RoleID = req.RoleID
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	utils.RespondWithSuccess(w, http.StatusOK, user)
}

// GetUser handles getting a single user
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	var user models.User
	if err := h.db.Preload("Role").First(&user, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	utils.RespondWithSuccess(w, http.StatusOK, user)
}

// ListUsers handles listing users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	roleID := r.URL.Query().Get("role_id")

	limit := 10
	offset := 0

	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := h.db.Model(&models.User{}).Preload("Role")

	if roleID != "" {
		query = query.Where("role_id = ?", roleID)
	}

	var users []models.User
	var total int64

	if err := query.Count(&total).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to count users")
		return
	}

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Don't return password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteUser handles user deletion
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if user is the last admin
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if user.RoleID == 1 { // Assuming admin role ID is 1
		var adminCount int64
		if err := h.db.Model(&models.User{}).Where("role_id = 1").Count(&adminCount).Error; err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check admin count")
			return
		}

		if adminCount <= 1 {
			utils.RespondWithError(w, http.StatusBadRequest, "Cannot delete the only admin user")
			return
		}
	}

	// Delete user
	if err := h.db.Delete(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// GetCurrentUser handles getting the current user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	var user models.User
	if err := h.db.Preload("Role").First(&user, claims.UserID).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	// Get user's workspaces
	var userWorkspaces []models.UserWorkspace
	if err := h.db.Where("user_id = ?", user.ID).Preload("Workspace").Find(&userWorkspaces).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user workspaces")
		return
	}

	var workspaces []models.Workspace
	for _, uw := range userWorkspaces {
		workspaces = append(workspaces, uw.Workspace)
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"user":       user,
		"workspaces": workspaces,
	})
}

// UpdateCurrentUserRequest represents a request to update the current user
type UpdateCurrentUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// UpdateCurrentUser handles updating the current user
func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	var req UpdateCurrentUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get user by ID
	var user models.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// If email is being changed, check if it's already taken
	if req.Email != "" && req.Email != user.Email {
		if !utils.ValidateEmail(req.Email) {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid email format")
			return
		}

		var count int64
		if err := h.db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check email availability")
			return
		}

		if count > 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is already taken")
			return
		}

		user.Email = req.Email
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if err := h.db.Save(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Don't return password hash
	user.PasswordHash = ""

	utils.RespondWithSuccess(w, http.StatusOK, user)
}

// ChangePasswordRequest represents a request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword handles changing the current user's password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.OldPassword == "" || req.NewPassword == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Old password and new password are required")
		return
	}

	if err := utils.ValidatePassword(req.NewPassword, 8); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get user by ID
	var user models.User
	if err := h.db.First(&user, claims.UserID).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid old password")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Update password
	user.PasswordHash = hashedPassword
	if err := h.db.Save(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}