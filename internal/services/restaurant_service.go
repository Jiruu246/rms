package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

const (
	ActionReadRestaurant   authz.Action = "restaurant:read"
	ActionUpdateRestaurant authz.Action = "restaurant:update"
	ActionDeleteRestaurant authz.Action = "restaurant:delete"
)

type RestaurantService interface {
	Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error)
	GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.RestaurantResponse, error)
	GetAll(ctx context.Context, actor authz.Actor) ([]*dto.RestaurantResponse, error)
	Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.RestaurantResponse, error)
	Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error
	// AuthorizeOwnership lets other entity services (e.g. category) check the
	// actor may act on a restaurant without duplicating the
	// GetAuthorizationResource + Authorizer.Authorize dance themselves.
	AuthorizeOwnership(ctx context.Context, actor authz.Actor, action authz.Action, restaurantID uuid.UUID) error
}

type restaurantService struct {
	repo       repos.RestaurantRepository
	authorizer authz.Authorizer
}

func NewRestaurantService(repo repos.RestaurantRepository) RestaurantService {
	return &restaurantService{
		repo:       repo,
		authorizer: authz.NewPolicyAuthorizer(),
	}
}

func (s *restaurantService) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error) {
	return s.repo.Create(ctx, data)
}

func (s *restaurantService) GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.RestaurantResponse, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionReadRestaurant, id); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *restaurantService) GetAll(ctx context.Context, actor authz.Actor) ([]*dto.RestaurantResponse, error) {
	return s.repo.GetAllForUser(ctx, actor.UserID)
}

func (s *restaurantService) Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.RestaurantResponse, error) {
	if err := s.AuthorizeOwnership(ctx, actor, ActionUpdateRestaurant, id); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, &dto.UpdateRestaurantData{
		Request: req,
		ID:      id,
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
