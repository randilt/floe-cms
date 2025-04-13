// internal/config/config.go
package config

import (
	"math/rand"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

// ServerConfig holds server related configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	GracefulShutdown int          `mapstructure:"graceful_shutdown"`
	Timeouts        TimeoutConfig `mapstructure:"timeouts"`
}

// TimeoutConfig holds server timeout configurations
type TimeoutConfig struct {
	Read  int `mapstructure:"read"`
	Write int `mapstructure:"write"`
	Idle  int `mapstructure:"idle"`
}

// DatabaseConfig holds database related configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// AuthConfig holds authentication related configuration
type AuthConfig struct {
	JWTSecret           string `mapstructure:"jwt_secret"`
	AccessTokenExpiry   int    `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry  int    `mapstructure:"refresh_token_expiry"`
	AdminEmail          string `mapstructure:"admin_email"`
	AdminPassword       string `mapstructure:"admin_password"`
	PasswordMinLength   int    `mapstructure:"password_min_length"`
	RateLimitRequests   int    `mapstructure:"rate_limit_requests"`
	RateLimitExpiry     int    `mapstructure:"rate_limit_expiry"`
}

// StorageConfig holds storage related configuration
type StorageConfig struct {
	Type       string `mapstructure:"type"`
	UploadsDir string `mapstructure:"uploads_dir"`
}

// CacheConfig holds cache related configuration
type CacheConfig struct {
	Type     string `mapstructure:"type"`
	RedisURL string `mapstructure:"redis_url"`
	TTL      int    `mapstructure:"ttl"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Set defaults
	config := defaultConfig()

	// Initialize viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetEnvPrefix("FLOE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file, don't error if not found
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Set JWT secret from environment if not set
	jwtSecret := os.Getenv("FLOE_AUTH_JWT_SECRET")
	if jwtSecret != "" {
		config.Auth.JWTSecret = jwtSecret
	}

	// Generate a JWT secret if not set
	if config.Auth.JWTSecret == "" {
		config.Auth.JWTSecret = generateRandomString(32)
	}

	return config, nil
}

// defaultConfig returns a configuration with default values
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			GracefulShutdown: 30,
			Timeouts: TimeoutConfig{
				Read:  15,
				Write: 15,
				Idle:  60,
			},
		},
		Database: DatabaseConfig{
			Type:     "sqlite",
			URL:      "floe.db",
			Host:     "localhost",
			Port:     5432,
			Username: "floe",
			Password: "floe",
			Name:     "floe",
			SSLMode:  "disable",
		},
		Auth: AuthConfig{
			JWTSecret:          "",
			AccessTokenExpiry:  15 * 60,  // 15 minutes
			RefreshTokenExpiry: 7 * 24 * 60 * 60, // 7 days
			AdminEmail:         "admin@floe.cms",
			AdminPassword:      "adminpassword",
			PasswordMinLength:  8,
			RateLimitRequests:  60,  // 60 requests
			RateLimitExpiry:    60,  // per minute
		},
		Storage: StorageConfig{
			Type:       "local",
			UploadsDir: "./uploads",
		},
		Cache: CacheConfig{
			Type:     "memory",
			RedisURL: "redis://localhost:6379/0",
			TTL:      300, // 5 minutes
		},
	}
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(rand.Int63())%len(charset)]
	}
	return string(b)
}