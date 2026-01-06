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
    Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AuthResponse, error)
    RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.RefreshTokenResponse, error)
    Logout(ctx context.Context, refreshTokenStr string) error
}

type authService struct {
    userRepo          repos.UserRepository
    refreshTokenRepo  repos.RefreshTokenRepository
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

func (s *authService) Login(ctx context.Context, req dto.LoginUserRequest, jwtSecret []byte) (*dto.AuthResponse, error) {
    user, err := s.userRepo.GetByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid email or password")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return nil, errors.New("invalid email or password")
    }

    accessTokenExp := time.Now().Add(15 * time.Minute)
    accessToken, err := utils.GenerateJWT(jwtSecret, user.ID, "user", accessTokenExp)
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    refreshTokenExp := time.Now().Add(7 * 24 * time.Hour)
    refreshTokenStr, err := generateRefreshToken()
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    if _, err = s.refreshTokenRepo.Create(ctx, user.ID, refreshTokenStr, refreshTokenExp); err != nil {
        return nil, errors.New("failed to store refresh token")
    }

    return &dto.AuthResponse{
        AccessToken:          accessToken,
        AccessTokenExpiresAt: accessTokenExp,
        User:                 user,
    }, nil
}

func (s *authService) RefreshAccessToken(ctx context.Context, refreshTokenStr string, jwtSecret []byte) (*dto.RefreshTokenResponse, error) {
    refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
    if err != nil {
        return nil, errors.New("invalid or expired refresh token")
    }

    if err = s.refreshTokenRepo.UpdateLastUsed(ctx, refreshToken.ID); err != nil {
        // best-effort update; do not fail request
    }

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

func (s *authService) Logout(ctx context.Context, refreshTokenStr string) error {
    refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, refreshTokenStr)
    if err != nil {
        return errors.New("invalid refresh token")
    }

    return s.refreshTokenRepo.RevokeToken(ctx, refreshToken.ID)
}

func generateRefreshToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}
