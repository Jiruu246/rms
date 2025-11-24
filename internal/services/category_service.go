package services

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type CategoryService interface {
	Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	GetAll(ctx context.Context) ([]*dto.Category, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	repo repos.CategoryRepository
}

func NewCategoryService(repo repos.CategoryRepository) CategoryService {
	return &categoryService{
		repo: repo,
	}
}

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.Category, error) {
	return s.repo.Create(ctx, req)
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid category id")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid category id")
	}

	return s.repo.Update(ctx, id, req)
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid category id")
	}

	return s.repo.Delete(ctx, id)
}

func (s *categoryService) GetAll(ctx context.Context) ([]*dto.Category, error) {
	return s.repo.GetAll(ctx)
}
