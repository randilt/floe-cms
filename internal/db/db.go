// internal/db/db.go
package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/randilt/floe-cms/internal/config"
	"github.com/randilt/floe-cms/internal/models"
)

// DB is a wrapper around gorm.DB
type DB struct {
	*gorm.DB
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Initialize initializes the database connection
func Initialize(config config.DatabaseConfig) (*DB, error) {
	var dialector gorm.Dialector

	switch config.Type {
	case "sqlite":
		dialector = sqlite.Open(config.URL)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Name)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// MigrateDatabase runs database migrations
func MigrateDatabase(db *DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Workspace{},
		&models.Content{},
		&models.Media{},
		&models.ContentType{},
		&models.UserWorkspace{},
		&models.RefreshToken{},
	)
}

// ExecuteWithTransaction executes the given function within a transaction
func ExecuteWithTransaction(db *DB, fn func(tx *gorm.DB) error) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}