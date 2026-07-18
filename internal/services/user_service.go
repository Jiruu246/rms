package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*dto.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*dto.User, error)
}

type userService struct {
	repo repos.UserRepository
}

func NewUserService(repo repos.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*dto.User, error) {
	if id == uuid.Nil {
		return nil, apperr.Invalid("invalid user id")
	}

	return s.repo.Update(ctx, id, updates)
}
