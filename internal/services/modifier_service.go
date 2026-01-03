package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type ModifierService interface {
	Create(ctx context.Context, data *dto.CreateModifierData) (*dto.Modifier, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Modifier, error)
	GetAll(ctx context.Context) ([]*dto.Modifier, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateModifierRequest) (*dto.Modifier, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type modifierService struct {
	repo repos.ModifierRepository
}

func NewModifierService(repo repos.ModifierRepository) ModifierService {
	return &modifierService{
		repo: repo,
	}
}

func (s *modifierService) Create(ctx context.Context, data *dto.CreateModifierData) (*dto.Modifier, error) {
	return s.repo.Create(ctx, data)
}

func (s *modifierService) GetByID(ctx context.Context, id uuid.UUID) (*dto.Modifier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *modifierService) GetAll(ctx context.Context) ([]*dto.Modifier, error) {
	return s.repo.GetAll(ctx)
}

func (s *modifierService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateModifierRequest) (*dto.Modifier, error) {
	return s.repo.Update(ctx, &dto.UpdateModifierData{
		Request: req,
		ID:      id,
	})
}

func (s *modifierService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
