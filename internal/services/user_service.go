package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req dto.RegisterUserRequest) (*ent.Customer, error)
	Login(ctx context.Context, req dto.LoginUserRequest) (*ent.Customer, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*ent.Customer, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*ent.Customer, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo repos.UserRepository
}

func NewUserService(repo repos.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, req dto.RegisterUserRequest) (*ent.Customer, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &ent.Customer{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	return s.repo.Create(ctx, user)
}

func (s *userService) Login(ctx context.Context, req dto.LoginUserRequest) (*ent.Customer, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*ent.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*ent.Customer, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	hasUpdates := false

	if updates.Name != nil {
		if strings.TrimSpace(*updates.Name) == "" {
			return nil, errors.New("name cannot be empty")
		}
		hasUpdates = true
	}

	if updates.Email != nil {
		if strings.TrimSpace(*updates.Email) == "" {
			return nil, errors.New("email cannot be empty")
		}
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no valid fields, provided for update")
	}

	return s.repo.Update(ctx, user)
}

func (s *userService) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
