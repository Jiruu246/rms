package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantRepository) Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.RestaurantResponse, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantRepository) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]*dto.RestaurantResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.RestaurantResponse), args.Error(1)
}

func (m *MockRestaurantRepository) GetAuthorizationResource(ctx context.Context, id uuid.UUID) (authz.Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(authz.Resource), args.Error(1)
}

func TestRestaurantService_Create(t *testing.T) {
	operatingHours := map[string]any{
		"monday": map[string]string{
			"open":  "09:00",
			"close": "22:00",
		},
	}

	type testCase struct {
		name          string
		input         *dto.CreateRestaurantData
		expected      *dto.RestaurantResponse
		expectedError string
	}

	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()
	uuid4 := uuid.New()
	uuid5 := uuid.New()
	uuid6 := uuid.New()

	testCases := []testCase{
		{
			name: "successful creation with all fields",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:           "Test Restaurant",
					Description:    "A wonderful test restaurant",
					Phone:          "+1234567890",
					Email:          "test@restaurant.com",
					Address:        "123 Test St",
					City:           "Test City",
					State:          "Test State",
					ZipCode:        "12345",
					Country:        "Test Country",
					LogoURL:        "https://example.com/logo.png",
					CoverImageURL:  "https://example.com/cover.jpg",
					Status:         restaurant.StatusActive.String(),
					OperatingHours: operatingHours,
					Currency:       "USD",
				},
				UserID: uuid1,
			},
			expected: &dto.RestaurantResponse{
				Name:           "Test Restaurant",
				Description:    "A wonderful test restaurant",
				Phone:          "+1234567890",
				Email:          "test@restaurant.com",
				Address:        "123 Test St",
				City:           "Test City",
				State:          "Test State",
				ZipCode:        "12345",
				Country:        "Test Country",
				LogoURL:        "https://example.com/logo.png",
				CoverImageURL:  "https://example.com/cover.jpg",
				Status:         restaurant.StatusActive.String(),
				OperatingHours: operatingHours,
				Currency:       "USD",
			},
			expectedError: "",
		},
		{
			name: "successful creation with minimal required fields",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Currency: "USD",
				},
				UserID: uuid2,
			},
			expected: &dto.RestaurantResponse{
				Name:     "Test Restaurant",
				Phone:    "+1234567890",
				Email:    "test@restaurant.com",
				Address:  "123 Test St",
				City:     "Test City",
				State:    "Test State",
				ZipCode:  "12345",
				Country:  "Test Country",
				Status:   restaurant.StatusActive.String(),
				Currency: "USD",
			},
			expectedError: "",
		},
		{
			name: "successful creation with inactive status",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   "inactive",
					Currency: "USD",
				},
				UserID: uuid3,
			},
			expected: &dto.RestaurantResponse{
				Name:     "Test Restaurant",
				Phone:    "+1234567890",
				Email:    "test@restaurant.com",
				Address:  "123 Test St",
				City:     "Test City",
				State:    "Test State",
				ZipCode:  "12345",
				Country:  "Test Country",
				Status:   restaurant.StatusInactive.String(),
				Currency: "USD",
			},
			expectedError: "",
		},
		{
			name: "successful creation with closed status",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   "closed",
					Currency: "USD",
				},
				UserID: uuid4,
			},
			expected: &dto.RestaurantResponse{
				Name:     "Test Restaurant",
				Phone:    "+1234567890",
				Email:    "test@restaurant.com",
				Address:  "123 Test St",
				City:     "Test City",
				State:    "Test State",
				ZipCode:  "12345",
				Country:  "Test Country",
				Status:   restaurant.StatusClosed.String(),
				Currency: "USD",
			},
			expectedError: "",
		},
		{
			name: "successful creation with invalid status defaults to active",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   "invalid_status",
					Currency: "USD",
				},
				UserID: uuid5,
			},
			expected: &dto.RestaurantResponse{
				Name:     "Test Restaurant",
				Phone:    "+1234567890",
				Email:    "test@restaurant.com",
				Address:  "123 Test St",
				City:     "Test City",
				State:    "Test State",
				ZipCode:  "12345",
				Country:  "Test Country",
				Status:   restaurant.StatusActive.String(),
				Currency: "USD",
			},
			expectedError: "",
		},
		{
			name: "repository error",
			input: &dto.CreateRestaurantData{
				Request: &dto.CreateRestaurantRequest{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Currency: "USD",
				},
				UserID: uuid6,
			},
			expected:      nil,
			expectedError: "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			if tc.expectedError != "" {
				mockRepo.On("Create", mock.Anything, tc.input).Return(nil, errors.New(tc.expectedError))
			} else {
				mockRepo.On("Create", mock.Anything, tc.input).Return(tc.expected, nil)
			}

			service := NewRestaurantService(mockRepo)
			result, err := service.Create(t.Context(), tc.input)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRestaurantService_GetByID(t *testing.T) {
	testId := uuid.New()
	ownerId := uuid.New()
	owner := authz.Actor{UserID: ownerId}
	stranger := authz.Actor{UserID: uuid.New()}

	testCases := []struct {
		name          string
		id            uuid.UUID
		actor         authz.Actor
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name:  "successful retrieval",
			id:    testId,
			actor: owner,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, testId).
					Return(authz.Resource{ID: testId, RestaurantID: testId, OwnerUserID: ownerId}, nil)
				expectedRestaurant := &dto.RestaurantResponse{
					ID:   testId,
					Name: "Test Restaurant",
				}
				mockRepo.On("GetByID", mock.Anything, testId).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name:  "repository error",
			id:    testId,
			actor: owner,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, testId).
					Return(authz.Resource{ID: testId, RestaurantID: testId, OwnerUserID: ownerId}, nil)
				mockRepo.On("GetByID", mock.Anything, testId).Return(nil, errors.New("restaurant not found"))
			},
			expectedError: "restaurant not found",
		},
		{
			name:  "forbidden - not the owner",
			id:    testId,
			actor: stranger,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, testId).
					Return(authz.Resource{ID: testId, RestaurantID: testId, OwnerUserID: ownerId}, nil)
			},
			expectedError: apperr.ErrForbidden.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
			result, err := service.GetByID(t.Context(), testCase.actor, testCase.id)

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

