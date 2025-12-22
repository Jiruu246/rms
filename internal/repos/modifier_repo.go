package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/modifier"
	"github.com/google/uuid"
)

type ModifierRepository interface {
	Create(ctx context.Context, data *dto.CreateModifierData) (*dto.ModifierResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierResponse, error)
	Update(ctx context.Context, data *dto.UpdateModifierData) (*dto.ModifierResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*dto.ModifierResponse, error)
}

type modifierRepository struct {
	client *ent.Client
}

func NewEntModifierRepository(client *ent.Client) ModifierRepository {
	return &modifierRepository{
		client: client,
	}
}

func (r *modifierRepository) Create(ctx context.Context, data *dto.CreateModifierData) (*dto.ModifierResponse, error) {
	create, err := r.client.Modifier.Create().
		SetName(data.Request.Name).
		SetRequired(data.Request.Required).
		SetMultiSelect(data.Request.MultiSelect).
		SetMax(data.Request.Max).
		SetRestaurantID(data.Request.RestaurantID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create modifier: %w", err)
	}
	return mapToModifierResponse(create), nil
}

func (r *modifierRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierResponse, error) {
	m, err := r.client.Modifier.Query().
		Where(modifier.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("modifier not found: %w", err)
	}
	return mapToModifierResponse(m), nil
}

func (r *modifierRepository) Update(ctx context.Context, data *dto.UpdateModifierData) (*dto.ModifierResponse, error) {
	update := r.client.Modifier.UpdateOneID(data.ID)
	if data.Request.Name != nil {
		update.SetName(*data.Request.Name)
	}
	if data.Request.Required != nil {
		update.SetRequired(*data.Request.Required)
	}
	if data.Request.MultiSelect != nil {
		update.SetMultiSelect(*data.Request.MultiSelect)
	}
	if data.Request.Max != nil {
		update.SetMax(*data.Request.Max)
	}
	if data.Request.RestaurantID != nil {
		update.SetRestaurantID(*data.Request.RestaurantID)
	}
	m, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update modifier: %w", err)
	}
	return mapToModifierResponse(m), nil
}

func (r *modifierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.Modifier.DeleteOneID(id).Exec(ctx)
}

func (r *modifierRepository) GetAll(ctx context.Context) ([]*dto.ModifierResponse, error) {
	modifiers, err := r.client.Modifier.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifiers: %w", err)
	}
	responses := make([]*dto.ModifierResponse, 0, len(modifiers))
	for _, m := range modifiers {
		responses = append(responses, mapToModifierResponse(m))
	}
	return responses, nil
}

func mapToModifierResponse(m *ent.Modifier) *dto.ModifierResponse {
	return &dto.ModifierResponse{
		ID:           m.ID,
		Name:         m.Name,
		Required:     m.Required,
		MultiSelect:  m.MultiSelect,
		Max:          m.Max,
		RestaurantID: m.RestaurantID,
	}
}
