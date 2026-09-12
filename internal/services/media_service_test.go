package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/mediaasset"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/storage"
)

// fakeStorageProvider is a real (non-mock) in-memory implementation of
// storage.StorageProvider, per the task's request for a fake rather than an
// assertion-driven mock — CreateUpload/ConsumeUpload exercise it exactly as
// a real provider would behave.
type fakeStorageProvider struct {
	grantErr  error
	deleteErr error
	objects   map[storage.ObjectKey]storage.ObjectMetadata
	statErr   map[storage.ObjectKey]error
}

func newFakeStorageProvider() *fakeStorageProvider {
	return &fakeStorageProvider{
		objects: make(map[storage.ObjectKey]storage.ObjectMetadata),
		statErr: make(map[storage.ObjectKey]error),
	}
}

func (f *fakeStorageProvider) CreateUploadGrant(_ context.Context, req storage.UploadRequest) (*storage.UploadGrant, error) {
	if f.grantErr != nil {
		return nil, f.grantErr
	}
	return &storage.UploadGrant{
		Method:    "PUT",
		URL:       "https://example.com/" + string(req.Key),
		ObjectKey: req.Key,
		ExpiresAt: time.Now().Add(req.Expiry),
	}, nil
}

func (f *fakeStorageProvider) StatObject(_ context.Context, key storage.ObjectKey) (*storage.StoredObject, error) {
	if err, ok := f.statErr[key]; ok {
		return nil, err
	}
	meta, ok := f.objects[key]
	if !ok {
		return nil, apperr.NotFound("object %q", key)
	}
	return &storage.StoredObject{Key: key, Metadata: meta}, nil
}

func (f *fakeStorageProvider) DeleteObject(_ context.Context, key storage.ObjectKey) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.objects, key)
	return nil
}

func (f *fakeStorageProvider) putObject(key storage.ObjectKey, meta storage.ObjectMetadata) {
	f.objects[key] = meta
}

type MockMediaUploadRepository struct {
	mock.Mock
}

func (m *MockMediaUploadRepository) Create(ctx context.Context, params repos.CreateMediaUploadParams) (*ent.MediaUpload, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.MediaUpload), args.Error(1)
}

func (m *MockMediaUploadRepository) GetByID(ctx context.Context, ownerID, id uuid.UUID) (*ent.MediaUpload, error) {
	args := m.Called(ctx, ownerID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.MediaUpload), args.Error(1)
}

func (m *MockMediaUploadRepository) Consume(ctx context.Context, ownerID, id uuid.UUID, params repos.ConsumeMediaUploadParams) (*ent.MediaAsset, error) {
	args := m.Called(ctx, ownerID, id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.MediaAsset), args.Error(1)
}

func (m *MockMediaUploadRepository) Fail(ctx context.Context, ownerID, id uuid.UUID) error {
	args := m.Called(ctx, ownerID, id)
	return args.Error(0)
}

type MockMediaAssetRepository struct {
	mock.Mock
}

func (m *MockMediaAssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.MediaAsset, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.MediaAsset), args.Error(1)
}

func (m *MockMediaAssetRepository) GetByUploadID(ctx context.Context, uploadID uuid.UUID) (*ent.MediaAsset, error) {
	args := m.Called(ctx, uploadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.MediaAsset), args.Error(1)
}

