package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

// R2Config holds credentials and settings for a Cloudflare R2 bucket, used by
// internal/r2storage to build a pkg/storage.StorageProvider.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	// PublicBaseURL is the base URL objects are served from once uploaded
	// (e.g. a custom domain or R2.dev subdomain), used to build display
	// URLs for stored objects. Optional: leave empty if objects are not
	// served publicly.
	PublicBaseURL string
	// MaxGrantExpiry caps how long a presigned upload grant may remain
	// usable, regardless of what a caller requests.
	MaxGrantExpiry time.Duration
	// UploadGrantExpiry is the expiry the media upload service requests when
	// issuing a new upload grant. It is independent of MaxGrantExpiry (that
	// is the provider-side cap); this is the application-level default,
	// tunable without touching provider settings.
	UploadGrantExpiry time.Duration
}

func NewR2Config(configurator *viper.Viper) R2Config {
	configurator.SetDefault("R2_MAX_GRANT_EXPIRY", 15*60)    // seconds
	configurator.SetDefault("R2_UPLOAD_GRANT_EXPIRY", 15*60) // seconds

	return R2Config{
		AccountID:         configurator.GetString("R2_ACCOUNT_ID"),
		AccessKeyID:       configurator.GetString("R2_ACCESS_KEY_ID"),
		SecretAccessKey:   configurator.GetString("R2_SECRET_ACCESS_KEY"),
		BucketName:        configurator.GetString("R2_BUCKET_NAME"),
		PublicBaseURL:     configurator.GetString("R2_PUBLIC_BASE_URL"),
		MaxGrantExpiry:    time.Duration(configurator.GetInt("R2_MAX_GRANT_EXPIRY")) * time.Second,
		UploadGrantExpiry: time.Duration(configurator.GetInt("R2_UPLOAD_GRANT_EXPIRY")) * time.Second,
	}
}

// Validate reports whether cfg has everything required to build an R2
// client. PublicBaseURL is intentionally not required.
func (cfg R2Config) Validate() error {
	if cfg.AccountID == "" {
		return errors.New("r2: account id must not be empty")
	}
	if cfg.AccessKeyID == "" {
		return errors.New("r2: access key id must not be empty")
	}
	if cfg.SecretAccessKey == "" {
		return errors.New("r2: secret access key must not be empty")
	}
	if cfg.BucketName == "" {
		return errors.New("r2: bucket name must not be empty")
	}
	if cfg.MaxGrantExpiry <= 0 {
		return errors.New("r2: max grant expiry must be positive")
	}
	if cfg.UploadGrantExpiry <= 0 {
		return errors.New("r2: upload grant expiry must be positive")
	}
	return nil
}
