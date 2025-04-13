// internal/handlers/content_handler.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/middleware"
	"github.com/randilt/floe-cms/internal/models"
	"github.com/randilt/floe-cms/internal/storage"
	"github.com/randilt/floe-cms/internal/utils"
)

// ContentHandler handles content-related requests
type ContentHandler struct {
	db      *db.DB
	storage storage.Manager
}

// NewContentHandler creates a new content handler
func NewContentHandler(db *db.DB, storage storage.Manager) *ContentHandler {
	return &ContentHandler{
		db:      db,
		storage: storage,
	}
}

// CreateContentRequest represents a request to create content
type CreateContentRequest struct {
	WorkspaceID   uint   `json:"workspace_id"`
	ContentTypeID uint   `json:"content_type_id"`
	Title         string `json:"title"`
	Slug          string `json:"slug"`
	Body          string `json:"body"`
	Status        string `json:"status"`
	MetaData      string `json:"meta_data"`
}

// CreateContent handles content creation
func (h *ContentHandler) CreateContent(w http.ResponseWriter, r *http.Request) {
	var req CreateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate fields
	if req.WorkspaceID == 0 || req.Title == "" || req.Body == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID, title, and body are required")
		return
	}

	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = utils.ToSlug(req.Title)
	}

	// Create content
	content := models.Content{
		WorkspaceID:   req.WorkspaceID,
		ContentTypeID: req.ContentTypeID,
		Title:         req.Title,
		Slug:          req.Slug,
		Body:          req.Body,
		Status:        req.Status,
		AuthorID:      claims.UserID,
		MetaData:      req.MetaData,
	}

	// Set publish date if status is published
	if req.Status == "published" {
		now := time.Now()
		content.PublishedAt = &now
	}

	if err := h.db.Create(&content).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create content")
		return
	}

	utils.RespondWithSuccess(w, http.StatusCreated, content)
}

// UpdateContentRequest represents a request to update content
type UpdateContentRequest struct {
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Body     string `json:"body"`
	Status   string `json:"status"`
	MetaData string `json:"meta_data"`
}

// UpdateContent handles content updates
func (h *ContentHandler) UpdateContent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Content ID is required")
		return
	}

	var req UpdateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get content by ID
	var content models.Content
	if err := h.db.First(&content, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Content not found")
		return
	}

	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	// Check if user has permission to update this content
	if claims.RoleName != "admin" && claims.UserID != content.AuthorID {
		utils.RespondWithError(w, http.StatusForbidden, "Permission denied")
		return
	}

	// Update fields
	if req.Title != "" {
		content.Title = req.Title
	}
	if req.Slug != "" {
		content.Slug = req.Slug
	}
	if req.Body != "" {
		content.Body = req.Body
	}
	if req.Status != "" {
		// Update publish date if status changed to published
		if content.Status != "published" && req.Status == "published" {
			now := time.Now()
			content.PublishedAt = &now
		}
		content.Status = req.Status
	}
	if req.MetaData != "" {
		content.MetaData = req.MetaData
	}

	if err := h.db.Save(&content).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update content")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, content)
}

// GetContent handles getting a single content item
func (h *ContentHandler) GetContent(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Content ID is required")
        return
    }

    var content models.Content
    if err := h.db.Preload("Author").Preload("ContentType").Preload("Workspace").First(&content, id).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Content not found")
        return
    }

    // Get user from context
    claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
    if !ok {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
        return
    }

    // Check if user has access to this content's workspace
    if claims.RoleName != "admin" {
        // For non-admin users, check if they have access to this workspace
        var count int64
        if err := h.db.Model(&models.UserWorkspace{}).Where("user_id = ? AND workspace_id = ?", claims.UserID, content.WorkspaceID).Count(&count).Error; err != nil {
            utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check workspace access")
            return
        }

        if count == 0 {
            utils.RespondWithError(w, http.StatusForbidden, "You don't have access to this content")
            return
        }
    }

    utils.RespondWithSuccess(w, http.StatusOK, content)
}

// ListContent handles listing content items
func (h *ContentHandler) ListContent(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	status := r.URL.Query().Get("status")
	contentTypeID := r.URL.Query().Get("content_type_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

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

    query := h.db.Model(&models.Content{}).Preload("Author").Preload("ContentType")

    if workspaceID != "" {
        query = query.Where("workspace_id = ?", workspaceID)
    }

    if status != "" {
        query = query.Where("status = ?", status)
    }

    if contentTypeID != "" {
        query = query.Where("content_type_id = ?", contentTypeID)
    }

    var contents []models.Content
    var total int64

    if err := query.Count(&total).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to count contents")
        return
    }

    if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&contents).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch contents")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
        "contents": contents,
        "total":    total,
        "limit":    limit,
        "offset":   offset,
    })
}

// DeleteContent handles content deletion
func (h *ContentHandler) DeleteContent(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Content ID is required")
        return
    }

    // Get content by ID
    var content models.Content
    if err := h.db.First(&content, id).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Content not found")
        return
    }

    // Get user from context
    claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
    if !ok {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
        return
    }

    // Check if user has permission to delete this content
    if claims.RoleName != "admin" && claims.UserID != content.AuthorID {
        utils.RespondWithError(w, http.StatusForbidden, "Permission denied")
        return
    }

    // Delete content
    if err := h.db.Delete(&content).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete content")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Content deleted successfully"})
}

