package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.Category, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Category), args.Error(1)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) GetAll(ctx context.Context) ([]*dto.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.Category), args.Error(1)
}

func TestCategoryService_Create(t *testing.T) {
	testCases := []struct {
		name          string
		request       *dto.CreateCategoryRequest
		mockSetup     func(*MockCategoryRepository, *dto.CreateCategoryRequest)
		expectedError string
	}{
		{
			name: "successful creation",
			request: &dto.CreateCategoryRequest{
				Name:         "Test Category",
				Description:  "Test Description",
				DisplayOrder: 1,
				IsActive:     true,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, req *dto.CreateCategoryRequest) {
				expectedCategory := &dto.Category{
					Name:         "Test Category",
					Description:  "Test Description",
					DisplayOrder: 1,
					IsActive:     true,
				}
				mockRepo.On("Create", mock.Anything, req).Return(expectedCategory, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			request: &dto.CreateCategoryRequest{
				Name:         "Test Category",
				Description:  "Test Description",
				DisplayOrder: 1,
				IsActive:     true,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, req *dto.CreateCategoryRequest) {
				mockRepo.On("Create", mock.Anything, req).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo, testCase.request)

			service := NewCategoryService(mockRepo)
			result, err := service.Create(t.Context(), testCase.request)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.request.Name, result.Name)
				assert.Equal(t, testCase.request.Description, result.Description)
				assert.Equal(t, testCase.request.DisplayOrder, result.DisplayOrder)
				assert.Equal(t, testCase.request.IsActive, result.IsActive)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_GetByID(t *testing.T) {
	testId := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful retrieval",
			id:   testId,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				expectedCategory := &dto.Category{
					ID: testId,
				}
				mockRepo.On("GetByID", mock.Anything, testId).Return(expectedCategory, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   testId,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetByID", mock.Anything, testId).Return(nil, errors.New("category not found"))
			},
			expectedError: "category not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			result, err := service.GetByID(t.Context(), testCase.id)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.id, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	id := uuid.New()
	name_new := "Updated Category"
	description_new := "Updated Description"
	displayOrder_new := 2
	isActive_new := false

	testCases := []struct {
		name          string
		id            uuid.UUID
		request       *dto.UpdateCategoryRequest
		mockSetup     func(*MockCategoryRepository, *dto.UpdateCategoryRequest)
		expectedError string
	}{
		{
			name: "successful update with all fields",
			id:   id,
			request: &dto.UpdateCategoryRequest{
				Name:         &name_new,
				Description:  &description_new,
				DisplayOrder: &displayOrder_new,
				IsActive:     &isActive_new,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, req *dto.UpdateCategoryRequest) {
				mockRepo.On("Update", mock.Anything, id, req).Return(&dto.Category{
					ID:           id,
					Name:         name_new,
					Description:  description_new,
					DisplayOrder: displayOrder_new,
					IsActive:     isActive_new,
				}, nil)
			},
			expectedError: "",
		},
		{
			name: "successful update with partial fields",
			id:   id,
			request: &dto.UpdateCategoryRequest{
				Name: &name_new,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, req *dto.UpdateCategoryRequest) {
				mockRepo.On("Update", mock.Anything, id, req).Return(&dto.Category{
					ID:           id,
					Name:         name_new,
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
				}, nil)
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo, testCase.request)

			service := NewCategoryService(mockRepo)
			result, err := service.Update(t.Context(), testCase.id, testCase.request)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.id, result.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCategoryService_Delete(t *testing.T) {
	id := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful deletion",
			id:   id,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("Delete", mock.Anything, id).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   id,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("Delete", mock.Anything, id).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			err := service.Delete(t.Context(), testCase.id)

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

func TestCategoryService_GetAll(t *testing.T) {
	testCases := []struct {
		name          string
		mockSetup     func(*MockCategoryRepository)
		expectedError string
		expectedCount int
	}{
		{
			name: "successful retrieval with categories",
			mockSetup: func(mockRepo *MockCategoryRepository) {
				categories := []*dto.Category{
					{
						Name:         "Category 1",
						Description:  "Description 1",
						DisplayOrder: 1,
						IsActive:     true,
					},
					{
						Name:         "Category 2",
						Description:  "Description 2",
						DisplayOrder: 2,
						IsActive:     true,
					},
				}
				mockRepo.On("GetAll", mock.Anything).Return(categories, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "successful retrieval with empty result",
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetAll", mock.Anything).Return([]*dto.Category{}, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
			expectedCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			result, err := service.GetAll(context.Background())

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, testCase.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
