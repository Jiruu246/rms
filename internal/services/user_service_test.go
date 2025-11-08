package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, req *dto.RegisterUserRequest) (*dto.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*dto.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserService_Register(t *testing.T) {
	testCases := []struct {
		name          string
		req           dto.RegisterUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError string
	}{
		{
			name: "successful registration",
			req: dto.RegisterUserRequest{
				Name:     "John Doe",
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(&dto.User{
					Name:  "John Doe",
					Email: "john.doe@example.com",
				}, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			req: dto.RegisterUserRequest{
				Name:     "John Doe",
				Email:    "john2.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("repository error"))
			},
			expectedError: "repository error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			testCase.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.Register(context.Background(), testCase.req)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.req.Name, result.Name)
				assert.Equal(t, testCase.req.Email, result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	testCases := []struct {
		name          string
		req           dto.LoginUserRequest
		mockSetup     func(*MockUserRepository)
		expectedError string
	}{
		{
			name: "successful login",
			req: dto.LoginUserRequest{
				Email:    "john.doe@example.com",
				Password: "password123",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(&dto.User{
					Email:    "john.doe@example.com",
					Password: hashPassword("password123"),
				}, nil)
			},
			expectedError: "",
		},
		{
			name: "invalid email or password",
			req: dto.LoginUserRequest{
				Email:    "john.doe@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(&dto.User{
					Email:    "john.doe@example.com",
					Password: hashPassword("password123"),
				}, nil)
			},
			expectedError: "invalid email or password",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			testCase.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.Login(context.Background(), testCase.req)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.req.Email, result.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetProfile(t *testing.T) {
	testUserId := uuid.New()

	testCases := []struct {
		name          string
		idInput       uuid.UUID
		mockSetup     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:    "successful profile retrieval",
			idInput: testUserId,
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, testUserId).Return(&dto.User{
					ID: testUserId,
				}, nil)
			},
			expectedError: "",
		},
		{
			name:    "user not found",
			idInput: uuid.New(),
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("user not found"))
			},
			expectedError: "user not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			testCase.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			result, err := service.GetProfile(context.Background(), testCase.idInput)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.idInput, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	id := uuid.New()
	name_new := "Updated Name"
	email_new := "updated.email@example.com"

	testCases := []struct {
		name          string
		idInput       uuid.UUID
		updateInput   *dto.UpdateUserRequest
		mockSetup     func(*MockUserRepository, *dto.UpdateUserRequest)
		expectedError string
	}{
		{
			name:    "successful update all fields",
			idInput: id,
			updateInput: &dto.UpdateUserRequest{
				Name:  &name_new,
				Email: &email_new,
			},
			mockSetup: func(mockRepo *MockUserRepository, updateReq *dto.UpdateUserRequest) {
				mockRepo.On("Update", mock.Anything, id, updateReq).Return(&dto.User{
					Name:  "Updated Name",
					Email: "updated.email@example.com",
				}, nil)
			},
			expectedError: "",
		},
		{
			name:    "successful update name only",
			idInput: id,
			updateInput: &dto.UpdateUserRequest{
				Name: &name_new,
			},
			mockSetup: func(mockRepo *MockUserRepository, updateReq *dto.UpdateUserRequest) {
				mockRepo.On("Update", mock.Anything, id, updateReq).Return(&dto.User{
					Name:  "Updated Name",
					Email: "old.email@example.com",
				}, nil)
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			testCase.mockSetup(mockRepo, testCase.updateInput)

			service := NewUserService(mockRepo)
			result, err := service.UpdateProfile(context.Background(), testCase.idInput, testCase.updateInput)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, *testCase.updateInput.Name, result.Name)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteAccount(t *testing.T) {
	testCases := []struct {
		name          string
		idInput       uuid.UUID
		mockSetup     func(*MockUserRepository)
		expectedError string
	}{
		{
			name:    "successful account deletion",
			idInput: uuid.New(),
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: "",
		},
		{
			name:    "user not found",
			idInput: uuid.New(),
			mockSetup: func(mockRepo *MockUserRepository) {
				mockRepo.On("Delete", mock.Anything, mock.Anything).Return(errors.New("user not found"))
			},
			expectedError: "user not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			testCase.mockSetup(mockRepo)

			service := NewUserService(mockRepo)
			err := service.DeleteAccount(context.Background(), testCase.idInput)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}
