package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewR2Config_UsesEnvSettings(t *testing.T) {
	configurator := viper.New()
	configurator.Set("R2_ACCOUNT_ID", "acct-123")
	configurator.Set("R2_ACCESS_KEY_ID", "key-id")
	configurator.Set("R2_SECRET_ACCESS_KEY", "key-secret")
	configurator.Set("R2_BUCKET_NAME", "media")
	configurator.Set("R2_PUBLIC_BASE_URL", "https://media.example.com")
	configurator.Set("R2_MAX_GRANT_EXPIRY_SECONDS", 60)
	configurator.Set("R2_UPLOAD_GRANT_EXPIRY_SECONDS", 30)

	cfg := NewR2Config(configurator)

	require.Equal(t, "acct-123", cfg.AccountID)
	require.Equal(t, "key-id", cfg.AccessKeyID)
	require.Equal(t, "key-secret", cfg.SecretAccessKey)
	require.Equal(t, "media", cfg.BucketName)
	require.Equal(t, "https://media.example.com", cfg.PublicBaseURL)
	require.Equal(t, 60*time.Second, cfg.MaxGrantExpiry)
	require.Equal(t, 30*time.Second, cfg.UploadGrantExpiry)
}

func TestNewR2Config_DefaultsGrantExpiries(t *testing.T) {
	configurator := viper.New()

	cfg := NewR2Config(configurator)

	require.Equal(t, 15*time.Minute, cfg.MaxGrantExpiry)
	require.Equal(t, 15*time.Minute, cfg.UploadGrantExpiry)
}

func TestR2Config_Validate(t *testing.T) {
	valid := func() R2Config {
		return R2Config{
			AccountID:         "acct-123",
			AccessKeyID:       "key-id",
			SecretAccessKey:   "key-secret",
			BucketName:        "media",
			MaxGrantExpiry:    15 * time.Minute,
			UploadGrantExpiry: 15 * time.Minute,
		}
	}

	tests := []struct {
		name    string
		mutate  func(*R2Config)
		wantErr bool
	}{
		{name: "valid", mutate: func(c *R2Config) {}, wantErr: false},
		{name: "valid without public base url", mutate: func(c *R2Config) { c.PublicBaseURL = "" }, wantErr: false},
		{name: "empty account id", mutate: func(c *R2Config) { c.AccountID = "" }, wantErr: true},
		{name: "empty access key id", mutate: func(c *R2Config) { c.AccessKeyID = "" }, wantErr: true},
		{name: "empty secret access key", mutate: func(c *R2Config) { c.SecretAccessKey = "" }, wantErr: true},
		{name: "empty bucket name", mutate: func(c *R2Config) { c.BucketName = "" }, wantErr: true},
		{name: "zero max grant expiry", mutate: func(c *R2Config) { c.MaxGrantExpiry = 0 }, wantErr: true},
		{name: "negative max grant expiry", mutate: func(c *R2Config) { c.MaxGrantExpiry = -time.Second }, wantErr: true},
		{name: "zero upload grant expiry", mutate: func(c *R2Config) { c.UploadGrantExpiry = 0 }, wantErr: true},
		{name: "negative upload grant expiry", mutate: func(c *R2Config) { c.UploadGrantExpiry = -time.Second }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid()
			tt.mutate(&cfg)

			err := cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
