package dto

import (
	"github.com/google/uuid"
)

// CreateRestaurantRequest represents the request body for creating a restaurant
type CreateRestaurantRequest struct {
	Name           string         `json:"name" validate:"required,min=1,max=255" binding:"required"`
	Description    string         `json:"description" validate:"max=1000"`
	Phone          string         `json:"phone" validate:"required" binding:"required"`
	Email          string         `json:"email" validate:"required,email" binding:"required"`
	Address        string         `json:"address" validate:"required" binding:"required"`
	City           string         `json:"city" validate:"required" binding:"required"`
	State          string         `json:"state" validate:"required" binding:"required"`
	ZipCode        string         `json:"zip_code" validate:"required" binding:"required"`
	Country        string         `json:"country" validate:"required" binding:"required"`
	LogoURL        string         `json:"logo_url" validate:"omitempty,url"`
	CoverImageURL  string         `json:"cover_image_url" validate:"omitempty,url"`
	Status         string         `json:"status" validate:"omitempty,oneof=active inactive closed"`
	OperatingHours map[string]any `json:"operating_hours"`
	Currency       string         `json:"currency" validate:"required" binding:"required"`
}

type CreateRestaurantData struct {
	Request *CreateRestaurantRequest
	UserID  uuid.UUID
}

// UpdateRestaurantRequest represents the request body for updating a restaurant
// Uses pointers to distinguish between omitted values (nil) and deliberately empty/zero values
type UpdateRestaurantRequest struct {
	Name           *string         `json:"name" validate:"omitempty,min=1,max=255"`
	Description    *string         `json:"description" validate:"omitempty,max=1000"`
	Phone          *string         `json:"phone" validate:"omitempty"`
	Email          *string         `json:"email" validate:"omitempty,email"`
	Address        *string         `json:"address" validate:"omitempty"`
	City           *string         `json:"city" validate:"omitempty"`
	State          *string         `json:"state" validate:"omitempty"`
	ZipCode        *string         `json:"zip_code" validate:"omitempty"`
	Country        *string         `json:"country" validate:"omitempty"`
	LogoURL        *string         `json:"logo_url" validate:"omitempty,url"`
	CoverImageURL  *string         `json:"cover_image_url" validate:"omitempty,url"`
	Status         *string         `json:"status" validate:"omitempty,oneof=active inactive closed"`
	OperatingHours *map[string]any `json:"operating_hours"`
	Currency       *string         `json:"currency" validate:"omitempty"`
}

type UpdateRestaurantData struct {
	Request *UpdateRestaurantRequest
	ID      uuid.UUID
}

// RestaurantResponse represents the response structure for restaurant data
type RestaurantResponse struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Phone          string         `json:"phone"`
	Email          string         `json:"email"`
	Address        string         `json:"address"`
	City           string         `json:"city"`
	State          string         `json:"state"`
	ZipCode        string         `json:"zip_code"`
	Country        string         `json:"country"`
	LogoURL        string         `json:"logo_url"`
	CoverImageURL  string         `json:"cover_image_url"`
	Status         string         `json:"status"`
	OperatingHours map[string]any `json:"operating_hours"`
	Currency       string         `json:"currency"`
}
