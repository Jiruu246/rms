package dto

import (
	"github.com/google/uuid"
)

type CreateModifierRequest struct {
	Name         string    `json:"name" validate:"required,min=1,max=255" binding:"required"`
	Required     bool      `json:"required"`
	MultiSelect  bool      `json:"multi_select"`
	Max          int       `json:"max" validate:"min=0"`
	RestaurantID uuid.UUID `json:"restaurant_id" validate:"required" binding:"required"`
}

type CreateModifierData struct {
	Request *CreateModifierRequest
}

type UpdateModifierRequest struct {
	Name         *string    `json:"name" validate:"omitempty,min=1,max=255"`
	Required     *bool      `json:"required"`
	MultiSelect  *bool      `json:"multi_select"`
	Max          *int       `json:"max" validate:"omitempty,min=1"`
	RestaurantID *uuid.UUID `json:"restaurant_id"`
}

type UpdateModifierData struct {
	Request *UpdateModifierRequest
	ID      uuid.UUID
}

type Modifier struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Required     bool      `json:"required"`
	MultiSelect  bool      `json:"multi_select"`
	Max          int       `json:"max"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
}
