// internal/models/models.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents a user in the system
type User struct {
	BaseModel
	Email          string          `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string          `gorm:"not null" json:"-"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	RoleID         uint            `json:"role_id"`
	Role           Role            `json:"role"`
	Active         bool            `gorm:"default:true" json:"active"`
	RefreshTokens  []RefreshToken  `json:"-"`
	UserWorkspaces []UserWorkspace `json:"-"`
}

// Role represents a user role in the system
type Role struct {
	BaseModel
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission represents a permission in the system
type Permission struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	Description string `json:"description"`
}

// Workspace represents a workspace/tenant in the system
type Workspace struct {
	BaseModel
	Name          string          `gorm:"not null" json:"name"`
	Slug          string          `gorm:"uniqueIndex;not null" json:"slug"`
	Description   string          `json:"description"`
	UserWorkspaces []UserWorkspace `json:"-"`
	Contents      []Content       `json:"-"`
	Media         []Media         `json:"-"`
	ContentTypes  []ContentType   `json:"-"`
}

// UserWorkspace represents the relationship between users and workspaces
type UserWorkspace struct {
	BaseModel
	UserID      uint      `gorm:"index:idx_user_workspace,unique" json:"user_id"`
	WorkspaceID uint      `gorm:"index:idx_user_workspace,unique" json:"workspace_id"`
	User        User      `json:"-"`
	Workspace   Workspace `json:"-"`
}

// ContentType represents a type of content in the system
type ContentType struct {
	BaseModel
	WorkspaceID  uint            `json:"workspace_id"`
	Workspace    Workspace       `json:"-"`
	Name         string          `gorm:"not null" json:"name"`
	Slug         string          `gorm:"not null" json:"slug"`
	Description  string          `json:"description"`
	Fields       []ContentField  `gorm:"serializer:json" json:"fields"`
	Contents     []Content       `json:"-"`
}

// ContentField represents a field definition for a content type
type ContentField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Content represents content in the system
type Content struct {
	BaseModel
	WorkspaceID   uint        `json:"workspace_id"`
	Workspace     Workspace   `json:"-"`
	ContentTypeID uint        `json:"content_type_id"`
	ContentType   ContentType `json:"content_type"`
	Title         string      `gorm:"not null" json:"title"`
	Slug          string      `gorm:"not null" json:"slug"`
	Body          string      `gorm:"type:text" json:"body"`
	Status        string      `gorm:"default:'draft'" json:"status"`
	AuthorID      uint        `json:"author_id"`
	Author        User        `json:"author"`
	PublishedAt   *time.Time  `json:"published_at"`
	MetaData      string      `gorm:"type:text" json:"meta_data"`
}

// Media represents media files in the system
// Media represents media files in the system
type Media struct {
    BaseModel
    WorkspaceID uint      `json:"workspace_id"`
    Workspace   Workspace `json:"-"`
    Name        string    `gorm:"not null" json:"name"`
    FileName    string    `gorm:"not null" json:"file_name"`
    FilePath    string    `gorm:"not null" json:"file_path"`
    MimeType    string    `json:"mime_type"`
    Size        int64     `json:"size"`
    UploadedBy  uint      `json:"uploaded_by"`
    User        User      `gorm:"foreignKey:UploadedBy" json:"user"` // Add foreignKey tag here
}

// RefreshToken represents a refresh token for a user
type RefreshToken struct {
	BaseModel
	UserID    uint      `gorm:"index" json:"user_id"`
	User      User      `json:"-"`
	Token     string    `gorm:"uniqueIndex" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
}