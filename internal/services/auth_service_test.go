package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockAuthUserRepository struct {
	mock.Mock
}

func (m *MockAuthUserRepository) Create(ctx context.Context, req *repos.RegisterUserData) (*dto.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockAuthUserRepository) GetByEmail(ctx context.Context, email string) (*dto.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockAuthUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockAuthUserRepository) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockAuthUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*ent.RefreshToken, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) GetByID(ctx context.Context, tokenID string) (*ent.RefreshToken, error) {
	args := m.Called(ctx, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) GetActiveTokensByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ent.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	testCases := []struct {
		name          string
		input         RegisterUserInput
		mockSetup     func(*MockAuthUserRepository)
		expectedError string
	}{
		{
			name: "successful registration",
			input: RegisterUserInput{
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockAuthUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(&dto.User{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			input: RegisterUserInput{
				Name:     "John Doe",
				Email:    "john2.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockAuthUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("repository error"))
			},
			expectedError: "repository error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockUserRepo := new(MockAuthUserRepository)
			mockRefreshRepo := new(MockRefreshTokenRepository)
			testCase.mockSetup(mockUserRepo)

			service := NewAuthService(mockUserRepo, mockRefreshRepo)
			result, err := service.Register(context.Background(), testCase.input)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.input.Name, result.Name)
				assert.Equal(t, testCase.input.Email, result.Email)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	userID := uuid.New()

	testCases := []struct {
		name          string
		req           dto.LoginUserRequest
		mockSetup     func(*MockAuthUserRepository, *MockRefreshTokenRepository)
		expectedError string
	}{
		{
			name: "successful login",
			req: dto.LoginUserRequest{
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockUserRepo *MockAuthUserRepository, mockRefreshRepo *MockRefreshTokenRepository) {
				mockUserRepo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(&dto.User{
					ID:       userID,
					Email:    "john.doe@example.com",
					Password: hashPassword("password123"),
				}, nil)
				mockRefreshRepo.On("Create", mock.Anything, userID, mock.Anything, mock.Anything).Return(&ent.RefreshToken{}, nil)
			},
			expectedError: "",
		},
		{
			name: "invalid email or password",
			req: dto.LoginUserRequest{
				Email:    "john.doe@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mockUserRepo *MockAuthUserRepository, mockRefreshRepo *MockRefreshTokenRepository) {
				mockUserRepo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(&dto.User{
					ID:       userID,
					Email:    "john.doe@example.com",
					Password: hashPassword("password123"),
				}, nil)
				// refresh token repo should not be called on failure
			},
			expectedError: "invalid email or password",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockUserRepo := new(MockAuthUserRepository)
			mockRefreshRepo := new(MockRefreshTokenRepository)
			testCase.mockSetup(mockUserRepo, mockRefreshRepo)

			service := NewAuthService(mockUserRepo, mockRefreshRepo)
			accessToken, refreshToken, err := service.Login(context.Background(), testCase.req, []byte("secret"))

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, accessToken)
				assert.Nil(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, accessToken)
				assert.NotNil(t, refreshToken)
				assert.NotEmpty(t, accessToken.Token)
				assert.NotEmpty(t, refreshToken.Token)
				mockRefreshRepo.AssertCalled(t, "Create", mock.Anything, userID, mock.Anything, mock.Anything)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshAccessToken(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()
	validator := "some-validator-token"
	selectorValidatorToken := tokenID.String() + ":" + validator

	mockUserRepo := new(MockAuthUserRepository)
	mockRefreshRepo := new(MockRefreshTokenRepository)

	// Hash the validator to simulate what's stored in database
	hashedValidator, _ := bcrypt.GenerateFromPassword([]byte(validator), bcrypt.DefaultCost)

	mockRefreshRepo.On("GetByID", mock.Anything, tokenID.String()).Return(&ent.RefreshToken{
		ID:     tokenID,
		UserID: userID,
		Token:  string(hashedValidator),
	}, nil)
	mockRefreshRepo.On("UpdateLastUsed", mock.Anything, tokenID).Return(nil)

	service := NewAuthService(mockUserRepo, mockRefreshRepo)
	accessToken, err := service.RefreshAccessToken(context.Background(), selectorValidatorToken, []byte("secret"))

	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotEmpty(t, accessToken.Token)
}

func TestAuthService_Logout(t *testing.T) {
	tokenID := uuid.New()
	validator := "some-validator-token"
	selectorValidatorToken := tokenID.String() + ":" + validator
	mockUserRepo := new(MockAuthUserRepository)
	mockRefreshRepo := new(MockRefreshTokenRepository)

	// Hash the validator to simulate what's stored in database
	hashedValidator, _ := bcrypt.GenerateFromPassword([]byte(validator), bcrypt.DefaultCost)

	mockRefreshRepo.On("GetByID", mock.Anything, tokenID.String()).Return(&ent.RefreshToken{
		ID:    tokenID,
		Token: string(hashedValidator),
	}, nil)
	mockRefreshRepo.On("RevokeToken", mock.Anything, tokenID).Return(nil)

	service := NewAuthService(mockUserRepo, mockRefreshRepo)
	err := service.Logout(context.Background(), selectorValidatorToken)

	assert.NoError(t, err)
	mockRefreshRepo.AssertCalled(t, "RevokeToken", mock.Anything, tokenID)
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}
