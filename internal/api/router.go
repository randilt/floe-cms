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

		// Content management routes with explicit workspace
		r.Route("/api/workspaces/{workspaceId}/content", func(r chi.Router) {
			r.Post("/", contentHandler.CreateContent)
			r.Get("/", contentHandler.ListContent)
			r.Get("/{id}", contentHandler.GetContent)
			r.Put("/{id}", contentHandler.UpdateContent)
			r.Delete("/{id}", contentHandler.DeleteContent)
		})

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

	// Extract the embedded filesystem
	adminUIFS, err := fs.Sub(adminUI, "web/admin/dist")
	if err != nil {
		fmt.Printf("Error creating subfolder for admin UI: %v\n", err)
		adminUIFS = adminUI
	}

	// Embed file logging for debugging - list all available files
	fmt.Println("Contents of embedded filesystem:")
	entries, err := fs.ReadDir(adminUIFS, ".")
	if err != nil {
		fmt.Printf("Error reading root directory: %v\n", err)
	} else {
		for _, entry := range entries {
			fmt.Printf("Root entry: %s (is dir: %v)\n", entry.Name(), entry.IsDir())
		}
		
		// Check assets directory specifically if it exists
		assetsEntries, err := fs.ReadDir(adminUIFS, "assets")
		if err != nil {
			fmt.Printf("Error reading assets directory: %v\n", err)
		} else {
			for _, entry := range assetsEntries {
				fmt.Printf("Asset file: %s\n", entry.Name())
			}
		}
	}

	// Create a custom file server wrapper that sets appropriate headers
	fileServerWithHeaders := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		// Set appropriate content type based on file extension
		if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		} else if strings.HasSuffix(path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		}
		
		// Set caching headers for static assets
		if strings.HasPrefix(path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}
		
		// Log the request
		fmt.Printf("Serving file: %s\n", path)
		
		// Serve the file using the standard file server
		http.FileServer(http.FS(adminUIFS)).ServeHTTP(w, r)
	})

	// Serve static files from the assets directory using a parameter instead of wildcard
	r.Get("/assets/{file}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Adjust path to include the assets directory
		filePath := strings.TrimPrefix(r.URL.Path, "/")
		r.URL.Path = "/" + filePath
		fileServerWithHeaders.ServeHTTP(w, r)
	}))

	// Serve common root files
	r.Get("/favicon.ico", fileServerWithHeaders)
	r.Get("/vite.svg", fileServerWithHeaders)

	// Special handler for CSS/JS files that might be requested directly
	r.Get("/{file}.css", fileServerWithHeaders)
	r.Get("/{file}.js", fileServerWithHeaders)
	r.Get("/{file}.js.map", fileServerWithHeaders)

	// SPA route handler - serve index.html for all other routes
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Skip API requests
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		
		fmt.Printf("SPA route request: %s\n", r.URL.Path)
		
		// Try to see if this is a static file first that we missed in our specific handlers
		if strings.Contains(r.URL.Path, ".") {
			fileServerWithHeaders.ServeHTTP(w, r)
			return
		}
		
		// Otherwise serve index.html for client-side routing
		indexPath := "index.html"
		indexData, err := fs.ReadFile(adminUIFS, indexPath)
		if err != nil {
			fmt.Printf("Error reading index.html: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexData)
	})

	return r
}