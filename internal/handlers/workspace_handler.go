// internal/handlers/workspace_handler.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/models"
	"github.com/randilt/floe-cms/internal/utils"
)

// WorkspaceHandler handles workspace-related requests
type WorkspaceHandler struct {
	db *db.DB
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(db *db.DB) *WorkspaceHandler {
	return &WorkspaceHandler{
		db: db,
	}
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// CreateWorkspace handles workspace creation
func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = utils.ToSlug(req.Name)
	}

	// Create workspace
	workspace := models.Workspace{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if err := h.db.Create(&workspace).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create workspace")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, workspace)
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

// UpdateWorkspace handles workspace updates
func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	var req UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get workspace by ID
	var workspace models.Workspace
	if err := h.db.First(&workspace, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	// Update fields
	if req.Name != "" {
		workspace.Name = req.Name
	}
	if req.Slug != "" {
		workspace.Slug = req.Slug
	}
	if req.Description != "" {
		workspace.Description = req.Description
	}

	if err := h.db.Save(&workspace).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update workspace")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, workspace)
}

// GetWorkspace handles getting a single workspace
func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	var workspace models.Workspace
	if err := h.db.First(&workspace, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, workspace)
}

// ListWorkspaces handles listing workspaces
func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	var workspaces []models.Workspace
	if err := h.db.Find(&workspaces).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch workspaces")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, workspaces)
}

// DeleteWorkspace handles workspace deletion
func (h *WorkspaceHandler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Check if workspace is in use
	var contentCount int64
	if err := h.db.Model(&models.Content{}).Where("workspace_id = ?", id).Count(&contentCount).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check workspace usage")
		return
	}

	if contentCount > 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "Cannot delete workspace that has content")
		return
	}

	// Delete workspace
	if err := h.db.Delete(&models.Workspace{}, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete workspace")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Workspace deleted successfully"})
}

// AddUserToWorkspaceRequest represents a request to add a user to a workspace
type AddUserToWorkspaceRequest struct {
	UserID uint `json:"user_id"`
}

// AddUserToWorkspace handles adding a user to a workspace
func (h *WorkspaceHandler) AddUserToWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	var req AddUserToWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if workspace exists
	var workspace models.Workspace
	if err := h.db.First(&workspace, workspaceID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Workspace not found")
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Check if user is already in workspace
	var existingCount int64
	if err := h.db.Model(&models.UserWorkspace{}).
		Where("user_id = ? AND workspace_id = ?", req.UserID, workspaceID).
		Count(&existingCount).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check user workspace association")
		return
	}

	if existingCount > 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "User is already in this workspace")
		return
	}

	// Add user to workspace
	userWorkspace := models.UserWorkspace{
		UserID:      req.UserID,
		WorkspaceID: utils.ParseUint(workspaceID),
	}

	if err := h.db.Create(&userWorkspace).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to add user to workspace")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, map[string]string{"message": "User added to workspace successfully"})
}

// RemoveUserFromWorkspace handles removing a user from a workspace
func (h *WorkspaceHandler) RemoveUserFromWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")

	if workspaceID == "" || userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID and User ID are required")
		return
	}

	// Delete user-workspace association
	result := h.db.Where("user_id = ? AND workspace_id = ?", userID, workspaceID).Delete(&models.UserWorkspace{})

	if result.Error != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to remove user from workspace")
		return
	}

	if result.RowsAffected == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "User is not in this workspace")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "User removed from workspace successfully"})
}