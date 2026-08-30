package storage

import (
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
)

type UploadRequest struct {
	// Key is the object key the caller wants the upload to end up under.
	Key ObjectKey
	// Expiry is how long the resulting grant should remain usable.
	Expiry time.Duration
}

func (r UploadRequest) Validate() error {
	if err := r.Key.Validate(); err != nil {
		return err
	}
	if r.Expiry <= 0 {
		return apperr.Invalid("upload request %q: expiry must be positive", r.Key)
	}
	return nil
}

// UploadGrant is everything a client needs to perform a direct upload to a
// storage provider without the application server proxying the bytes.
type UploadGrant struct {
	// Method is the HTTP method the client must use (e.g. "PUT").
	Method string
	// URL is the address the client uploads to.
	URL string
	// Headers are the request headers the client must send with the upload.
	Headers map[string]string
	// ObjectKey is the key the uploaded object will be stored under.
	ObjectKey ObjectKey
	// ExpiresAt is when the grant stops being usable.
	ExpiresAt time.Time
}

func (g UploadGrant) Expired(at time.Time) bool {
	return !at.Before(g.ExpiresAt)
}
