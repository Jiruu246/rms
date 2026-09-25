package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/storage"
)

// uploadConstraints is the per-purpose policy a requested upload is checked
// against. It is deliberately code, not config: the set of purposes is fixed
// and small, and each one maps to a specific place in the product (a menu
// item image, a restaurant logo, ...) that changing requires a code change
// anyway.
type uploadConstraints struct {
	AllowedContentTypes map[string]struct{}
	MaxSizeBytes        int64
}

func allowedContentTypes(types ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(types))
	for _, t := range types {
		set[t] = struct{}{}
	}
	return set
}

// toDTO exposes the constraints to the client so it can validate a file
// before upload instead of guessing at limits the server enforces anyway.
func (c uploadConstraints) toDTO() dto.UploadConstraints {
	types := make([]string, 0, len(c.AllowedContentTypes))
	for t := range c.AllowedContentTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return dto.UploadConstraints{
		MaxSizeBytes:        c.MaxSizeBytes,
		AllowedContentTypes: types,
	}
}

// purposeConstraints declares what each upload purpose accepts. SVG is
// deliberately excluded even for logos: it can embed scripts and this
// content is potentially served back to other users.
var purposeConstraints = map[mediaupload.Purpose]uploadConstraints{
	mediaupload.PurposeMenuItemImage: {
		AllowedContentTypes: allowedContentTypes("image/jpeg", "image/png", "image/webp"),
		MaxSizeBytes:        10 << 20, // 10 MiB
	},
	mediaupload.PurposeRestaurantLogo: {
		AllowedContentTypes: allowedContentTypes("image/jpeg", "image/png", "image/webp"),
		MaxSizeBytes:        5 << 20, // 5 MiB
	},
	mediaupload.PurposeRestaurantCoverImage: {
		AllowedContentTypes: allowedContentTypes("image/jpeg", "image/png", "image/webp"),
		MaxSizeBytes:        10 << 20, // 10 MiB
	},
}

type MediaService interface {
	CreateUpload(ctx context.Context, actor authz.Actor, purpose mediaupload.Purpose) (*dto.CreateUploadResult, error)
	ConsumeUpload(ctx context.Context, actor authz.Actor, uploadID uuid.UUID, expectedPurpose mediaupload.Purpose) (*dto.MediaAsset, error)
	DeleteMedia(ctx context.Context, actor authz.Actor, mediaID uuid.UUID) error
}

type mediaService struct {
	uploadRepo        repos.MediaUploadRepository
	assetRepo         repos.MediaAssetRepository
	provider          storage.StorageProvider
	uploadGrantExpiry time.Duration
	transactor        repos.Transactor
}

func NewMediaService(
	transactor repos.Transactor,
	uploadRepo repos.MediaUploadRepository,
	assetRepo repos.MediaAssetRepository,
	provider storage.StorageProvider,
	uploadGrantExpiry time.Duration,
) MediaService {
	return &mediaService{
		uploadRepo:        uploadRepo,
		assetRepo:         assetRepo,
		provider:          provider,
		uploadGrantExpiry: uploadGrantExpiry,
		transactor:        transactor,
	}
}

