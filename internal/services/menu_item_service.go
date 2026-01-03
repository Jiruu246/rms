package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
)

type MenuItemService interface {
	Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItem, error)
	GetAll(ctx context.Context) ([]*dto.MenuItem, error)
	GetByID(ctx context.Context, id int64) (*dto.MenuItem, error)
	Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItem, error)
	Delete(ctx context.Context, id int64) error
}

type menuItemService struct {
	repo repos.MenuItemRepository
}

func NewMenuItemService(repo repos.MenuItemRepository) MenuItemService {
	return &menuItemService{repo: repo}
}

func (s *menuItemService) Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItem, error) {
	return s.repo.Create(ctx, req)
}

func (s *menuItemService) GetAll(ctx context.Context) ([]*dto.MenuItem, error) {
	return s.repo.GetAll(ctx)
}

func (s *menuItemService) GetByID(ctx context.Context, id int64) (*dto.MenuItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *menuItemService) Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItem, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *menuItemService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
