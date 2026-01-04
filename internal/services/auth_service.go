package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUserInput captures the fields required to create a user account.
type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

// AuthService handles authentication and session lifecycle.
type AuthService interface {
	Register(ctx context.Context, req RegisterUserInput) (*dto.User, error)
	Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AccessToken, *dto.RefreshToken, error)
	RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.AccessToken, error)
	Logout(ctx context.Context, refreshTokenStr string) error
}

type authService struct {
	userRepo         repos.UserRepository
	refreshTokenRepo repos.RefreshTokenRepository
}

func NewAuthService(userRepo repos.UserRepository, refreshTokenRepo repos.RefreshTokenRepository) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterUserInput) (*dto.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	reqUser := &repos.RegisterUserData{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	return s.userRepo.Create(ctx, reqUser)
}

func (s *authService) Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AccessToken, *dto.RefreshToken, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, errors.New("invalid email or password")
	}

	accessTokenExp := time.Now().Add(15 * time.Minute) //TODO: make configurable
	accessToken, err := utils.GenerateJWT(jwtSecret, user.ID, "user", accessTokenExp)
	if err != nil {
		return nil, nil, errors.New("failed to generate access token")
	}

	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour) //TODO: make configurable
	refreshTokenStr, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, nil, errors.New("failed to generate refresh token")
	}
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenStr), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errors.New("failed to hash refresh token")
	}

	refreshToken, err := s.refreshTokenRepo.Create(ctx, user.ID, string(hashedRefreshToken), refreshTokenExp)
	if err != nil {
		return nil, nil, errors.New("failed to store refresh token")
	}

	formattedToken := refreshToken.ID.String() + ":" + refreshTokenStr

	return &dto.AccessToken{
			Token:     accessToken,
			ExpiresAt: accessTokenExp,
		}, &dto.RefreshToken{
			Token:     formattedToken,
			ExpiresAt: refreshTokenExp,
		}, nil
}

func (s *authService) RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.AccessToken, error) {
	// Parse selector:validator format
	parts := strings.Split(refreshTokenStr, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid refresh token format")
	}
	selector, validator := parts[0], parts[1]

	refreshToken, err := s.refreshTokenRepo.GetByID(ctx, selector)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(refreshToken.Token), []byte(validator)); err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if err = s.refreshTokenRepo.UpdateLastUsed(ctx, refreshToken.ID); err != nil {
		// best-effort update; do not fail request
		// or if failed then log out
	}

	accessTokenExp := time.Now().Add(15 * time.Minute)
	accessToken, err := utils.GenerateJWT(jwtSecret, refreshToken.UserID, "user", accessTokenExp)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	return &dto.AccessToken{
		Token:     accessToken,
		ExpiresAt: accessTokenExp,
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshTokenStr string) error {
	// Parse selector:validator format
	parts := strings.Split(refreshTokenStr, ":")
	if len(parts) != 2 {
		return errors.New("invalid refresh token format")
	}
	selector, validator := parts[0], parts[1]

	// Get token by selector (ID)
	refreshToken, err := s.refreshTokenRepo.GetByID(ctx, selector)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	// Validate the validator against the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(refreshToken.Token), []byte(validator)); err != nil {
		return errors.New("invalid refresh token")
	}

	return s.refreshTokenRepo.RevokeToken(ctx, refreshToken.ID)
}
