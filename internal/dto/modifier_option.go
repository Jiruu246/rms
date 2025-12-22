package dto

import (
	"github.com/google/uuid"
)

type CreateModifierOptionRequest struct {
	Name       string    `json:"name" validate:"required,min=1,max=255" binding:"required"`
	Price      float64   `json:"price" validate:"min=0"`
	ImageURL   string    `json:"image_url"`
	Available  bool      `json:"available"`
	PreSelect  bool      `json:"pre_select"`
	ModifierID uuid.UUID `json:"modifier_id" validate:"required" binding:"required"`
}

type CreateModifierOptionData struct {
	Request *CreateModifierOptionRequest
}

type UpdateModifierOptionRequest struct {
	Name       *string    `json:"name" validate:"omitempty,min=1,max=255"`
	Price      *float64   `json:"price"`
	ImageURL   *string    `json:"image_url"`
	Available  *bool      `json:"available"`
	PreSelect  *bool      `json:"pre_select"`
	ModifierID *uuid.UUID `json:"modifier_id"`
}

type UpdateModifierOptionData struct {
	Request *UpdateModifierOptionRequest
	ID      uuid.UUID
}

type ModifierOptionResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Price      float64   `json:"price"`
	ImageURL   string    `json:"image_url"`
	Available  bool      `json:"available"`
	PreSelect  bool      `json:"pre_select"`
	ModifierID uuid.UUID `json:"modifier_id"`
}