func (s *mediaService) CreateUpload(ctx context.Context, actor authz.Actor, purpose mediaupload.Purpose) (*dto.CreateUploadResult, error) {
	constraints, err := s.resolveConstraints(purpose)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	key := storage.ObjectKey(fmt.Sprintf("media/%04d/%02d/%s", now.Year(), now.Month(), uuid.New()))

	grantReq := storage.UploadRequest{
		Key:    key,
		Expiry: s.uploadGrantExpiry,
	}
	grant, err := s.provider.CreateUploadGrant(ctx, grantReq)
	if err != nil {
		return nil, err
	}

	upload, err := s.uploadRepo.Create(ctx, repos.CreateMediaUploadParams{
		OwnerID:   actor.UserID,
		Purpose:   purpose,
		ObjectKey: string(key),
		ExpiresAt: grant.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &dto.CreateUploadResult{
		UploadID:    upload.ID,
		ExpiresAt:   grant.ExpiresAt,
		Upload:      dto.NewUploadGrant(grant),
		Constraints: constraints.toDTO(),
	}, nil
}

func (s *mediaService) resolveConstraints(purpose mediaupload.Purpose) (uploadConstraints, error) {
	if err := mediaupload.PurposeValidator(purpose); err != nil {
		return uploadConstraints{}, apperr.Invalid("unknown upload purpose %q", purpose)
	}
	// purposeConstraints is total over every value PurposeValidator accepts,
	// so a lookup miss here would mean the two fell out of sync.
	constraints, ok := purposeConstraints[purpose]
	if !ok {
		return uploadConstraints{}, fmt.Errorf("media service: no upload constraints registered for purpose %q", purpose)
	}
	return constraints, nil
}

func (s *mediaService) ConsumeUpload(ctx context.Context, actor authz.Actor, uploadID uuid.UUID, expectedPurpose mediaupload.Purpose) (*dto.MediaAsset, error) {
	upload, err := s.uploadRepo.GetByID(ctx, actor.UserID, uploadID)
	if err != nil {
		return nil, err
	}

	if err := s.validateConsumedUpload(upload, expectedPurpose); err != nil {
		return nil, err
	}

	obj, err := s.provider.StatObject(ctx, storage.ObjectKey(upload.ObjectKey))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.Conflict("no object has been uploaded for media upload %s", uploadID)
		}
		return nil, err
	}
	if obj.Key != storage.ObjectKey(upload.ObjectKey) {
		return nil, fmt.Errorf("media service: storage provider returned object for key %q, expected %q", obj.Key, upload.ObjectKey)
	}

	if err := s.validateUploadedObject(upload, obj.Metadata); err != nil {
		s.rejectUpload(ctx, actor.UserID, uploadID, obj.Key)
		return nil, err
	}

	asset, err := repos.WithinTxResult(ctx, s.transactor, func(ctx context.Context) (*ent.MediaAsset, error) {
		return s.uploadRepo.Consume(ctx, actor.UserID, uploadID, repos.ConsumeMediaUploadParams{
			ContentType: obj.Metadata.ContentType,
			SizeBytes:   obj.Metadata.SizeBytes,
		})
	})
	if err != nil {
		return nil, err
	}

	return dto.NewMediaAsset(asset), nil
}

func (s *mediaService) validateConsumedUpload(upload *ent.MediaUpload, expectedPurpose mediaupload.Purpose) error {
	switch upload.Status {
	case mediaupload.StatusIssued:
		// good to go
	case mediaupload.StatusFailed:
		return apperr.Conflict("media upload %s was rejected and cannot be consumed; request a new upload", upload.ID)
	case mediaupload.StatusConsumed:
		return apperr.Conflict("media upload %s has already been consumed", upload.ID)
	default:
		return apperr.Invalid("media upload %s has unexpected status %q", upload.ID, upload.Status)
	}

	if !upload.ExpiresAt.After(time.Now()) {
		return apperr.Conflict("media upload %s has expired", upload.ID)
	}

	if expectedPurpose != "" && upload.Purpose != expectedPurpose {
		return apperr.Invalid("media upload %s was created for purpose %q, not %q", upload.ID, upload.Purpose, expectedPurpose)
	}

	return nil
}

func (s *mediaService) validateUploadedObject(upload *ent.MediaUpload, actual storage.ObjectMetadata) error {
	constraints := purposeConstraints[upload.Purpose]
	if actual.SizeBytes <= 0 || actual.SizeBytes > constraints.MaxSizeBytes {
		return apperr.Invalid("uploaded object size %d exceeds allowed %d bytes", actual.SizeBytes, constraints.MaxSizeBytes)
	}
	if _, ok := constraints.AllowedContentTypes[actual.ContentType]; !ok {
		return apperr.Invalid("uploaded object content type %q is not allowed for this upload purpose", actual.ContentType)
	}
	return nil
}

func (s *mediaService) rejectUpload(ctx context.Context, ownerID, uploadID uuid.UUID, key storage.ObjectKey) {
	if err := s.provider.DeleteObject(ctx, key); err != nil {
		//TODO: Using outbox pattern to delete the object
		log.Printf("media service: failed to delete rejected object %q for upload %s: %v", key, uploadID, err)
	}
	// This is use the same transaction as the upload consumption, so if it fails the upload will remain in "issued" state and can be retried. This is a risk,\
	//  but I think it's acceptable
	if err := s.uploadRepo.Fail(ctx, ownerID, uploadID); err != nil {
		// Similar with the above
		// Risk: if we failed then the upload session will remain in the "issued" state and can be retried
		// I think it's fine, we can just treat it as an app error and would require the user to retry
		log.Printf("media service: failed to mark upload %s failed: %v", uploadID, err)
	}
}

func (s *mediaService) DeleteMedia(ctx context.Context, _ authz.Actor, mediaID uuid.UUID) error {
	return s.assetRepo.Delete(ctx, mediaID)
}
