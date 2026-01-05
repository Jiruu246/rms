package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

type UserService interface {
	Register(ctx context.Context, req RegisterUserInput) (*dto.User, error)
	Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AuthResponse, error)
	RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshTokenStr string) error
	GetProfile(ctx context.Context, id uuid.UUID) (*dto.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates *dto.UpdateUserRequest) (*dto.User, error)
	DeleteAccount(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo             repos.UserRepository
	refreshTokenRepo repos.RefreshTokenRepository
}

func NewUserService(repo repos.UserRepository, refreshTokenRepo repos.RefreshTokenRepository) UserService {
	return &userService{
		repo:             repo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *userService) Register(ctx context.Context, req RegisterUserInput) (*dto.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	reqUser := &repos.RegisterUserData{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	return s.repo.Create(ctx, reqUser)
}

func (s *userService) Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate access token (expires in 15 minutes)
	accessTokenExp := time.Now().Add(15 * time.Minute)
	accessToken, err := utils.GenerateJWT(jwtSecret, user.ID, "user", accessTokenExp)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token (expires in 7 days)
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour)
	refreshTokenStr, err := generateRefreshToken()
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Store refresh token in database
	_, err = s.refreshTokenRepo.Create(ctx, user.ID, refreshTokenStr, refreshTokenExp)
	if err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &dto.AuthResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenExp,
		User:                 user,
	}, nil
}

func (s *userService) RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.RefreshTokenResponse, error) {
	// Validate refresh token
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Update last used timestamp
	err = s.refreshTokenRepo.UpdateLastUsed(ctx, refreshToken.ID)
	if err != nil {
		// Log error but don't fail the request
	}

	// Generate new access token
	accessTokenExp := time.Now().Add(15 * time.Minute)
	accessToken, err := utils.GenerateJWT(jwtSecret, refreshToken.UserID, "user", accessTokenExp)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &dto.RefreshTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenExp,
	}, nil
}

func (s *userService) Logout(ctx context.Context, refreshTokenStr string) error {
	// Get the refresh token
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	// Revoke the token
	return s.refreshTokenRepo.RevokeToken(ctx, refreshToken.ID)
}

// generateRefreshToken creates a cryptographically secure random token
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
