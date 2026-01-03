package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type ModifierOptionService interface {
	Create(ctx context.Context, data *dto.CreateModifierOptionData) (*dto.ModifierOption, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierOption, error)
	GetAll(ctx context.Context) ([]*dto.ModifierOption, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateModifierOptionRequest) (*dto.ModifierOption, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type modifierOptionService struct {
	repo repos.ModifierOptionRepository
}

func NewModifierOptionService(repo repos.ModifierOptionRepository) ModifierOptionService {
	return &modifierOptionService{
		repo: repo,
	}
}

func (s *modifierOptionService) Create(ctx context.Context, data *dto.CreateModifierOptionData) (*dto.ModifierOption, error) {
	return s.repo.Create(ctx, data)
}

func (s *modifierOptionService) GetByID(ctx context.Context, id uuid.UUID) (*dto.ModifierOption, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *modifierOptionService) GetAll(ctx context.Context) ([]*dto.ModifierOption, error) {
	return s.repo.GetAll(ctx)
}

func (s *modifierOptionService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateModifierOptionRequest) (*dto.ModifierOption, error) {
	return s.repo.Update(ctx, &dto.UpdateModifierOptionData{
		Request: req,
		ID:      id,
	})
}

func (s *modifierOptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
