package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// adminActor bypasses ownership checks in PolicyAuthorizer, letting these tests
// focus on service/repo wiring rather than re-testing authz.PolicyAuthorizer itself.
var adminActor = authz.Actor{Role: authz.RoleAdmin}

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

func (m *MockCategoryRepository) GetByID(ctx context.Context, restaurantID, id uuid.UUID) (*dto.Category, error) {
	args := m.Called(ctx, restaurantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Category), args.Error(1)
}

func (m *MockCategoryRepository) Update(ctx context.Context, restaurantID, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	args := m.Called(ctx, restaurantID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Category), args.Error(1)
}

func (m *MockCategoryRepository) Delete(ctx context.Context, restaurantID, id uuid.UUID) error {
	args := m.Called(ctx, restaurantID, id)
	return args.Error(0)
}

func (m *MockCategoryRepository) List(ctx context.Context, restaurantID uuid.UUID, req pagination.PageRequest) (*pagination.PageResponse[*dto.CategoryListItem], error) {
	args := m.Called(ctx, restaurantID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pagination.PageResponse[*dto.CategoryListItem]), args.Error(1)
}

func (m *MockCategoryRepository) GetAuthorizationResource(ctx context.Context, id uuid.UUID) (authz.Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(authz.Resource), args.Error(1)
}

// MockRestaurantService is a mock implementation of RestaurantService for use
// by CategoryService tests, which only ever call AuthorizeOwnership.
type MockRestaurantService struct {
	mock.Mock
}

