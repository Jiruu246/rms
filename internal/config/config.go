package config

import (
	"errors"
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
	ShutdownTimeout  int
	AllowedOrigins   []string
	AccessTokenExp   time.Duration
	RefreshTokenExp  time.Duration
	CookieConfig     CookieConfig
	AuthConfig       AuthConfig
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
	configurator.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"})

	CookieConfig := NewCookieConfig(configurator)
	AuthConfig := NewAuthConfig(configurator)

	cfg := &Config{
		Env:              configurator.GetString("ENV"),
		Port:             configurator.GetInt("PORT"),
		LogLevel:         configurator.GetString("LOG_LEVEL"),
		DatabaseURL:      configurator.GetString("DATABASE_URL"),
		PostgresUser:     configurator.GetString("POSTGRES_USER"),
		PostgresPassword: configurator.GetString("POSTGRES_PASSWORD"),
		ReadTimeout:      time.Duration(configurator.GetInt("READ_TIMEOUT")) * time.Second,
		WriteTimeout:     time.Duration(configurator.GetInt("WRITE_TIMEOUT")) * time.Second,
		ShutdownTimeout:  configurator.GetInt("SHUTDOWN_TIMEOUT"),
		AllowedOrigins:   configurator.GetStringSlice("ALLOWED_ORIGINS"),
		CookieConfig:     CookieConfig,
		AuthConfig:       AuthConfig,
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

	cookieConfig := NewCookieConfig(configurator)
	AuthConfig := NewAuthConfig(configurator)

	cfg := &Config{
		Env:              configurator.GetString("ENV"),
		DatabaseURL:      configurator.GetString("DATABASE_URL"),
		PostgresUser:     configurator.GetString("POSTGRES_USER"),
		PostgresPassword: configurator.GetString("POSTGRES_PASSWORD"),
		CookieConfig:     cookieConfig,
		AuthConfig:       AuthConfig,
	}

	return cfg, nil
}
