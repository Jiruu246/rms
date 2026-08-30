package integration_tests

import (
	"context"
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/pkg/storage"
)

// fakeStorageProvider is an in-memory storage.StorageProvider used to build
// a *server.Server for HTTP-level integration tests, so those tests don't
// need real R2 credentials just to start a server.
type fakeStorageProvider struct {
	objects map[storage.ObjectKey]storage.ObjectMetadata
}

func newFakeStorageProvider() *fakeStorageProvider {
	return &fakeStorageProvider{
		objects: make(map[storage.ObjectKey]storage.ObjectMetadata),
	}
}

func (f *fakeStorageProvider) CreateUploadGrant(_ context.Context, req storage.UploadRequest) (*storage.UploadGrant, error) {
	return &storage.UploadGrant{
		Method:    "PUT",
		URL:       "https://example.com/" + string(req.Key),
		ObjectKey: req.Key,
		ExpiresAt: time.Now().Add(req.Expiry),
	}, nil
}

func (f *fakeStorageProvider) StatObject(_ context.Context, key storage.ObjectKey) (*storage.StoredObject, error) {
	meta, ok := f.objects[key]
	if !ok {
		return nil, apperr.NotFound("object %q", key)
	}
	return &storage.StoredObject{Key: key, Metadata: meta}, nil
}

func (f *fakeStorageProvider) DeleteObject(_ context.Context, key storage.ObjectKey) error {
	delete(f.objects, key)
	return nil
}
