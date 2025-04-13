// internal/handlers/media_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/middleware"
	"github.com/randilt/floe-cms/internal/models"
	"github.com/randilt/floe-cms/internal/storage"
	"github.com/randilt/floe-cms/internal/utils"
)

// MediaHandler handles media-related requests
type MediaHandler struct {
	db      *db.DB
	storage storage.Manager
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(db *db.DB, storage storage.Manager) *MediaHandler {
	return &MediaHandler{
		db:      db,
		storage: storage,
	}
}

// UploadMedia handles media uploads
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to parse form")
		return
	}

	// Get workspace ID
	workspaceIDStr := r.FormValue("workspace_id")
	if workspaceIDStr == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 32)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid workspace ID")
		return
	}

	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	// Get file
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "No file provided")
		return
	}
	defer file.Close()

	// Save file
	originalName, filePath, err := h.storage.Save(file, header, claims.UserID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save file: "+err.Error())
		return
	}

	// Create media record
	media := models.Media{
		WorkspaceID: uint(workspaceID),
		Name:        r.FormValue("name"),
		FileName:    originalName,
		FilePath:    filePath,
		MimeType:    header.Header.Get("Content-Type"),
		Size:        header.Size,
		UploadedBy:  claims.UserID,
	}

	// Use name from form or fallback to filename
	if media.Name == "" {
		media.Name = originalName
	}

	if err := h.db.Create(&media).Error; err != nil {
		// Delete the file if database save fails
		h.storage.Delete(filePath)
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create media record")
		return
	}

	// Add URL to response
	media.FilePath = h.storage.GetURL(filePath)

	utils.RespondWithSuccess(w, http.StatusCreated, media)
}

// GetMedia handles getting a single media item
func (h *MediaHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Media ID is required")
		return
	}

	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Media not found")
		return
	}

	// Add URL to response
	media.FilePath = h.storage.GetURL(media.FilePath)

	utils.RespondWithSuccess(w, http.StatusOK, media)
}

// ListMedia handles listing media items
func (h *MediaHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if workspaceID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

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

	var media []models.Media
	var total int64

	query := h.db.Model(&models.Media{}).Where("workspace_id = ?", workspaceID)

	if err := query.Count(&total).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to count media")
		return
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&media).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch media")
		return
	}

	// Add URLs to response
	for i := range media {
		media[i].FilePath = h.storage.GetURL(media[i].FilePath)
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"media":  media,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteMedia handles media deletion
func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Media ID is required")
		return
	}

	// Get media by ID
	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Media not found")
		return
	}

	// Get user from context
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
		return
	}

	// Check if user has permission to delete this media
	if claims.RoleName != "admin" && claims.UserID != media.UploadedBy {
		utils.RespondWithError(w, http.StatusForbidden, "Permission denied")
		return
	}

	// Delete file
	if err := h.storage.Delete(media.FilePath); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		return
	}

	// Delete media record
	if err := h.db.Delete(&media).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete media record")
		return
	}

	utils.RespondWithSuccess(w, http.StatusOK, map[string]string{"message": "Media deleted successfully"})
}