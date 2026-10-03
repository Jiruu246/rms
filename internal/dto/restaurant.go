package dto

import (
	"github.com/google/uuid"
)

type CreateRestaurantRequest struct {
	Name           string         `json:"name" validate:"required,min=1,max=255"`
	Description    string         `json:"description" validate:"max=1000"`
	Phone          string         `json:"phone"`
	Email          string         `json:"email" validate:"omitempty,email"`
	Address        string         `json:"address"`
	City           string         `json:"city"`
	State          string         `json:"state"`
	ZipCode        string         `json:"zip_code"`
	Country        string         `json:"country"`
	Status         string         `json:"status" validate:"omitempty,oneof=active inactive closed"`
	OperatingHours map[string]any `json:"operating_hours"`
	Currency       string         `json:"currency"`
}

type CreateRestaurantData struct {
	Request *CreateRestaurantRequest
	UserID  uuid.UUID
}

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
	Status         *string         `json:"status" validate:"omitempty,oneof=active inactive closed"`
	OperatingHours *map[string]any `json:"operating_hours"`
	Currency       *string         `json:"currency" validate:"omitempty"`
}

type UpdateRestaurantData struct {
	Request *UpdateRestaurantRequest
	ID      uuid.UUID
}

type RestaurantImageSlot string

const (
	RestaurantImageSlotLogo  RestaurantImageSlot = "logo"
	RestaurantImageSlotCover RestaurantImageSlot = "cover"
)

func ParseRestaurantImageSlot(s string) (slot RestaurantImageSlot, ok bool) {
	switch RestaurantImageSlot(s) {
	case RestaurantImageSlotLogo, RestaurantImageSlotCover:
		return RestaurantImageSlot(s), true
	default:
		return "", false
	}
}

type UpdateRestaurantImageRequest struct {
	UploadID uuid.UUID `json:"upload_id" validate:"required"`
}

type SetRestaurantImageData struct {
	RestaurantID uuid.UUID
	Slot         RestaurantImageSlot
	MediaAssetID uuid.UUID
}

type ClearRestaurantImageData struct {
	RestaurantID uuid.UUID
	Slot         RestaurantImageSlot
}

type Restaurant struct {
	ID                     uuid.UUID      `json:"id"`
	Name                   string         `json:"name"`
	Description            string         `json:"description"`
	Phone                  string         `json:"phone"`
	Email                  string         `json:"email"`
	Address                string         `json:"address"`
	City                   string         `json:"city"`
	State                  string         `json:"state"`
	ZipCode                string         `json:"zip_code"`
	Country                string         `json:"country"`
	LogoMediaAssetID       *uuid.UUID     `json:"-"`
	CoverImageMediaAssetID *uuid.UUID     `json:"-"`
	LogoURL                string         `json:"logo_url,omitempty"`
	CoverImageURL          string         `json:"cover_image_url,omitempty"`
	Status                 string         `json:"status"`
	OperatingHours         map[string]any `json:"operating_hours"`
	Currency               string         `json:"currency"`
}
