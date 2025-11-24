package services

import (
	"context"
	"errors"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.User, error)
	Login(ctx context.Context, req dto.LoginUserRequest) (*dto.User, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*dto.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*dto.User, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo repos.UserRepository
}

func NewUserService(repo repos.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	reqUser := &dto.RegisterUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	return s.repo.Create(ctx, reqUser)
}

func (s *userService) Login(ctx context.Context, req dto.LoginUserRequest) (*dto.User, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
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