func (m *MockMediaAssetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestMediaService_CreateUpload(t *testing.T) {
	actor := authz.Actor{UserID: uuid.New()}

	t.Run("successful upload grant", func(t *testing.T) {
		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		uploadID := uuid.New()
		mockUploadRepo.On("Create", mock.Anything, mock.MatchedBy(func(p repos.CreateMediaUploadParams) bool {
			return p.OwnerID == actor.UserID &&
				p.Purpose == mediaupload.PurposeMenuItemImage
		})).Return(&ent.MediaUpload{ID: uploadID}, nil)

		result, err := service.CreateUpload(t.Context(), actor, mediaupload.PurposeMenuItemImage)

		assert.NoError(t, err)
		assert.Equal(t, uploadID, result.UploadID)
		assert.Equal(t, "PUT", result.Upload.Method)
		assert.NotEmpty(t, result.Upload.ObjectKey)
		assert.Contains(t, result.Upload.ObjectKey, "media/"+actor.UserID.String()+"/")
		assert.False(t, result.ExpiresAt.IsZero())
		mockUploadRepo.AssertExpectations(t)
	})

	t.Run("invalid purpose", func(t *testing.T) {
		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		result, err := service.CreateUpload(t.Context(), actor, mediaupload.Purpose("not_a_real_purpose"))

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrInvalid)
		mockUploadRepo.AssertNotCalled(t, "Create")
	})

	t.Run("storage provider fails to grant", func(t *testing.T) {
		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.grantErr = assert.AnError
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		result, err := service.CreateUpload(t.Context(), actor, mediaupload.PurposeMenuItemImage)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockUploadRepo.AssertNotCalled(t, "Create")
	})
}

