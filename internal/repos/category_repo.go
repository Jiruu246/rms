package repos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/category"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *dto.CreateCategoryRequest) (*dto.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	Update(ctx context.Context, id uuid.UUID, category *dto.UpdateCategoryRequest) (*dto.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*dto.Category, error)
}

type categoryRepository struct {
	client *ent.Client
}

// NewEntCategoryRepository creates a new Ent-based category repository
func NewEntCategoryRepository(client *ent.Client) CategoryRepository {
	return &categoryRepository{
		client: client,
	}
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, cat *dto.CreateCategoryRequest) (*dto.Category, error) {
	created, err := r.client.Category.
		Create().
		SetName(cat.Name).
		SetDescription(cat.Description).
		SetDisplayOrder(cat.DisplayOrder).
		SetIsActive(cat.IsActive).
		SetCreatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &dto.Category{
		ID:           created.ID,
		Name:         created.Name,
		Description:  created.Description,
		DisplayOrder: created.DisplayOrder,
		IsActive:     created.IsActive,
	}, nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	cat, err := r.client.Category.
		Query().
		Where(category.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &dto.Category{
		ID:           cat.ID,
		Name:         cat.Name,
		Description:  cat.Description,
		DisplayOrder: cat.DisplayOrder,
		IsActive:     cat.IsActive,
	}, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]*dto.Category, error) {
	categories, err := r.client.Category.
		Query().
		Order(category.ByDisplayOrder(), category.ByName()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	var dtoCategories []*dto.Category
	for _, cat := range categories {
		dtoCategories = append(dtoCategories, &dto.Category{
			ID:           cat.ID,
			Name:         cat.Name,
			Description:  cat.Description,
			DisplayOrder: cat.DisplayOrder,
			IsActive:     cat.IsActive,
		})
	}

	return dtoCategories, nil
}

func (r *categoryRepository) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	updateBuilder := r.client.Category.UpdateOneID(id)

	hasUpdates := false

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		updateBuilder.SetName(*req.Name)
		hasUpdates = true
	}

	if req.Description != nil {
		updateBuilder.SetDescription(*req.Description)
		hasUpdates = true
	}

	if req.DisplayOrder != nil {
		updateBuilder.SetDisplayOrder(*req.DisplayOrder)
		hasUpdates = true
	}

	if req.IsActive != nil {
		updateBuilder.SetIsActive(*req.IsActive)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no valid fields provided for update")
	}

	updatedCat, err := updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &dto.Category{
		ID:           updatedCat.ID,
		Name:         updatedCat.Name,
		Description:  updatedCat.Description,
		DisplayOrder: updatedCat.DisplayOrder,
		IsActive:     updatedCat.IsActive,
	}, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Category.
		DeleteOneID(id).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("category not found")
		}
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
