// internal/middleware/middleware.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/utils"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user context
	UserContextKey ContextKey = "user"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware handles authentication
func AuthMiddleware(authManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			// Check if the header has the Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			claims, err := authManager.ValidateToken(tokenString)
			if err != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Add user claims to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnly ensures only admins can access the route
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
			return
		}

		if claims.RoleName != "admin" {
			utils.RespondWithError(w, http.StatusForbidden, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EditorOrAbove ensures only editors or admins can access the route
func EditorOrAbove(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
			return
		}

		if claims.RoleName != "admin" && claims.RoleName != "editor" {
			utils.RespondWithError(w, http.StatusForbidden, "Editor or admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireWorkspace ensures the user has access to the specified workspace
func RequireWorkspace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get user from context")
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Workspace ID required")
			return
		}

		// TODO: Check if user has access to workspace
		// This would require a database lookup to check the user's workspace associations
		// For now, we'll allow admins to access any workspace
		if claims.RoleName != "admin" {
			// Check if the workspace ID in the request matches the one in the claims
			// This is a simplified check and should be replaced with a proper database lookup
			if claims.WorkspaceID != 0 && claims.WorkspaceID != uint(utils.ParseUint(workspaceID)) {
				utils.RespondWithError(w, http.StatusForbidden, "Access denied to this workspace")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}