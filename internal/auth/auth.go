// internal/auth/auth.go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/randilt/floe-cms/internal/config"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/models"
)

// Claims represents JWT claims
type Claims struct {
	UserID      uint   `json:"user_id"`
	Email       string `json:"email"`
	RoleID      uint   `json:"role_id"`
	RoleName    string `json:"role_name"`
	WorkspaceID uint   `json:"workspace_id,omitempty"`
	jwt.RegisteredClaims
}

// Manager handles authentication operations
type Manager struct {
	db         *db.DB
	jwtSecret  []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

// NewManager creates a new authentication manager
func NewManager(db *db.DB, config config.AuthConfig) *Manager {
	return &Manager{
		db:         db,
		jwtSecret:  []byte(config.JWTSecret),
		accessExp:  time.Duration(config.AccessTokenExpiry) * time.Second,
		refreshExp: time.Duration(config.RefreshTokenExpiry) * time.Second,
	}
}

// Login attempts to log in a user with the given credentials
func (m *Manager) Login(email, password string) (string, string, error) {
	var user models.User
	if err := m.db.Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	if !user.Active {
		return "", "", errors.New("user account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	accessToken, err := m.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken refreshes an access token using a refresh token
func (m *Manager) RefreshToken(refreshTokenString string) (string, error) {
	var refreshToken models.RefreshToken
	if err := m.db.Preload("User.Role").Where("token = ? AND revoked = ? AND expires_at > ?", refreshTokenString, false, time.Now()).First(&refreshToken).Error; err != nil {
		return "", errors.New("invalid refresh token")
	}

	accessToken, err := m.generateAccessToken(refreshToken.User)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// RevokeToken revokes a refresh token
func (m *Manager) RevokeToken(refreshTokenString string) error {
	result := m.db.Model(&models.RefreshToken{}).Where("token = ?", refreshTokenString).Update("revoked", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}
	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// generateAccessToken generates a JWT access token for a user
func (m *Manager) generateAccessToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(m.accessExp)
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecret)
}

// generateRefreshToken generates a refresh token for a user
func (m *Manager) generateRefreshToken(userID uint) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	tokenString := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(m.refreshExp)

	refreshToken := models.RefreshToken{
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: expiresAt,
		Revoked:   false,
	}

	if err := m.db.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	return tokenString, nil
}

// EnsureAdminExists ensures that an admin user exists in the system
func EnsureAdminExists(db *db.DB, email, password string) error {
	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create admin role if it doesn't exist
			adminRole = models.Role{
				Name:        "admin",
				Description: "Administrator with full access",
			}
			if err := db.Create(&adminRole).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create editor and viewer roles if they don't exist
	var editorRole models.Role
	if err := db.Where("name = ?", "editor").First(&editorRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			editorRole = models.Role{
				Name:        "editor",
				Description: "Editor with content management access",
			}
			if err := db.Create(&editorRole).Error; err != nil {
				return err
			}
		}
	}

	var viewerRole models.Role
	if err := db.Where("name = ?", "viewer").First(&viewerRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			viewerRole = models.Role{
				Name:        "viewer",
				Description: "Viewer with read-only access",
			}
			if err := db.Create(&viewerRole).Error; err != nil {
				return err
			}
		}
	}

	// Check if admin user exists
	var count int64
	if err := db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		// Create admin user if it doesn't exist
		hashedPassword, err := HashPassword(password)
		if err != nil {
			return err
		}

		adminUser := models.User{
			Email:        email,
			PasswordHash: hashedPassword,
			FirstName:    "Admin",
			LastName:     "User",
			RoleID:       adminRole.ID,
			Active:       true,
		}

		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}

		// Create default workspace
		defaultWorkspace := models.Workspace{
			Name:        "Default",
			Slug:        "default",
			Description: "Default workspace",
		}

		if err := db.Create(&defaultWorkspace).Error; err != nil {
			return err
		}

		// Associate admin with default workspace
		userWorkspace := models.UserWorkspace{
			UserID:      adminUser.ID,
			WorkspaceID: defaultWorkspace.ID,
		}

		if err := db.Create(&userWorkspace).Error; err != nil {
			return err
		}
	}

	return nil
}

// ResetAdminUser resets the admin user's password
func ResetAdminUser(db *db.DB, email, password string) error {
	// Hash the new password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	// Update the admin user
	result := db.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"password_hash": hashedPassword,
		"active":        true,
	})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("admin user not found")
	}

	return nil
}