func TestMediaService_ConsumeUpload(t *testing.T) {
	newIssuedUpload := func(actor authz.Actor, uploadID uuid.UUID, objectKey string) *ent.MediaUpload {
		return &ent.MediaUpload{
			ID:        uploadID,
			OwnerID:   actor.UserID,
			Purpose:   mediaupload.PurposeMenuItemImage,
			ObjectKey: objectKey,
			Status:    mediaupload.StatusIssued,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}
	}

	t.Run("ownership failure", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).
			Return(nil, apperr.NotFound("media upload %s", uploadID))

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrNotFound)
		mockUploadRepo.AssertNotCalled(t, "Consume")
	})

	t.Run("expired upload", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)
		upload.ExpiresAt = time.Now().Add(-time.Minute)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrConflict)
		mockUploadRepo.AssertNotCalled(t, "Consume")
	})

	t.Run("duplicate consumption", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)
		upload.Status = mediaupload.StatusConsumed

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrConflict)
		mockUploadRepo.AssertNotCalled(t, "Consume")
	})

	t.Run("missing object", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider() // no object put for objectKey
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrConflict)
		mockUploadRepo.AssertNotCalled(t, "Consume")
	})

	t.Run("content type mismatch", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.putObject(storage.ObjectKey(objectKey), storage.ObjectMetadata{
			SizeBytes:   2048,
			ContentType: "application/pdf", // not in menu_item_image's allowed set
		})
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)
		mockUploadRepo.On("Fail", mock.Anything, actor.UserID, uploadID).Return(nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrInvalid)
		mockUploadRepo.AssertNotCalled(t, "Consume")
		mockUploadRepo.AssertExpectations(t)
		_, stillExists := provider.objects[storage.ObjectKey(objectKey)]
		assert.False(t, stillExists, "rejected object should have been deleted from storage")
	})

	t.Run("size exceeds purpose limit", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.putObject(storage.ObjectKey(objectKey), storage.ObjectMetadata{
			SizeBytes:   11 << 20, // menu_item_image caps at 10 MiB
			ContentType: "image/png",
		})
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)
		mockUploadRepo.On("Fail", mock.Anything, actor.UserID, uploadID).Return(nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrInvalid)
		mockUploadRepo.AssertNotCalled(t, "Consume")
		mockUploadRepo.AssertExpectations(t)
		_, stillExists := provider.objects[storage.ObjectKey(objectKey)]
		assert.False(t, stillExists, "rejected object should have been deleted from storage")
	})

	t.Run("cleanup failures do not mask the validation error", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.putObject(storage.ObjectKey(objectKey), storage.ObjectMetadata{
			SizeBytes:   2048,
			ContentType: "application/pdf", // upload expects image/png
		})
		provider.deleteErr = assert.AnError
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)
		mockUploadRepo.On("Fail", mock.Anything, actor.UserID, uploadID).Return(assert.AnError)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrInvalid)
		assert.Contains(t, err.Error(), "content type")
		mockUploadRepo.AssertExpectations(t)
	})

	t.Run("previously rejected upload cannot be retried", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)
		upload.Status = mediaupload.StatusFailed

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrConflict)
		mockUploadRepo.AssertNotCalled(t, "Consume")
		mockUploadRepo.AssertNotCalled(t, "Fail")
	})

	t.Run("successful consumption", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.putObject(storage.ObjectKey(objectKey), storage.ObjectMetadata{
			SizeBytes:   2048,
			ContentType: "image/png",
		})
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		assetID := uuid.New()
		expectedAsset := &ent.MediaAsset{
			ID:               assetID,
			UploadedByUserID: actor.UserID,
			UploadID:         uploadID,
			StorageKey:       objectKey,
			ContentType:      "image/png",
			SizeBytes:        2048,
			Status:           mediaasset.StatusActive,
		}
		mockUploadRepo.On("Consume", mock.Anything, actor.UserID, uploadID, repos.ConsumeMediaUploadParams{
			ContentType: "image/png",
			SizeBytes:   2048,
		}).Return(expectedAsset, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, "")

		assert.NoError(t, err)
		assert.Equal(t, assetID, result.ID)
		assert.Equal(t, objectKey, result.StorageKey)
		assert.Equal(t, "image/png", result.ContentType)
		assert.Equal(t, int64(2048), result.SizeBytes)
		assert.Equal(t, "active", result.Status)
		mockUploadRepo.AssertExpectations(t)
	})

	t.Run("matching expected purpose succeeds", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		upload := newIssuedUpload(actor, uploadID, objectKey)
		upload.Purpose = mediaupload.PurposeRestaurantLogo

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider()
		provider.putObject(storage.ObjectKey(objectKey), storage.ObjectMetadata{
			SizeBytes:   2048,
			ContentType: "image/png",
		})
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)
		mockUploadRepo.On("Consume", mock.Anything, actor.UserID, uploadID, mock.Anything).
			Return(&ent.MediaAsset{ID: uuid.New(), Status: mediaasset.StatusActive}, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, mediaupload.PurposeRestaurantLogo)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockUploadRepo.AssertExpectations(t)
	})

	t.Run("mismatched expected purpose is rejected before touching storage", func(t *testing.T) {
		actor := authz.Actor{UserID: uuid.New()}
		uploadID := uuid.New()
		objectKey := "media/" + actor.UserID.String() + "/" + uuid.NewString()

		// Upload session was created for menu_item_image (10 MiB cap), but
		// the caller wants to attach it as a restaurant_logo (5 MiB cap) —
		// this must be rejected without ever consulting the storage
		// provider, since it's a caller-side mismatch, not anything about
		// the object itself.
		upload := newIssuedUpload(actor, uploadID, objectKey)

		mockUploadRepo := new(MockMediaUploadRepository)
		mockAssetRepo := new(MockMediaAssetRepository)
		provider := newFakeStorageProvider() // no object put — proves StatObject is never called
		service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

		mockUploadRepo.On("GetByID", mock.Anything, actor.UserID, uploadID).Return(upload, nil)

		result, err := service.ConsumeUpload(t.Context(), actor, uploadID, mediaupload.PurposeRestaurantLogo)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, apperr.ErrInvalid)
		mockUploadRepo.AssertNotCalled(t, "Consume")
		mockUploadRepo.AssertNotCalled(t, "Fail")
	})
}

func TestMediaService_DeleteMedia(t *testing.T) {
	mockUploadRepo := new(MockMediaUploadRepository)
	mockAssetRepo := new(MockMediaAssetRepository)
	provider := newFakeStorageProvider()
	service := NewMediaService(noopTransactor{}, mockUploadRepo, mockAssetRepo, provider, 15*time.Minute)

	mediaID := uuid.New()
	actor := authz.Actor{UserID: uuid.New()}

	mockAssetRepo.On("Delete", mock.Anything, mediaID).Return(nil)

	err := service.DeleteMedia(t.Context(), actor, mediaID)

	assert.NoError(t, err)
	mockAssetRepo.AssertExpectations(t)
}
