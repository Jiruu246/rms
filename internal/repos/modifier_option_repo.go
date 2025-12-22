package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/modifieroption"
	"github.com/google/uuid"
)

type ModifierOptionRepository interface {
	Create(ctx context.Context, data *dto.CreateModifierOptionData) (*dto.ModifierOptionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierOptionResponse, error)
	Update(ctx context.Context, data *dto.UpdateModifierOptionData) (*dto.ModifierOptionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*dto.ModifierOptionResponse, error)
}

type modifierOptionRepository struct {
	client *ent.Client
}

func NewEntModifierOptionRepository(client *ent.Client) ModifierOptionRepository {
	return &modifierOptionRepository{
		client: client,
	}
}

func (r *modifierOptionRepository) Create(ctx context.Context, data *dto.CreateModifierOptionData) (*dto.ModifierOptionResponse, error) {
	create := r.client.ModifierOption.Create().
		SetName(data.Request.Name).
		SetPrice(data.Request.Price).
		SetImageURL(data.Request.ImageURL).
		SetAvailable(data.Request.Available).
		SetPreSelect(data.Request.PreSelect).
		SetModifierID(data.Request.ModifierID)
	m, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create modifier option: %w", err)
	}
	return mapToModifierOptionResponse(m), nil
}

func (r *modifierOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierOptionResponse, error) {
	m, err := r.client.ModifierOption.Query().
		Where(modifieroption.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("modifier option not found: %w", err)
	}
	return mapToModifierOptionResponse(m), nil
}

func (r *modifierOptionRepository) Update(ctx context.Context, data *dto.UpdateModifierOptionData) (*dto.ModifierOptionResponse, error) {
	update := r.client.ModifierOption.UpdateOneID(data.ID)
	if data.Request.Name != nil {
		update.SetName(*data.Request.Name)
	}
	if data.Request.Price != nil {
		update.SetPrice(*data.Request.Price)
	}
	if data.Request.ImageURL != nil {
		update.SetImageURL(*data.Request.ImageURL)
	}
	if data.Request.Available != nil {
		update.SetAvailable(*data.Request.Available)
	}
	if data.Request.PreSelect != nil {
		update.SetPreSelect(*data.Request.PreSelect)
	}
	if data.Request.ModifierID != nil {
		update.SetModifierID(*data.Request.ModifierID)
	}
	m, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update modifier option: %w", err)
	}
	return mapToModifierOptionResponse(m), nil
}

func (r *modifierOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.ModifierOption.DeleteOneID(id).Exec(ctx)
}

func (r *modifierOptionRepository) GetAll(ctx context.Context) ([]*dto.ModifierOptionResponse, error) {
	options, err := r.client.ModifierOption.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifier options: %w", err)
	}
	responses := make([]*dto.ModifierOptionResponse, 0, len(options))
	for _, m := range options {
		responses = append(responses, mapToModifierOptionResponse(m))
	}
	return responses, nil
}

func mapToModifierOptionResponse(m *ent.ModifierOption) *dto.ModifierOptionResponse {
	return &dto.ModifierOptionResponse{
		ID:         m.ID,
		Name:       m.Name,
		Price:      m.Price,
		ImageURL:   m.ImageURL,
		Available:  m.Available,
		PreSelect:  m.PreSelect,
		ModifierID: m.ModifierID,
	}
}
