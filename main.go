// cmd/floe-cms/main.go
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/randilt/floe-cms/internal/api"
	"github.com/randilt/floe-cms/internal/auth"
	"github.com/randilt/floe-cms/internal/config"
	"github.com/randilt/floe-cms/internal/db"
	"github.com/randilt/floe-cms/internal/storage"
)

//go:embed web/admin/dist
var AdminUIAssets embed.FS

func main() {
	var configPath string
	var port int
	var resetAdmin bool
	var dbURL string

	// Parse command line flags
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.IntVar(&port, "port", 0, "Override port defined in configuration")
	flag.BoolVar(&resetAdmin, "reset-admin", false, "Reset admin credentials")
	flag.StringVar(&dbURL, "db-url", "", "Override database URL defined in configuration")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override configuration with flags if provided
	if port > 0 {
		cfg.Server.Port = port
	}
	if dbURL != "" {
		cfg.Database.URL = dbURL
	}

	// Configure logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize database connection
	database, err := db.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := db.MigrateDatabase(database); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize storage
	storageManager := storage.NewLocalStorage(cfg.Storage.UploadsDir)

	// Initialize authentication
	authManager := auth.NewManager(database, cfg.Auth)

	// Reset admin user if requested
	if resetAdmin {
		if err := auth.ResetAdminUser(database, cfg.Auth.AdminEmail, cfg.Auth.AdminPassword); err != nil {
			log.Fatalf("Failed to reset admin user: %v", err)
		}
		fmt.Println("Admin user reset successfully")
		return
	}

	// Create default admin if doesn't exist
	if err := auth.EnsureAdminExists(database, cfg.Auth.AdminEmail, cfg.Auth.AdminPassword); err != nil {
		log.Fatalf("Failed to ensure admin exists: %v", err)
	}

	// Initialize API router
	router := api.NewRouter(authManager, database, storageManager, AdminUIAssets, cfg)

	// Configure HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.Timeouts.Read) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeouts.Write) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.Timeouts.Idle) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Floe CMS server", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.GracefulShutdown)*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}