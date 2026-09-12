package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
)

type RestaurantRepository interface {
	Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Restaurant, error)
	Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.Restaurant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAllForUser(ctx context.Context, userID uuid.UUID) ([]*dto.Restaurant, error)
	GetAuthorizationResource(ctx context.Context, id uuid.UUID) (authz.Resource, error)
	SetImage(ctx context.Context, data *dto.SetRestaurantImageData) (*dto.Restaurant, error)
	ClearImage(ctx context.Context, data *dto.ClearRestaurantImageData) (*dto.Restaurant, error)
}

type restaurantRepository struct {
	client *ent.Client
	// mediaPublicBaseURL builds the read-only display URL for an attached
	// logo/cover MediaAsset from its storage key (see
	// dto.RestaurantResponse). Empty means objects aren't served publicly —
	// LogoURL/CoverImageURL are then left blank rather than guessed at.
	mediaPublicBaseURL string
}

// NewEntRestaurantRepository creates a new Ent-based restaurant repository
func NewEntRestaurantRepository(client *ent.Client, mediaPublicBaseURL string) RestaurantRepository {
	return &restaurantRepository{
		client:             client,
		mediaPublicBaseURL: mediaPublicBaseURL,
	}
}

func (r *restaurantRepository) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	create, err := c.Restaurant.Create().
		SetName(data.Request.Name).
		SetDescription(data.Request.Description).
		SetPhone(data.Request.Phone).
		SetEmail(data.Request.Email).
		SetAddress(data.Request.Address).
		SetCity(data.Request.City).
		SetState(data.Request.State).
		SetZipCode(data.Request.ZipCode).
		SetCountry(data.Request.Country).
		SetStatus(restaurant.StatusActive).
		SetCurrency(data.Request.Currency).
		SetOperatingHours(data.Request.OperatingHours).
		SetOwnerID(data.UserID).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	return r.GetByID(ctx, create.ID)
}

// TODO This should not always join to the assets (e.g. for internal processing, we don't need the assets)
// We can use query options to control this behavior see menu_item_repo.go for an example
func (r *restaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	row, err := c.Restaurant.Query().
		Where(restaurant.IDEQ(id)).
		WithLogoAsset().
		WithCoverImageAsset().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("restaurant %s", id)
		}
		return nil, fmt.Errorf("failed to get restaurant: %w", err)
	}

	return r.mapToRestaurantResponse(row), nil
}

func (r *restaurantRepository) Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	update := c.Restaurant.UpdateOneID(data.ID)

	if data.Request.Name != nil {
		update.SetName(*data.Request.Name)
	}

	if data.Request.Description != nil {
		update.SetDescription(*data.Request.Description)
	}

	if data.Request.Phone != nil {
		update.SetPhone(*data.Request.Phone)
	}

	if data.Request.Email != nil {
		update.SetEmail(*data.Request.Email)
	}

	if data.Request.Address != nil {
		update.SetAddress(*data.Request.Address)
	}

	if data.Request.City != nil {
		update.SetCity(*data.Request.City)
	}

	if data.Request.State != nil {
		update.SetState(*data.Request.State)
	}

	if data.Request.ZipCode != nil {
		update.SetZipCode(*data.Request.ZipCode)
	}

	if data.Request.Country != nil {
		update.SetCountry(*data.Request.Country)
	}

	if data.Request.Status != nil {
		update.SetStatus(restaurant.Status(*data.Request.Status))
	}

	if data.Request.OperatingHours != nil {
		update.SetOperatingHours(*data.Request.OperatingHours)
	}

	if data.Request.Currency != nil {
		update.SetCurrency(*data.Request.Currency)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("restaurant %s", data.ID)
		}
		return nil, fmt.Errorf("failed to update restaurant: %w", err)
	}

	return r.GetByID(ctx, updated.ID)
}

// SetImage overwrites one image slot's column with the given asset. Whatever
// asset (if any) was previously in that slot is left completely alone here —
// callers decide separately whether/when that old asset is reconciled (see
// RestaurantService.UpdateImage).
func (r *restaurantRepository) SetImage(ctx context.Context, data *dto.SetRestaurantImageData) (*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	update := c.Restaurant.UpdateOneID(data.RestaurantID)

	switch data.Slot {
	case dto.RestaurantImageSlotLogo:
		update.SetLogoMediaAssetID(data.MediaAssetID)
	case dto.RestaurantImageSlotCover:
		update.SetCoverImageMediaAssetID(data.MediaAssetID)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("restaurant %s", data.RestaurantID)
		}
		return nil, fmt.Errorf("failed to set restaurant image: %w", err)
	}

	return r.GetByID(ctx, updated.ID)
}

