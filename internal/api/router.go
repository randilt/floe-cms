// internal/api/router.go
package api

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/config"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/handlers"
	mw "github.com/randilt/floe-cms/internal/middleware"
	"github.com/randilt/floe-cms/internal/storage"
)

// NewRouter creates a new router for the API
func NewRouter(authManager *auth.Manager, db *db.DB, storage storage.Manager, adminUI embed.FS, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting middleware
	r.Use(httprate.LimitByIP(
		cfg.Auth.RateLimitRequests,
		time.Duration(cfg.Auth.RateLimitExpiry)*time.Second,
	))

	// Security headers middleware
	r.Use(mw.SecurityHeaders)

	// Create handlers
	authHandler := handlers.NewAuthHandler(authManager, db)
	contentHandler := handlers.NewContentHandler(db, storage)
	mediaHandler := handlers.NewMediaHandler(db, storage)
	workspaceHandler := handlers.NewWorkspaceHandler(db)
	userHandler := handlers.NewUserHandler(db)

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Floe CMS is running"))
	})

	// Authentication routes
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/refresh", authHandler.RefreshToken)

	// Public content routes
	r.Get("/api/content/{workspace}", contentHandler.GetPublishedContent)
	r.Get("/api/content/{workspace}/{slug}", contentHandler.GetContentBySlug)

	// Serve uploads
	fileServer := http.FileServer(http.Dir(cfg.Storage.UploadsDir))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware(authManager))

		// Auth routes
		r.Post("/api/auth/logout", authHandler.Logout)

		// Content routes
		r.Route("/api/content", func(r chi.Router) {
			r.Post("/", contentHandler.CreateContent)
			r.Get("/", contentHandler.ListContent)
			r.Get("/{id}", contentHandler.GetContent)
			r.Put("/{id}", contentHandler.UpdateContent)
			r.Delete("/{id}", contentHandler.DeleteContent)
		})

		// Content type routes
		r.Route("/api/content-types", func(r chi.Router) {
			r.Post("/", contentHandler.CreateContentType)
			r.Get("/", contentHandler.ListContentTypes)
			r.Get("/{id}", contentHandler.GetContentType)
			r.Put("/{id}", contentHandler.UpdateContentType)
			r.Delete("/{id}", contentHandler.DeleteContentType)
		})

		// Media routes
		r.Route("/api/media", func(r chi.Router) {
			r.Post("/", mediaHandler.UploadMedia)
			r.Get("/", mediaHandler.ListMedia)
			r.Get("/{id}", mediaHandler.GetMedia)
			r.Delete("/{id}", mediaHandler.DeleteMedia)
		})

		// Workspace routes
		r.Route("/api/workspaces", func(r chi.Router) {
			r.Use(mw.AdminOnly) // Only admins can manage workspaces
			r.Post("/", workspaceHandler.CreateWorkspace)
			r.Get("/", workspaceHandler.ListWorkspaces)
			r.Get("/{id}", workspaceHandler.GetWorkspace)
			r.Put("/{id}", workspaceHandler.UpdateWorkspace)
			r.Delete("/{id}", workspaceHandler.DeleteWorkspace)
			
			// User-workspace association routes
			r.Post("/{id}/users", workspaceHandler.AddUserToWorkspace)
			r.Delete("/{id}/users/{userId}", workspaceHandler.RemoveUserFromWorkspace)
		})

		// User routes
		r.Route("/api/users", func(r chi.Router) {
			r.Use(mw.AdminOnly) // Only admins can manage users
			r.Post("/", userHandler.CreateUser)
			r.Get("/", userHandler.ListUsers)
			r.Get("/{id}", userHandler.GetUser)
			r.Put("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
		})

		// Current user info
		r.Get("/api/me", userHandler.GetCurrentUser)
		r.Put("/api/me", userHandler.UpdateCurrentUser)
		r.Put("/api/me/password", userHandler.ChangePassword)
	})

	// Set up admin UI
	fmt.Println("Setting up admin UI routes...")
	
	// Find the embedded filesystem
	adminUIFS, err := fs.Sub(adminUI, "web/admin/dist")
	if err != nil {
		fmt.Printf("Error creating subfolder for admin UI: %v\n", err)
		// Fall back to root if subfolder fails
		adminUIFS = adminUI
	}
	
	// List all files in the embedded filesystem for debugging
	fmt.Println("Contents of embedded filesystem:")
	listEmbeddedFiles(adminUIFS, "")
	
	// First try direct HTTP file server - this is the cleanest approach if it works
	fsServer := http.FileServer(http.FS(adminUIFS))
	
	// Serve static files and SPA routes
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Skip API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		fmt.Printf("Request path: %s\n", r.URL.Path)
		
		// For root path or paths that don't exist as files, serve index.html
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		if reqPath == "" || reqPath == "admin" {
			reqPath = "index.html"
		}
		
		// Try to open the file to check if it exists
		f, err := adminUIFS.Open(reqPath)
		if err != nil {
			// If file doesn't exist, serve index.html for SPA routing
			fmt.Printf("File not found: %s, serving index.html instead\n", reqPath)
			indexData, err := fs.ReadFile(adminUIFS, "index.html")
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(indexData)
			return
		}
		f.Close()
		
		// File exists, serve it directly
		fmt.Printf("Serving file: %s\n", reqPath)
		
		// Need to rewrite the URL Path to match the expected file structure
		originalPath := r.URL.Path
		r.URL.Path = "/" + reqPath
		fsServer.ServeHTTP(w, r)
		// Restore the original path in case middleware depends on it
		r.URL.Path = originalPath
	})

	return r
}

// Helper function to list all files in the embedded filesystem
func listEmbeddedFiles(fsys fs.FS, dir string) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dir, err)
		return
	}
	
	for _, entry := range entries {
		path := dir
		if path != "" {
			path += "/"
		}
		path += entry.Name()
		
		if entry.IsDir() {
			fmt.Printf("Directory: %s\n", path)
			listEmbeddedFiles(fsys, path)
		} else {
			info, err := entry.Info()
			if err == nil {
				fmt.Printf("File: %s (size: %d bytes)\n", path, info.Size())
			} else {
				fmt.Printf("File: %s\n", path)
			}
		}
	}
}