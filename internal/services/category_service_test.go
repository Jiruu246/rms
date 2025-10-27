package services

import (
	"errors"
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) Create(category *models.Category) (*models.Category, error) {
	args := m.Called(category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(id uint) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(category *models.Category) (*models.Category, error) {
	args := m.Called(category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestCategoryService_Create(t *testing.T) {
	testCases := []struct {
		name          string
		request       *models.CreateCategoryRequest
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful creation",
			request: &models.CreateCategoryRequest{
				Name:         "Test Category",
				Description:  "Test Description",
				DisplayOrder: 1,
				IsActive:     true,
			},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				expectedCategory := &models.Category{
					ID:           1,
					Name:         "Test Category",
					Description:  "Test Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(expectedCategory, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			request: &models.CreateCategoryRequest{
				Name:         "Test Category",
				Description:  "Test Description",
				DisplayOrder: 1,
				IsActive:     true,
			},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			result, err := service.Create(testCase.request)

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
	testCases := []struct {
		name          string
		id            uint
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful retrieval",
			id:   1,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				expectedCategory := &models.Category{
					ID:           1,
					Name:         "Test Category",
					Description:  "Test Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("GetByID", uint(1)).Return(expectedCategory, nil)
			},
			expectedError: "",
		},
		{
			name:          "invalid id zero",
			id:            0,
			mockSetup:     func(mockRepo *MockCategoryRepository) {},
			expectedError: "invalid category id",
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetByID", uint(1)).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			result, err := service.GetByID(testCase.id)

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
	name := "Updated Category"
	description := "Updated Description"
	displayOrder := 2
	isActive := false

	testCases := []struct {
		name          string
		id            uint
		request       *models.UpdateCategoryRequest
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful update with all fields",
			id:   1,
			request: &models.UpdateCategoryRequest{
				Name:         &name,
				Description:  &description,
				DisplayOrder: &displayOrder,
				IsActive:     &isActive,
			},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				existingCategory := &models.Category{
					ID:           1,
					Name:         "Old Category",
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				updatedCategory := &models.Category{
					ID:           1,
					Name:         name,
					Description:  description,
					DisplayOrder: displayOrder,
					IsActive:     isActive,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("GetByID", uint(1)).Return(existingCategory, nil)
				mockRepo.On("Update", mock.AnythingOfType("*models.Category")).Return(updatedCategory, nil)
			},
			expectedError: "",
		},
		{
			name: "successful update with partial fields",
			id:   1,
			request: &models.UpdateCategoryRequest{
				Name: &name,
			},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				existingCategory := &models.Category{
					ID:           1,
					Name:         "Old Category",
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				updatedCategory := &models.Category{
					ID:           1,
					Name:         name,
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("GetByID", uint(1)).Return(existingCategory, nil)
				mockRepo.On("Update", mock.AnythingOfType("*models.Category")).Return(updatedCategory, nil)
			},
			expectedError: "",
		},
		{
			name:          "invalid id zero",
			id:            0,
			request:       &models.UpdateCategoryRequest{Name: &name},
			mockSetup:     func(mockRepo *MockCategoryRepository) {},
			expectedError: "invalid category id",
		},
		{
			name:    "empty name validation",
			id:      1,
			request: &models.UpdateCategoryRequest{Name: func() *string { s := ""; return &s }()},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				existingCategory := &models.Category{
					ID:           1,
					Name:         "Old Category",
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("GetByID", uint(1)).Return(existingCategory, nil)
			},
			expectedError: "name cannot be empty",
		},
		{
			name:    "no fields provided",
			id:      1,
			request: &models.UpdateCategoryRequest{},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				existingCategory := &models.Category{
					ID:           1,
					Name:         "Old Category",
					Description:  "Old Description",
					DisplayOrder: 1,
					IsActive:     true,
					CreatedAt:    time.Now(),
				}
				mockRepo.On("GetByID", uint(1)).Return(existingCategory, nil)
			},
			expectedError: "no valid fields provided for update",
		},
		{
			name:    "category not found",
			id:      999,
			request: &models.UpdateCategoryRequest{Name: &name},
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetByID", uint(999)).Return(nil, errors.New("not found"))
			},
			expectedError: "category not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			result, err := service.Update(testCase.id, testCase.request)

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
	testCases := []struct {
		name          string
		id            uint
		mockSetup     func(*MockCategoryRepository)
		expectedError string
	}{
		{
			name: "successful deletion",
			id:   1,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("Delete", uint(1)).Return(nil)
			},
			expectedError: "",
		},
		{
			name:          "invalid id zero",
			id:            0,
			mockSetup:     func(mockRepo *MockCategoryRepository) {},
			expectedError: "invalid category id",
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("Delete", uint(1)).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo)
			err := service.Delete(testCase.id)

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
