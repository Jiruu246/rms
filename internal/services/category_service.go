package services

import (
	"fmt"
	"strings"

	"github.com/Jiruu246/rms/internal/models"
	"github.com/Jiruu246/rms/internal/repos"
)

type CategoryService interface {
	Create(req *models.CreateCategoryRequest) (*models.Category, error)
	GetByID(id uint) (*models.Category, error)
	Update(id uint, req *models.UpdateCategoryRequest) (*models.Category, error)
	Delete(id uint) error
}

type categoryService struct {
	repo repos.CategoryRepository
}

func NewCategoryService(repo repos.CategoryRepository) CategoryService {
	return &categoryService{
		repo: repo,
	}
}

func (s *categoryService) Create(req *models.CreateCategoryRequest) (*models.Category, error) {
	// Create Category from request
	category := &models.Category{
		Name:         req.Name,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
	}

	return s.repo.Create(category)
}

func (s *categoryService) GetByID(id uint) (*models.Category, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	return s.repo.GetByID(id)
}

func (s *categoryService) Update(id uint, req *models.UpdateCategoryRequest) (*models.Category, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid category id")
	}

	category, err := s.repo.GetByID(id)
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

	return s.repo.Update(category)
}

func (s *categoryService) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid category id")
	}

	return s.repo.Delete(id)
}
