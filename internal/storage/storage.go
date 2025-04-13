// internal/storage/storage.go
package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/randilt/floe-cms/internal/utils"
)

// Manager defines the interface for storage operations
type Manager interface {
	Save(file multipart.File, header *multipart.FileHeader, userID uint) (string, string, error)
	Delete(path string) error
	GetURL(path string) string
}

// LocalStorage implements storage operations on local filesystem
type LocalStorage struct {
	uploadsDir string
}

// NewLocalStorage creates a new local storage manager
func NewLocalStorage(uploadsDir string) *LocalStorage {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create uploads directory: %v", err))
	}
	return &LocalStorage{
		uploadsDir: uploadsDir,
	}
}

// Save saves a file to the local filesystem
func (ls *LocalStorage) Save(file multipart.File, header *multipart.FileHeader, userID uint) (string, string, error) {
	// Generate a unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", userID, utils.GenerateRandomString(16), ext)

	// Create a subdirectory based on current date
	now := time.Now()
	subdir := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	fullDir := filepath.Join(ls.uploadsDir, subdir)
	
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the full file path
	fullPath := filepath.Join(fullDir, filename)
	
	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy the file data
	if _, err := io.Copy(dst, file); err != nil {
		return "", "", fmt.Errorf("failed to copy file: %v", err)
	}

	// Return the relative path for storage in the database
	relativePath := filepath.Join(subdir, filename)
	return header.Filename, relativePath, nil
}

// Delete deletes a file from the local filesystem
func (ls *LocalStorage) Delete(path string) error {
	if path == "" {
		return errors.New("empty file path")
	}

	fullPath := filepath.Join(ls.uploadsDir, path)
	
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return errors.New("file not found")
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// GetURL returns the URL for a file
func (ls *LocalStorage) GetURL(path string) string {
	return "/uploads/" + path
}