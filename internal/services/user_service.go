package services

import (
	"context"
	"errors"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*dto.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*dto.User, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
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
		return nil, errors.New("invalid user id")
	}

	return s.repo.Update(ctx, id, updates)
}

func (s *userService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
