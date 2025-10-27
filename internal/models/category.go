package models

import "time"

type Category struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" validate:"required"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255" binding:"required"`
	Description  string `json:"description" validate:"max=1000"`
	DisplayOrder int    `json:"display_order" validate:"min=0"`
	IsActive     bool   `json:"is_active"`
}

// UpdateCategoryRequest represents the request body for updating a category
// Uses pointers to distinguish between omitted values (nil) and deliberately empty/zero values
type UpdateCategoryRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description  *string `json:"description" validate:"omitempty,max=1000"`
	DisplayOrder *int    `json:"display_order" validate:"omitempty,min=0"`
	IsActive     *bool   `json:"is_active" validate:"omitempty"`
}
