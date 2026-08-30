package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestUploadRequest_Validate(t *testing.T) {
	valid := func() storage.UploadRequest {
		return storage.UploadRequest{
			Key:    "media/avatars/123.png",
			Expiry: 5 * time.Minute,
		}
	}

	tests := []struct {
		name    string
		mutate  func(*storage.UploadRequest)
		wantErr bool
	}{
		{name: "valid", mutate: func(r *storage.UploadRequest) {}, wantErr: false},
		{name: "empty key", mutate: func(r *storage.UploadRequest) { r.Key = "" }, wantErr: true},
		{name: "zero expiry", mutate: func(r *storage.UploadRequest) { r.Expiry = 0 }, wantErr: true},
		{name: "negative expiry", mutate: func(r *storage.UploadRequest) { r.Expiry = -time.Second }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid()
			tt.mutate(&req)

			err := req.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, apperr.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// fakeStorageProvider is a minimal storage.StorageProvider for testing pure
// composition logic like EnforceMaxSize without any real backend.
type fakeStorageProvider struct {
	statObject  *storage.StoredObject
	statErr     error
	deleteErr   error
	deletedKeys []storage.ObjectKey
}

func (f *fakeStorageProvider) CreateUploadGrant(ctx context.Context, req storage.UploadRequest) (*storage.UploadGrant, error) {
	panic("not used by these tests")
}

func (f *fakeStorageProvider) StatObject(ctx context.Context, key storage.ObjectKey) (*storage.StoredObject, error) {
	return f.statObject, f.statErr
}

func (f *fakeStorageProvider) DeleteObject(ctx context.Context, key storage.ObjectKey) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return f.deleteErr
}

func TestUploadGrant_Expired(t *testing.T) {
	expiresAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	grant := storage.UploadGrant{ExpiresAt: expiresAt}

	tests := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "before expiry", at: expiresAt.Add(-time.Minute), want: false},
		{name: "at expiry", at: expiresAt, want: true},
		{name: "after expiry", at: expiresAt.Add(time.Minute), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, grant.Expired(tt.at))
		})
	}
}
