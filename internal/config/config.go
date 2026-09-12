package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
	EnvTest Environment = "test"
)

type Config struct {
	Env              Environment
	Port             int
	LogLevel         string
	DatabaseURL      string
	PostgresUser     string
	PostgresPassword string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	ShutdownTimeout  time.Duration
	AllowedOrigins   []string
	CookieConfig     CookieConfig
	AuthConfig       AuthConfig
}

const configDir = "configs"

func Load() (*Config, error) {
	env, err := resolveEnv()
	if err != nil {
		return nil, err
	}

	configurator, err := readConfigFiles(configDir, env)
	if err != nil {
		return nil, err
	}
	bindEnv(configurator)

	cfg := buildConfig(configurator, env)

	if cfg.Port <= 0 {
		return nil, errors.New("invalid port")
	}

	return cfg, nil
}

// LoadTestConfig loads configuration for integration tests, layering
// configs/base.yaml + configs/test.yaml exactly like Load() does, plus
// whatever APP_* environment variables are set (secrets, DB connection info).
// It locates configs/ relative to this source file rather than the process's
// working directory, since `go test` runs each package's binary from that
// package's own directory (e.g. integration_tests/), not the repo root.

// TODO: Consider refactoring this to all use Load()
func LoadTestConfig() (*Config, error) {
	configurator, err := readConfigFiles(testConfigDir(), EnvTest)
	if err != nil {
		return nil, err
	}
	bindEnv(configurator)

	return buildConfig(configurator, EnvTest), nil
}

func testConfigDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", configDir)
}

func resolveEnv() (Environment, error) {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		return "", errors.New("APP_ENV is required (dev, prod, or test)")
	}
	return Environment(env), nil
}

func readConfigFiles(dir string, env Environment) (*viper.Viper, error) {
	configurator := viper.New()
	configurator.SetConfigType("yaml")
	configurator.AddConfigPath(dir)

	configurator.SetConfigName("base")
	if err := configurator.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read base config: %w", err)
	}

	configurator.SetConfigName(string(env))
	if err := configurator.MergeInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("failed to read %s config: %w", env, err)
		}
	}

	return configurator, nil
}

func bindEnv(configurator *viper.Viper) {
	configurator.SetEnvPrefix("APP")
	configurator.AutomaticEnv()
}

func buildConfig(configurator *viper.Viper, env Environment) *Config {
	cookieConfig := NewCookieConfig(configurator)
	authConfig := NewAuthConfig(configurator)

	return &Config{
		Env:              env,
		Port:             configurator.GetInt("PORT"),
		LogLevel:         configurator.GetString("LOG_LEVEL"),
		DatabaseURL:      configurator.GetString("DATABASE_URL"),
		PostgresUser:     configurator.GetString("POSTGRES_USER"),
		PostgresPassword: configurator.GetString("POSTGRES_PASSWORD"),
		ReadTimeout:      time.Duration(configurator.GetInt("READ_TIMEOUT")) * time.Second,
		WriteTimeout:     time.Duration(configurator.GetInt("WRITE_TIMEOUT")) * time.Second,
		ShutdownTimeout:  time.Duration(configurator.GetInt("SHUTDOWN_TIMEOUT")) * time.Second,
		AllowedOrigins:   allowedOrigins(configurator),
		CookieConfig:     cookieConfig,
		AuthConfig:       authConfig,
	}
}

// allowedOrigins reads ALLOWED_ORIGINS as a comma-separated string — it's only
// ever set via the APP_ALLOWED_ORIGINS env var, never a configs/*.yaml list.
func allowedOrigins(configurator *viper.Viper) []string {
	v := configurator.GetString("ALLOWED_ORIGINS")
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}
