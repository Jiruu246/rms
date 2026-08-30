package config

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration.
type Config struct {
	Env              string
	Port             int
	LogLevel         string
	DatabaseURL      string
	PostgresUser     string
	PostgresPassword string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	ShutdownTimeout  time.Duration
	AllowedOrigins   []string
	AccessTokenExp   time.Duration
	RefreshTokenExp  time.Duration
	CookieConfig     CookieConfig
	AuthConfig       AuthConfig
	R2Config         R2Config
}

// Load reads configuration from environment variables and optional file.
func Load() (*Config, error) {
	configurator := viper.New()
	configurator.SetEnvPrefix("APP")
	configurator.AutomaticEnv()

	// defaults
	configurator.SetDefault("ENV", "development")
	configurator.SetDefault("PORT", 8080)
	configurator.SetDefault("LOG_LEVEL", "info")
	configurator.SetDefault("READ_TIMEOUT", 15)
	configurator.SetDefault("WRITE_TIMEOUT", 15)
	configurator.SetDefault("SHUTDOWN_TIMEOUT", 15)
	configurator.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")

	CookieConfig := NewCookieConfig(configurator)
	AuthConfig := NewAuthConfig(configurator)
	R2Config := NewR2Config(configurator)

	cfg := &Config{
		Env:              configurator.GetString("ENV"),
		Port:             configurator.GetInt("PORT"),
		LogLevel:         configurator.GetString("LOG_LEVEL"),
		DatabaseURL:      configurator.GetString("DATABASE_URL"),
		PostgresUser:     configurator.GetString("POSTGRES_USER"),
		PostgresPassword: configurator.GetString("POSTGRES_PASSWORD"),
		ReadTimeout:      time.Duration(configurator.GetInt("READ_TIMEOUT")) * time.Second,
		WriteTimeout:     time.Duration(configurator.GetInt("WRITE_TIMEOUT")) * time.Second,
		ShutdownTimeout:  time.Duration(configurator.GetInt("SHUTDOWN_TIMEOUT")) * time.Second,
		AllowedOrigins:   strings.Split(configurator.GetString("ALLOWED_ORIGINS"), ","),
		CookieConfig:     CookieConfig,
		AuthConfig:       AuthConfig,
		R2Config:         R2Config,
	}

	if cfg.Port <= 0 {
		return nil, errors.New("invalid port")
	}

	return cfg, nil
}

func LoadTestConfig() (*Config, error) {
	configurator := viper.New()
	configurator.SetEnvPrefix("APP")
	configurator.AutomaticEnv()

	configurator.SetDefault("ENV", "testing")

	// R2 credentials default to fake-but-valid placeholders so tests can
	// construct a full server (and its storage.StorageProvider) without real
	// credentials. Safe because CreateUploadGrant only presigns locally — it
	// never makes a network call (see internal/r2storage/provider.go) — so
	// nothing here ever actually talks to R2 unless a test calls
	// StatObject/DeleteObject against a real bucket, which none do today.
	configurator.SetDefault("R2_ACCOUNT_ID", "test-account-id")
	configurator.SetDefault("R2_ACCESS_KEY_ID", "test-access-key-id")
	configurator.SetDefault("R2_SECRET_ACCESS_KEY", "test-secret-access-key")
	configurator.SetDefault("R2_BUCKET_NAME", "test-bucket")

	cookieConfig := NewCookieConfig(configurator)
	AuthConfig := NewAuthConfig(configurator)
	R2Config := NewR2Config(configurator)

	cfg := &Config{
		Env:              configurator.GetString("ENV"),
		DatabaseURL:      configurator.GetString("DATABASE_URL"),
		PostgresUser:     configurator.GetString("POSTGRES_USER"),
		PostgresPassword: configurator.GetString("POSTGRES_PASSWORD"),
		CookieConfig:     cookieConfig,
		AuthConfig:       AuthConfig,
		R2Config:         R2Config,
	}

	return cfg, nil
}
