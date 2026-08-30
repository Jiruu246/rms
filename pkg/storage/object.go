package storage

import (
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
)

// ObjectKey identifies an object within a storage provider's namespace. It is
// provider-neutral: no bucket, container, or path-separator semantics are
// implied here — a provider implementation is free to interpret it.
type ObjectKey string

// Validate reports whether k is usable as an object key.
func (k ObjectKey) Validate() error {
	if k == "" {
		return apperr.Invalid("object key must not be empty")
	}
	return nil
}

// ObjectMetadata is what a storage provider reports about an object it holds.
type ObjectMetadata struct {
	SizeBytes    int64
	ContentType  string
	LastModified time.Time
}

// StoredObject is an object confirmed present in a storage provider,
// identified by its key together with the metadata the provider reported.
type StoredObject struct {
	Key      ObjectKey
	Metadata ObjectMetadata
}
