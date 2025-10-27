package repos

import (
	"database/sql"
	"fmt"

	"github.com/Jiruu246/rms/internal/models"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository interface {
	Create(category *models.Category) (*models.Category, error)
	GetByID(id uint) (*models.Category, error)
	Update(category *models.Category) (*models.Category, error)
	Delete(id uint) error
}

type categoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) Create(category *models.Category) (*models.Category, error) {
	query := `
		INSERT INTO categories (name, description, display_order, is_active, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, name, description, display_order, is_active, created_at
	`

	var created models.Category
	err := r.db.Get(&created, query,
		category.Name,
		category.Description,
		category.DisplayOrder,
		category.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &created, nil
}

func (r *categoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	query := `
		SELECT id, name, description, display_order, is_active, created_at
		FROM categories 
		WHERE id = $1
	`

	err := r.db.Get(&category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

func (r *categoryRepository) Update(category *models.Category) (*models.Category, error) {
	query := `
		UPDATE categories 
		SET name = $2, description = $3, display_order = $4, is_active = $5
		WHERE id = $1
		RETURNING id, name, description, display_order, is_active, created_at
	`

	var updated models.Category
	err := r.db.Get(&updated, query, category.ID, category.Name, category.Description, category.DisplayOrder, category.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category with id %d not found", category.ID)
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &updated, nil
}

func (r *categoryRepository) Delete(id uint) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with id %d not found", id)
	}

	return nil
}
