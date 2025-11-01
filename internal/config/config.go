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
}

// Load reads configuration from environment variables and optional file.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// defaults
	v.SetDefault("ENV", "development")
	v.SetDefault("PORT", 8080)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("READ_TIMEOUT", 15)
	v.SetDefault("WRITE_TIMEOUT", 15)
	v.SetDefault("SHUTDOWN_TIMEOUT", 15)
	v.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"})

	cfg := &Config{
		Env:              v.GetString("ENV"),
		Port:             v.GetInt("PORT"),
		LogLevel:         v.GetString("LOG_LEVEL"),
		DatabaseURL:      v.GetString("DATABASE_URL"),
		PostgresUser:     v.GetString("POSTGRES_USER"),
		PostgresPassword: v.GetString("POSTGRES_PASSWORD"),
		ReadTimeout:      time.Duration(v.GetInt("READ_TIMEOUT")) * time.Second,
		WriteTimeout:     time.Duration(v.GetInt("WRITE_TIMEOUT")) * time.Second,
		ShutdownTimeout:  v.GetInt("SHUTDOWN_TIMEOUT"),
		AllowedOrigins:   v.GetStringSlice("ALLOWED_ORIGINS"),
	}

	if cfg.Port <= 0 {
		return nil, errors.New("invalid port")
	}

	return cfg, nil
}
