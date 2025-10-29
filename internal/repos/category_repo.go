package repos

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/category"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *ent.Category) (*ent.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error)
	Update(ctx context.Context, category *ent.Category) (*ent.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*ent.Category, error)
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
func (r *categoryRepository) Create(ctx context.Context, cat *ent.Category) (*ent.Category, error) {
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

	return created, nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error) {
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

	return cat, nil
}

// GetAll retrieves all categories
func (r *categoryRepository) GetAll(ctx context.Context) ([]*ent.Category, error) {
	categories, err := r.client.Category.
		Query().
		Order(category.ByDisplayOrder(), category.ByName()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// Update updates an existing category
func (r *categoryRepository) Update(ctx context.Context, cat *ent.Category) (*ent.Category, error) {
	updated, err := r.client.Category.
		UpdateOneID(cat.ID).
		SetName(cat.Name).
		SetDescription(cat.Description).
		SetDisplayOrder(cat.DisplayOrder).
		SetIsActive(cat.IsActive).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return updated, nil
}

// Delete deletes a category by ID
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