// ClearImage detaches whatever asset is assigned to one image slot. Setting
// an already-nil column to nil again is not an error, so a repeated clear on
// an already-empty slot succeeds the same way.
func (r *restaurantRepository) ClearImage(ctx context.Context, data *dto.ClearRestaurantImageData) (*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	update := c.Restaurant.UpdateOneID(data.RestaurantID)

	switch data.Slot {
	case dto.RestaurantImageSlotLogo:
		update.ClearLogoMediaAssetID()
	case dto.RestaurantImageSlotCover:
		update.ClearCoverImageMediaAssetID()
	}

	updated, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("restaurant %s", data.RestaurantID)
		}
		return nil, fmt.Errorf("failed to clear restaurant image: %w", err)
	}

	return r.GetByID(ctx, updated.ID)
}

func (r *restaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	c := clientFromContext(ctx, r.client)
	err := c.Restaurant.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperr.NotFound("restaurant %s", id)
		}
		return fmt.Errorf("failed to delete restaurant: %w", err)
	}
	return nil
}

func (r *restaurantRepository) GetAuthorizationResource(ctx context.Context, id uuid.UUID) (authz.Resource, error) {
	c := clientFromContext(ctx, r.client)
	row, err := c.Restaurant.Query().
		Where(restaurant.IDEQ(id)).
		Select(restaurant.FieldID, restaurant.FieldOwnerID).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return authz.Resource{}, apperr.NotFound("restaurant %s", id)
		}
		return authz.Resource{}, fmt.Errorf("failed to get restaurant: %w", err)
	}

	return authz.Resource{
		Type:         "restaurant",
		ID:           row.ID,
		RestaurantID: row.ID,
		OwnerUserID:  row.OwnerID,
	}, nil
}

func (r *restaurantRepository) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]*dto.Restaurant, error) {
	c := clientFromContext(ctx, r.client)
	restaurants, err := c.Restaurant.Query().
		Where(restaurant.OwnerIDEQ(userID)).
		WithLogoAsset().
		WithCoverImageAsset().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get restaurants for user: %w", err)
	}

	var responses []*dto.Restaurant
	for _, res := range restaurants {
		responses = append(responses, r.mapToRestaurantResponse(res))
	}

	return responses, nil
}

// buildAssetURL derives a read-only display URL for a MediaAsset's storage
// key. Returns "" when no public base URL is configured, or when there is no
// asset to build one for — the storage key itself must never reach the
// client (see documentation/MediaUploadFramework.md).
func (r *restaurantRepository) buildAssetURL(asset *ent.MediaAsset) string {
	if asset == nil || r.mediaPublicBaseURL == "" {
		return ""
	}
	return strings.TrimSuffix(r.mediaPublicBaseURL, "/") + "/" + asset.StorageKey
}

func (r *restaurantRepository) mapToRestaurantResponse(restaurant *ent.Restaurant) *dto.Restaurant {
	resp := &dto.Restaurant{
		ID:                     restaurant.ID,
		Name:                   restaurant.Name,
		Description:            restaurant.Description,
		Phone:                  restaurant.Phone,
		Email:                  restaurant.Email,
		Address:                restaurant.Address,
		City:                   restaurant.City,
		State:                  restaurant.State,
		ZipCode:                restaurant.ZipCode,
		Country:                restaurant.Country,
		LogoMediaAssetID:       restaurant.LogoMediaAssetID,
		LogoURL:                r.buildAssetURL(restaurant.Edges.LogoAsset),
		CoverImageMediaAssetID: restaurant.CoverImageMediaAssetID,
		CoverImageURL:          r.buildAssetURL(restaurant.Edges.CoverImageAsset),
		Status:                 restaurant.Status.String(),
		OperatingHours:         restaurant.OperatingHours,
		Currency:               restaurant.Currency,
	}
	return resp
}