// GetContentBySlug handles getting content by slug
func (h *ContentHandler) GetContentBySlug(w http.ResponseWriter, r *http.Request) {
    workspace := chi.URLParam(r, "workspace")
    slug := chi.URLParam(r, "slug")

    if workspace == "" || slug == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Workspace and slug are required")
        return
    }

    var workspaceObj models.Workspace
    if err := h.db.Where("slug = ?", workspace).First(&workspaceObj).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Workspace not found")
        return
    }

    var content models.Content
    if err := h.db.Where("workspace_id = ? AND slug = ? AND status = ?", workspaceObj.ID, slug, "published").
        Preload("Author").
        Preload("ContentType").
        First(&content).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Content not found")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, content)
}

// GetPublishedContent handles getting published content for a workspace
func (h *ContentHandler) GetPublishedContent(w http.ResponseWriter, r *http.Request) {
    workspace := chi.URLParam(r, "workspace")
    if workspace == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Workspace is required")
        return
    }

    var workspaceObj models.Workspace
    if err := h.db.Where("slug = ?", workspace).First(&workspaceObj).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Workspace not found")
        return
    }

    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    contentTypeID := r.URL.Query().Get("content_type_id")

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

    query := h.db.Model(&models.Content{}).
        Where("workspace_id = ? AND status = ?", workspaceObj.ID, "published").
        Preload("Author").
        Preload("ContentType")

    if contentTypeID != "" {
        query = query.Where("content_type_id = ?", contentTypeID)
    }

    var contents []models.Content
    var total int64

    if err := query.Count(&total).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to count contents")
        return
    }

    if err := query.Limit(limit).Offset(offset).Order("published_at desc").Find(&contents).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch contents")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
        "contents": contents,
        "total":    total,
        "limit":    limit,
        "offset":   offset,
    })
}

// CreateContentTypeRequest represents a request to create a content type
type CreateContentTypeRequest struct {
    WorkspaceID uint                 `json:"workspace_id"`
    Name        string               `json:"name"`
    Slug        string               `json:"slug"`
    Description string               `json:"description"`
    Fields      []models.ContentField `json:"fields"`
}

// CreateContentType handles content type creation
func (h *ContentHandler) CreateContentType(w http.ResponseWriter, r *http.Request) {
    var req CreateContentTypeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // Validate fields
    if req.WorkspaceID == 0 || req.Name == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID and name are required")
        return
    }

    // Generate slug if not provided
    if req.Slug == "" {
        req.Slug = utils.ToSlug(req.Name)
    }

    // Create content type
    contentType := models.ContentType{
        WorkspaceID: req.WorkspaceID,
        Name:        req.Name,
        Slug:        req.Slug,
        Description: req.Description,
        Fields:      req.Fields,
    }

    if err := h.db.Create(&contentType).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create content type")
        return
    }

    utils.RespondWithSuccess(w, http.StatusCreated, contentType)
}

// UpdateContentTypeRequest represents a request to update a content type
type UpdateContentTypeRequest struct {
    Name        string               `json:"name"`
    Slug        string               `json:"slug"`
    Description string               `json:"description"`
    Fields      []models.ContentField `json:"fields"`
}

// UpdateContentType handles content type updates
func (h *ContentHandler) UpdateContentType(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Content type ID is required")
        return
    }

    var req UpdateContentTypeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    // Get content type by ID
    var contentType models.ContentType
    if err := h.db.First(&contentType, id).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Content type not found")
        return
    }

    // Update fields
    if req.Name != "" {
        contentType.Name = req.Name
    }
    if req.Slug != "" {
        contentType.Slug = req.Slug
    }
    if req.Description != "" {
        contentType.Description = req.Description
    }
    if req.Fields != nil {
        contentType.Fields = req.Fields
    }

    if err := h.db.Save(&contentType).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update content type")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, contentType)
}

// GetContentType handles getting a single content type
func (h *ContentHandler) GetContentType(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Content type ID is required")
        return
    }

    var contentType models.ContentType
    if err := h.db.First(&contentType, id).Error; err != nil {
        utils.RespondWithError(w, http.StatusNotFound, "Content type not found")
        return
    }

    utils.RespondWithSuccess(w, http.StatusOK, contentType)
}

// ListContentTypes handles listing content types
func (h *ContentHandler) ListContentTypes(w http.ResponseWriter, r *http.Request) {
    workspaceID := r.URL.Query().Get("workspace_id")
    
    if workspaceID == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
        return
    }
    
    var contentTypes []models.ContentType
    if err := h.db.Where("workspace_id = ?", workspaceID).Find(&contentTypes).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch content types")
        return
    }
    
    utils.RespondWithSuccess(w, http.StatusOK, contentTypes)
}

// DeleteContentType handles content type deletion
func (h *ContentHandler) DeleteContentType(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        utils.RespondWithError(w, http.StatusBadRequest, "Content type ID is required")
        return
    }
    
    // Check if content type is in use
    var count int64
    if err := h.db.Model(&models.Content{}).Where("content_type_id = ?", id).Count(&count).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check content type usage")
        return
    }
    
    if count > 0 {
        utils.RespondWithError(w, http.StatusBadRequest, "Cannot delete content type that is in use")
        return
    }
    
    // Delete content type
    if err := h.db.Delete(&models.ContentType{}, id).Error; err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete content type")
        return
    }
    
    utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Content type deleted successfully"})
}