func TestRestaurantService_Update(t *testing.T) {
	restaurantId := uuid.New()
	ownerId := uuid.New()
	owner := authz.Actor{UserID: ownerId}
	stranger := authz.Actor{UserID: uuid.New()}
	resource := authz.Resource{ID: restaurantId, RestaurantID: restaurantId, OwnerUserID: ownerId}

	nameNew := "Updated Restaurant"
	descriptionNew := "Updated Description"
	phoneNew := "+0987654321"
	emailNew := "updated@restaurant.com"
	statusNew := "inactive"
	currencyNew := "EUR"

	testCases := []struct {
		name          string
		actor         authz.Actor
		request       *dto.UpdateRestaurantRequest
		mockSetup     func(*MockRestaurantRepository, *dto.UpdateRestaurantRequest)
		expected      *dto.RestaurantResponse
		expectedError string
	}{
		{
			name:  "successful update with all fields",
			actor: owner,
			request: &dto.UpdateRestaurantRequest{
				Name:        &nameNew,
				Description: &descriptionNew,
				Phone:       &phoneNew,
				Email:       &emailNew,
				Status:      &statusNew,
				Currency:    &currencyNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository, req *dto.UpdateRestaurantRequest) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantId).Return(resource, nil)
				mockRepo.On("Update", mock.Anything, &dto.UpdateRestaurantData{Request: req, ID: restaurantId}).
					Return(&dto.RestaurantResponse{
						ID:          restaurantId,
						Name:        nameNew,
						Description: descriptionNew,
						Phone:       phoneNew,
						Email:       emailNew,
						Address:     "123 Test St",
						City:        "Test City",
						State:       "Test State",
						ZipCode:     "12345",
						Country:     "Test Country",
						Status:      restaurant.StatusInactive.String(),
						Currency:    currencyNew,
					}, nil)
			},
			expected: &dto.RestaurantResponse{
				ID:          restaurantId,
				Name:        nameNew,
				Description: descriptionNew,
				Phone:       phoneNew,
				Email:       emailNew,
				Address:     "123 Test St",
				City:        "Test City",
				State:       "Test State",
				ZipCode:     "12345",
				Country:     "Test Country",
				Status:      restaurant.StatusInactive.String(),
				Currency:    currencyNew,
			},
			expectedError: "",
		},
		{
			name:  "successful update with partial fields",
			actor: owner,
			request: &dto.UpdateRestaurantRequest{
				Name:  &nameNew,
				Email: &emailNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository, req *dto.UpdateRestaurantRequest) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantId).Return(resource, nil)
				mockRepo.On("Update", mock.Anything, &dto.UpdateRestaurantData{Request: req, ID: restaurantId}).
					Return(&dto.RestaurantResponse{
						ID:          restaurantId,
						Name:        nameNew,
						Description: "Old Description",
						Phone:       "+1234567890",
						Email:       emailNew,
						Address:     "123 Test St",
						City:        "Test City",
						State:       "Test State",
						ZipCode:     "12345",
						Country:     "Test Country",
						Status:      restaurant.StatusActive.String(),
						Currency:    "USD",
					}, nil)
			},
			expected: &dto.RestaurantResponse{
				ID:          restaurantId,
				Name:        nameNew,
				Description: "Old Description",
				Phone:       "+1234567890",
				Email:       emailNew,
				Address:     "123 Test St",
				City:        "Test City",
				State:       "Test State",
				ZipCode:     "12345",
				Country:     "Test Country",
				Status:      restaurant.StatusActive.String(),
				Currency:    "USD",
			},
			expectedError: "",
		},
		{
			name:  "forbidden - not the owner",
			actor: stranger,
			request: &dto.UpdateRestaurantRequest{
				Name: &nameNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository, req *dto.UpdateRestaurantRequest) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantId).Return(resource, nil)
			},
			expectedError: apperr.ErrForbidden.Error(),
		},
		{
			name:  "restaurant not found",
			actor: owner,
			request: &dto.UpdateRestaurantRequest{
				Name: &nameNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository, req *dto.UpdateRestaurantRequest) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantId).
					Return(authz.Resource{}, errors.New("restaurant not found"))
			},
			expectedError: "restaurant not found",
		},
		{
			name:  "repository update error",
			actor: owner,
			request: &dto.UpdateRestaurantRequest{
				Name: &nameNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository, req *dto.UpdateRestaurantRequest) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantId).Return(resource, nil)
				mockRepo.On("Update", mock.Anything, &dto.UpdateRestaurantData{Request: req, ID: restaurantId}).
					Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo, testCase.request)

			service := NewRestaurantService(mockRepo)
			result, err := service.Update(t.Context(), testCase.actor, restaurantId, testCase.request)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRestaurantService_Delete(t *testing.T) {
	id := uuid.New()
	ownerId := uuid.New()
	owner := authz.Actor{UserID: ownerId}
	stranger := authz.Actor{UserID: uuid.New()}
	resource := authz.Resource{ID: id, RestaurantID: id, OwnerUserID: ownerId}

	testCases := []struct {
		name          string
		id            uuid.UUID
		actor         authz.Actor
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name:  "successful deletion",
			id:    id,
			actor: owner,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).Return(resource, nil)
				mockRepo.On("Delete", mock.Anything, id).Return(nil)
			},
			expectedError: "",
		},
		{
			name:  "forbidden - not the owner",
			id:    id,
			actor: stranger,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).Return(resource, nil)
			},
			expectedError: apperr.ErrForbidden.Error(),
		},
		{
			name:  "repository error",
			id:    id,
			actor: owner,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAuthorizationResource", mock.Anything, id).Return(resource, nil)
				mockRepo.On("Delete", mock.Anything, id).Return(errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
			err := service.Delete(t.Context(), testCase.actor, testCase.id)

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

func TestRestaurantService_GetAll(t *testing.T) {
	ownerID := uuid.New()

	testCases := []struct {
		name          string
		actor         authz.Actor
		expected      []*dto.RestaurantResponse
		expectedError string
	}{
		{
			name:  "scoped to actor's own restaurants",
			actor: authz.Actor{UserID: ownerID, Role: "owner"},
			expected: []*dto.RestaurantResponse{
				{
					ID:       uuid.New(),
					Name:     "Restaurant 1",
					Phone:    "+1234567890",
					Email:    "rest1@example.com",
					Currency: "USD",
					Status:   restaurant.StatusActive.String(),
				},
				{
					ID:       uuid.New(),
					Name:     "Restaurant 2",
					Phone:    "+0987654321",
					Email:    "rest2@example.com",
					Currency: "EUR",
					Status:   restaurant.StatusInactive.String(),
				},
			},
			expectedError: "",
		},
		{
			name:  "admin is scoped too — no sees-all support yet",
			actor: authz.Actor{UserID: ownerID, Role: authz.RoleAdmin},
			expected: []*dto.RestaurantResponse{
				{
					ID:       uuid.New(),
					Name:     "Restaurant 1",
					Phone:    "+1234567890",
					Email:    "rest1@example.com",
					Currency: "USD",
					Status:   restaurant.StatusActive.String(),
				},
			},
			expectedError: "",
		},
		{
			name:          "successful retrieval with empty result",
			actor:         authz.Actor{UserID: ownerID, Role: "owner"},
			expected:      []*dto.RestaurantResponse{},
			expectedError: "",
		},
		{
			name:          "repository error",
			actor:         authz.Actor{UserID: ownerID, Role: "owner"},
			expected:      nil,
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)

			if testCase.expectedError != "" {
				mockRepo.On("GetAllForUser", mock.Anything, testCase.actor.UserID).Return(nil, errors.New(testCase.expectedError))
			} else {
				mockRepo.On("GetAllForUser", mock.Anything, testCase.actor.UserID).Return(testCase.expected, nil)
			}

			service := NewRestaurantService(mockRepo)
			result, err := service.GetAll(t.Context(), testCase.actor)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, result, testCase.expected)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
