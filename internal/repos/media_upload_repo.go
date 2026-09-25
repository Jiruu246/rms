package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/google/uuid"
)

type MediaUploadRepository interface {
	Create(ctx context.Context, params CreateMediaUploadParams) (*ent.MediaUpload, error)
	GetByID(ctx context.Context, ownerID, id uuid.UUID) (*ent.MediaUpload, error)
	Consume(ctx context.Context, ownerID, id uuid.UUID, params ConsumeMediaUploadParams) (*ent.MediaAsset, error)
	Fail(ctx context.Context, ownerID, id uuid.UUID) error
}

type CreateMediaUploadParams struct {
	OwnerID   uuid.UUID
	Purpose   mediaupload.Purpose
	ObjectKey string
	ExpiresAt time.Time
}

type ConsumeMediaUploadParams struct {
	ContentType string
	SizeBytes   int64
}

type mediaUploadRepository struct {
	client *ent.Client
}

func NewEntMediaUploadRepository(client *ent.Client) MediaUploadRepository {
	return &mediaUploadRepository{client: client}
}

func (r *mediaUploadRepository) Create(ctx context.Context, params CreateMediaUploadParams) (*ent.MediaUpload, error) {
	created, err := r.client.MediaUpload.
		Create().
		SetOwnerID(params.OwnerID).
		SetPurpose(params.Purpose).
		SetObjectKey(params.ObjectKey).
		SetExpiresAt(params.ExpiresAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperr.Conflict("object key %q is already in use", params.ObjectKey)
		}
		return nil, fmt.Errorf("failed to create media upload: %w", err)
	}

	return created, nil
}

func (r *mediaUploadRepository) GetByID(ctx context.Context, ownerID, id uuid.UUID) (*ent.MediaUpload, error) {
	upload, err := r.client.MediaUpload.
		Query().
		Where(mediaupload.ID(id), mediaupload.OwnerIDEQ(ownerID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("media upload %s", id)
		}
		return nil, fmt.Errorf("failed to get media upload: %w", err)
	}

	return upload, nil
}

func (r *mediaUploadRepository) Consume(ctx context.Context, ownerID, id uuid.UUID, params ConsumeMediaUploadParams) (*ent.MediaAsset, error) {
	c := clientFromContext(ctx, r.client)

	upload, err := c.MediaUpload.
		Query().
		// This is strictly limited to the owner of the upload request
		Where(mediaupload.ID(id), mediaupload.OwnerIDEQ(ownerID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("media upload %s", id)
		}
		return nil, fmt.Errorf("failed to get media upload: %w", err)
	}

	if upload.Status != mediaupload.StatusIssued || !upload.ExpiresAt.After(time.Now()) {
		return nil, apperr.Conflict("media upload %s is not issued or has expired", id)
	}

	updated, err := c.MediaUpload.
		UpdateOneID(id).
		Where(mediaupload.StatusEQ(mediaupload.StatusIssued), mediaupload.ExpiresAtGT(time.Now())).
		SetStatus(mediaupload.StatusConsumed).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.Conflict("media upload %s is not issued or has expired", id)
		}
		return nil, fmt.Errorf("failed to consume media upload: %w", err)
	}

	created, err := c.MediaAsset.
		Create().
		SetUploadedByUserID(ownerID).
		SetUploadID(updated.ID).
		SetStorageKey(updated.ObjectKey).
		SetContentType(params.ContentType).
		SetSizeBytes(params.SizeBytes).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperr.Conflict("media upload %s has already been consumed", id)
		}
		return nil, fmt.Errorf("failed to create media asset: %w", err)
	}

	return created, nil
}

// Fail terminally marks an issued upload session as failed — used when the
// uploaded object was rejected at consume time (e.g. a content-type/size
// mismatch) and the caller has already cleaned up the corresponding storage
// object. Once failed, a session can never be consumed; the caller must
// request a fresh upload via Create instead of retrying this one.
//
// The conditional Where(status=issued) guards the same race Consume guards
// against: if another call already transitioned this session (e.g. a
// concurrent successful Consume), this affects zero rows and reports a
// conflict instead of clobbering that outcome.
func (r *mediaUploadRepository) Fail(ctx context.Context, ownerID, id uuid.UUID) error {
	_, err := r.client.MediaUpload.
		Query().
		Where(mediaupload.ID(id), mediaupload.OwnerIDEQ(ownerID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperr.NotFound("media upload %s", id)
		}
		return fmt.Errorf("failed to get media upload: %w", err)
	}

	_, err = r.client.MediaUpload.
		UpdateOneID(id).
		Where(mediaupload.StatusEQ(mediaupload.StatusIssued)).
		SetStatus(mediaupload.StatusFailed).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperr.Conflict("media upload %s is not issued", id)
		}
		return fmt.Errorf("failed to mark media upload failed: %w", err)
	}

	return nil
}