func (m *MockRestaurantService) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantService) GetByID(ctx context.Context, actor authz.Actor, id uuid.UUID) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, actor, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantService) GetAll(ctx context.Context, actor authz.Actor) ([]*dto.RestaurantResponse, error) {
	args := m.Called(ctx, actor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantService) Update(ctx context.Context, actor authz.Actor, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, actor, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantService) Delete(ctx context.Context, actor authz.Actor, id uuid.UUID) error {
	args := m.Called(ctx, actor, id)
	return args.Error(0)
}

func (m *MockRestaurantService) AuthorizeOwnership(ctx context.Context, actor authz.Actor, action authz.Action, restaurantID uuid.UUID) error {
	args := m.Called(ctx, actor, action, restaurantID)
	return args.Error(0)
}

func TestCategoryService_Create(t *testing.T) {
	restaurantID := uuid.New()

	testCases := []struct {
		name          string
		request       *dto.CreateCategoryRequest
		mockSetup     func(*MockCategoryRepository, *MockRestaurantService, *dto.CreateCategoryRequest)
		expectedError string
	}{
		{
			name: "successful creation",
			request: &dto.CreateCategoryRequest{
				Name:         "Test Category",
				Description:  "Test Description",
				DisplayOrder: 1,
				IsActive:     true,
				RestaurantID: restaurantID,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, mockRestaurantService *MockRestaurantService, req *dto.CreateCategoryRequest) {
				mockRestaurantService.On("AuthorizeOwnership", mock.Anything, adminActor, ActionCreateCategory, restaurantID).
					Return(nil)
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
				RestaurantID: restaurantID,
			},
			mockSetup: func(mockRepo *MockCategoryRepository, mockRestaurantService *MockRestaurantService, req *dto.CreateCategoryRequest) {
				mockRestaurantService.On("AuthorizeOwnership", mock.Anything, adminActor, ActionCreateCategory, restaurantID).
					Return(nil)
				mockRepo.On("Create", mock.Anything, req).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			mockRestaurantService := new(MockRestaurantService)
			testCase.mockSetup(mockRepo, mockRestaurantService, testCase.request)

			service := NewCategoryService(mockRepo, mockRestaurantService)
			result, err := service.Create(t.Context(), adminActor, testCase.request)

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
			mockRestaurantService.AssertExpectations(t)
		})
	}
}

func TestCategoryService_GetByID(t *testing.T) {
	testId := uuid.New()
	restaurantID := uuid.New()

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
				mockRepo.On("GetAuthorizationResource", mock.Anything, testId).
					Return(authz.Resource{Type: "category", ID: testId, RestaurantID: restaurantID}, nil)
				expectedCategory := &dto.Category{
					ID: testId,
				}
				mockRepo.On("GetByID", mock.Anything, restaurantID, testId).Return(expectedCategory, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   testId,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, testId).
					Return(authz.Resource{Type: "category", ID: testId, RestaurantID: restaurantID}, nil)
				mockRepo.On("GetByID", mock.Anything, restaurantID, testId).Return(nil, errors.New("category not found"))
			},
			expectedError: "category not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo, new(MockRestaurantService))
			result, err := service.GetByID(t.Context(), adminActor, testCase.id)

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
	restaurantID := uuid.New()
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
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).
					Return(authz.Resource{Type: "category", ID: id, RestaurantID: restaurantID}, nil)
				mockRepo.On("Update", mock.Anything, restaurantID, id, req).Return(&dto.Category{
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
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).
					Return(authz.Resource{Type: "category", ID: id, RestaurantID: restaurantID}, nil)
				mockRepo.On("Update", mock.Anything, restaurantID, id, req).Return(&dto.Category{
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

			service := NewCategoryService(mockRepo, new(MockRestaurantService))
			result, err := service.Update(t.Context(), adminActor, testCase.id, testCase.request)

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
	restaurantID := uuid.New()

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
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).
					Return(authz.Resource{Type: "category", ID: id, RestaurantID: restaurantID}, nil)
				mockRepo.On("Delete", mock.Anything, restaurantID, id).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   id,
			mockSetup: func(mockRepo *MockCategoryRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).
					Return(authz.Resource{Type: "category", ID: id, RestaurantID: restaurantID}, nil)
				mockRepo.On("Delete", mock.Anything, restaurantID, id).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			testCase.mockSetup(mockRepo)

			service := NewCategoryService(mockRepo, new(MockRestaurantService))
			err := service.Delete(t.Context(), adminActor, testCase.id)

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

func TestCategoryService_List(t *testing.T) {
	req := pagination.PageRequest{Limit: 20}
	restaurantID := uuid.New()

	testCases := []struct {
		name          string
		mockSetup     func(*MockCategoryRepository, *MockRestaurantService)
		expectedError string
		expectedCount int
	}{
		{
			name: "successful retrieval with categories",
			mockSetup: func(mockRepo *MockCategoryRepository, mockRestaurantService *MockRestaurantService) {
				mockRestaurantService.On("AuthorizeOwnership", mock.Anything, adminActor, ActionReadCategory, restaurantID).
					Return(nil)
				page := &pagination.PageResponse[*dto.CategoryListItem]{
					Data: []*dto.CategoryListItem{
						{Name: "Category 1"},
						{Name: "Category 2"},
					},
				}
				mockRepo.On("List", mock.Anything, restaurantID, req).Return(page, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "successful retrieval with empty result",
			mockSetup: func(mockRepo *MockCategoryRepository, mockRestaurantService *MockRestaurantService) {
				mockRestaurantService.On("AuthorizeOwnership", mock.Anything, adminActor, ActionReadCategory, restaurantID).
					Return(nil)
				page := &pagination.PageResponse[*dto.CategoryListItem]{Data: []*dto.CategoryListItem{}}
				mockRepo.On("List", mock.Anything, restaurantID, req).Return(page, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockCategoryRepository, mockRestaurantService *MockRestaurantService) {
				mockRestaurantService.On("AuthorizeOwnership", mock.Anything, adminActor, ActionReadCategory, restaurantID).
					Return(nil)
				mockRepo.On("List", mock.Anything, restaurantID, req).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
			expectedCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockCategoryRepository)
			mockRestaurantService := new(MockRestaurantService)
			testCase.mockSetup(mockRepo, mockRestaurantService)

			service := NewCategoryService(mockRepo, mockRestaurantService)
			result, err := service.List(context.Background(), adminActor, restaurantID, req)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Data, testCase.expectedCount)
			}

			mockRepo.AssertExpectations(t)
			mockRestaurantService.AssertExpectations(t)
		})
	}
}
