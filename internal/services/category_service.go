package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/pagination"
	"github.com/google/uuid"
)

const (
	ActionCreateCategory authz.Action = "category:create"
	ActionReadCategory   authz.Action = "category:read"
	ActionUpdateCategory authz.Action = "category:update"
	ActionDeleteCategory authz.Action = "category:delete"
)

type CategoryService interface {
	Create(ctx context.Context, actor authz.Actor, req *dto.CreateCategoryRequest) (*dto.Category, error)
	GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.Category, error)
	List(ctx context.Context, actor authz.Actor, restaurantID uuid.UUID, req pagination.PageRequest) (*pagination.PageResponse[*dto.CategoryListItem], error)
	Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error)
	Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error
}

type categoryService struct {
	repo              repos.CategoryRepository
	restaurantService RestaurantService
	authorizer        authz.Authorizer
}

func NewCategoryService(repo repos.CategoryRepository, restaurantService RestaurantService) CategoryService {
	return &categoryService{
		repo:              repo,
		restaurantService: restaurantService,
		authorizer:        authz.NewPolicyAuthorizer(),
	}
}

// Create requires the actor to own the target restaurant — a category has no
// owner of its own before it exists, so the check resolves against the
// restaurant named by req.RestaurantID rather than the (not-yet-created) category.
func (s *categoryService) Create(ctx context.Context, actor authz.Actor, req *dto.CreateCategoryRequest) (*dto.Category, error) {
	if err := s.restaurantService.AuthorizeOwnership(ctx, actor, ActionCreateCategory, req.RestaurantID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, req)
}

func (s *categoryService) GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.Category, error) {
	if id == uuid.Nil {
		return nil, apperr.Invalid("invalid category id")
	}

	resource, err := s.authorize(ctx, actor, ActionReadCategory, id)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, resource.RestaurantID, id)
}

func (s *categoryService) List(ctx context.Context, actor authz.Actor, restaurantID uuid.UUID, req pagination.PageRequest) (*pagination.PageResponse[*dto.CategoryListItem], error) {
	if err := s.restaurantService.AuthorizeOwnership(ctx, actor, ActionReadCategory, restaurantID); err != nil {
		return nil, err
	}

	return s.repo.List(ctx, restaurantID, req)
}

func (s *categoryService) Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	if id == uuid.Nil {
		return nil, apperr.Invalid("invalid category id")
	}

	resource, err := s.authorize(ctx, actor, ActionUpdateCategory, id)
	if err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, resource.RestaurantID, id, req)
}

func (s *categoryService) Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error {
	if id == uuid.Nil {
		return apperr.Invalid("invalid category id")
	}

	resource, err := s.authorize(ctx, actor, ActionDeleteCategory, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, resource.RestaurantID, id)
}

// authorize resolves the category's authz.Resource, checks the actor may perform
// action on it, and returns the resource so callers can scope the follow-up repo
// query by RestaurantID (defense in depth — see documentation/AuthzFramework.md).
func (s *categoryService) authorize(ctx context.Context, actor authz.Actor, action authz.Action, id uuid.UUID) (authz.Resource, error) {
	resource, err := s.repo.GetAuthorizationResource(ctx, id)
	if err != nil {
		return authz.Resource{}, err
	}

	if _, err := s.authorizer.Authorize(ctx, authz.Request{
		Actor:    actor,
		Action:   action,
		Resource: resource,
	}); err != nil {
		return authz.Resource{}, err
	}
	return resource, nil
}
