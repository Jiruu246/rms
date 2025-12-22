package dto

import (
	"github.com/google/uuid"
)

// TODO: put rquest struct into Handler layer
type CreateMenuItemRequest struct {
	Name         string    `json:"name" validate:"required" binding:"required"`
	Description  string    `json:"description"`
	Price        float64   `json:"price" validate:"required" binding:"required"`
	IsAvailable  bool      `json:"is_available"`
	RestaurantID uuid.UUID `json:"restaurant_id" validate:"required" binding:"required"`
	CategoryID   uuid.UUID `json:"category_id"`
}

type UpdateMenuItemRequest struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Price       *float64   `json:"price"`
	IsAvailable *bool      `json:"is_available"`
	CategoryID  *uuid.UUID `json:"category_id"`
}

type MenuItemResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	IsAvailable  bool      `json:"is_available"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	CategoryID   uuid.UUID `json:"category_id"`
}

// type MenuItemQueryParams struct {
// 	RestaurantID string
// 	CategoryID   string
// 	IsAvailable  string
// 	Search       string
// 	SortBy       string
// 	Order        string
// 	Page         string
// 	Limit        string
// }
