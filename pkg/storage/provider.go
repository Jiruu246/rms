package storage

import "context"

// StorageProvider is the boundary between the application and a concrete
// object-storage backend (e.g. Cloudflare R2, S3). Implementations live in
// infrastructure packages; none of their provider-specific concepts may leak
// through this interface.
type StorageProvider interface {
	// CreateUploadGrant issues a grant that lets a client upload an object
	// directly to the provider without proxying bytes through the
	// application server. req should pass Validate before this is called.
	CreateUploadGrant(ctx context.Context, req UploadRequest) (*UploadGrant, error)

	// StatObject returns metadata for the object stored under key. It
	// returns an apperr.ErrNotFound-wrapping error if no object exists
	// under that key.
	StatObject(ctx context.Context, key ObjectKey) (*StoredObject, error)

	// DeleteObject removes the object stored under key. It is idempotent:
	// deleting a key that does not exist is not an error, matching the
	// native behavior of object stores like S3/R2.
	DeleteObject(ctx context.Context, key ObjectKey) error
}
