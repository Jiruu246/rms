package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type CategoryService interface {
	Create(ctx context.Context, req *dto.CreateCategoryRequest) (*ent.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error)
	GetAll(ctx context.Context) ([]*ent.Category, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*ent.Category, error)
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

func (s *categoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*ent.Category, error) {
	// Create Category from request
	category := &ent.Category{
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}

	return s.repo.Create(ctx, category)
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid category id")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*ent.Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid category id")
	}

	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	hasUpdates := false

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		category.Name = *req.Name
		hasUpdates = true
	}

	if req.Description != nil {
		// Description can be empty string - that's a valid value
		category.Description = *req.Description
		hasUpdates = true
	}

	if req.DisplayOrder != nil {
		// DisplayOrder can be 0 - that's a valid value
		category.DisplayOrder = *req.DisplayOrder
		hasUpdates = true
	}

	if req.IsActive != nil {
		// IsActive can be false - that's a valid value
		category.IsActive = *req.IsActive
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no valid fields provided for update")
	}

	return s.repo.Update(ctx, category)
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid category id")
	}

	return s.repo.Delete(ctx, id)
}

func (s *categoryService) GetAll(ctx context.Context) ([]*ent.Category, error) {
	return s.repo.GetAll(ctx)
}
