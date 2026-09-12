package services

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

const (
	ActionReadRestaurant   authz.Action = "restaurant:read"
	ActionUpdateRestaurant authz.Action = "restaurant:update"
	ActionDeleteRestaurant authz.Action = "restaurant:delete"
)

var restaurantImageSlotPurposes = map[dto.RestaurantImageSlot]mediaupload.Purpose{
	dto.RestaurantImageSlotLogo:  mediaupload.PurposeRestaurantLogo,
	dto.RestaurantImageSlotCover: mediaupload.PurposeRestaurantCoverImage,
}

func purposeForSlot(slot dto.RestaurantImageSlot) (mediaupload.Purpose, error) {
	purpose, ok := restaurantImageSlotPurposes[slot]
	if !ok {
		return "", fmt.Errorf("restaurant service: no purpose registered for image slot %q", slot)
	}
	return purpose, nil
}

type RestaurantService interface {
	Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.Restaurant, error)
	GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.Restaurant, error)
	GetAll(ctx context.Context, actor authz.Actor) ([]*dto.Restaurant, error)
	Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.Restaurant, error)
	Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error
	CreateImageUpload(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot) (*dto.CreateUploadResult, error)
	UpdateImage(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot, uploadID uuid.UUID) (*dto.Restaurant, error)
	ClearImage(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot) (*dto.Restaurant, error)
	AuthorizeOwnership(ctx context.Context, actor authz.Actor, action authz.Action, restaurantID uuid.UUID) error
}

type restaurantService struct {
	repo         repos.RestaurantRepository
	mediaService MediaService
	authorizer   authz.Authorizer
	transactor   repos.Transactor
}

func NewRestaurantService(transactor repos.Transactor, repo repos.RestaurantRepository, mediaService MediaService) RestaurantService {
	return &restaurantService{
		repo:         repo,
		mediaService: mediaService,
		authorizer:   authz.NewPolicyAuthorizer(),
		transactor:   transactor,
	}
}

func (s *restaurantService) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.Restaurant, error) {
	return s.repo.Create(ctx, data)
}

func (s *restaurantService) GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.Restaurant, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionReadRestaurant, id); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *restaurantService) GetAll(ctx context.Context, actor authz.Actor) ([]*dto.Restaurant, error) {
	return s.repo.GetAllForUser(ctx, actor.UserID)
}

func (s *restaurantService) Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.Restaurant, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionUpdateRestaurant, id); err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, &dto.UpdateRestaurantData{Request: req, ID: id})
}

func (s *restaurantService) CreateImageUpload(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot) (*dto.CreateUploadResult, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionUpdateRestaurant, id); err != nil {
		return nil, err
	}

	purpose, err := purposeForSlot(slot)
	if err != nil {
		return nil, err
	}

	return s.mediaService.CreateUpload(ctx, actor, purpose)
}

func (s *restaurantService) UpdateImage(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot, uploadID uuid.UUID) (*dto.Restaurant, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionUpdateRestaurant, id); err != nil {
		return nil, err
	}

	purpose, err := purposeForSlot(slot)
	if err != nil {
		return nil, err
	}

	return repos.WithinTxResult(ctx, s.transactor, func(ctx context.Context) (*dto.Restaurant, error) {
		asset, err := s.mediaService.ConsumeUpload(ctx, actor, uploadID, purpose)
		if err != nil {
			return nil, err
		}

		// Whatever asset (if any) was previously in this slot is deliberately
		// left completely alone from here on — not soft-deleted, not touched
		// in storage. Once SetImage below repoints the column at the new
		// asset, that old asset becomes an orphan: an active MediaAsset row
		// (and its R2 object) nothing references anymore. Reconciling
		// orphaned assets is deferred to a later iteration — see
		// documentation/MediaUploadFramework.md.
		return s.repo.SetImage(ctx, &dto.SetRestaurantImageData{
			RestaurantID: id,
			Slot:         slot,
			MediaAssetID: asset.ID,
		})
	})
}

func (s *restaurantService) ClearImage(ctx context.Context, actor authz.Actor, id uuid.UUID, slot dto.RestaurantImageSlot) (*dto.Restaurant, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionUpdateRestaurant, id); err != nil {
		return nil, err
	}

	// TODO: implement outbox pattern to asynchronously delete the orphaned asset after clearing the image slot.
	return s.repo.ClearImage(ctx, &dto.ClearRestaurantImageData{
		RestaurantID: id,
		Slot:         slot,
	})
}

func (s *restaurantService) Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error {
	if err := s.AuthorizeOwnership(ctx, actor, ActionDeleteRestaurant, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *restaurantService) AuthorizeOwnership(ctx context.Context, actor authz.Actor, action authz.Action, restaurantID uuid.UUID) error {
	resource, err := s.repo.GetAuthorizationResource(ctx, restaurantID)
	if err != nil {
		return err
	}

	_, err = s.authorizer.Authorize(ctx, authz.Request{
		Actor:    actor,
		Action:   action,
		Resource: resource,
	})
	return err
}
