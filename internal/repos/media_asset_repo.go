package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/mediaasset"
	"github.com/google/uuid"
)

// MediaAssetRepository reads and soft-deletes finalized media assets. Assets
// are only ever created via MediaUploadRepository.Consume, which is the sole
// codepath that can produce one (see the unique upload_id constraint) — this
// interface has no Create for that reason.
type MediaAssetRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ent.MediaAsset, error)
	GetByUploadID(ctx context.Context, uploadID uuid.UUID) (*ent.MediaAsset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type mediaAssetRepository struct {
	client *ent.Client
}

func NewEntMediaAssetRepository(client *ent.Client) MediaAssetRepository {
	return &mediaAssetRepository{client: client}
}

func (r *mediaAssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.MediaAsset, error) {
	asset, err := r.client.MediaAsset.
		Query().
		Where(mediaasset.ID(id), mediaasset.StatusEQ(mediaasset.StatusActive)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("media asset %s", id)
		}
		return nil, fmt.Errorf("failed to get media asset: %w", err)
	}

	return asset, nil
}

func (r *mediaAssetRepository) GetByUploadID(ctx context.Context, uploadID uuid.UUID) (*ent.MediaAsset, error) {
	asset, err := r.client.MediaAsset.
		Query().
		Where(mediaasset.UploadIDEQ(uploadID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("media asset for upload %s", uploadID)
		}
		return nil, fmt.Errorf("failed to get media asset: %w", err)
	}

	return asset, nil
}

func (r *mediaAssetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.MediaAsset.
		UpdateOneID(id).
		Where(mediaasset.StatusEQ(mediaasset.StatusActive)).
		SetStatus(mediaasset.StatusDeleted).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperr.NotFound("media asset %s", id)
		}
		return fmt.Errorf("failed to delete media asset: %w", err)
	}

	return nil
}
