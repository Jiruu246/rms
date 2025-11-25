package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRestaurantRepository is a mock implementation of RestaurantRepository
type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Create(ctx context.Context, restaurant *ent.Restaurant) (*ent.Restaurant, error) {
	args := m.Called(ctx, restaurant)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Update(ctx context.Context, restaurant *ent.Restaurant) (*ent.Restaurant, error) {
	args := m.Called(ctx, restaurant)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ent.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantRepository) GetAll(ctx context.Context) ([]*ent.Restaurant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ent.Restaurant), args.Error(1)
}

func TestRestaurantService_Create(t *testing.T) {
	operatingHours := map[string]any{
		"monday": map[string]string{
			"open":  "09:00",
			"close": "22:00",
		},
	}

	testCases := []struct {
		name          string
		request       *dto.CreateRestaurantRequest
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name: "successful creation with all fields",
			request: &dto.CreateRestaurantRequest{
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
				Status:         "active",
				OperatingHours: operatingHours,
				Currency:       "USD",
			},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
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
					Status:         restaurant.StatusActive,
					OperatingHours: operatingHours,
					Currency:       "USD",
				}
				mockRepo.On("Create", mock.Anything, expectedRestaurant).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful creation with minimal required fields",
			request: &dto.CreateRestaurantRequest{
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
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   restaurant.StatusActive,
					Currency: "USD",
				}
				mockRepo.On("Create", mock.Anything, expectedRestaurant).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful creation with inactive status",
			request: &dto.CreateRestaurantRequest{
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
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   restaurant.StatusInactive,
					Currency: "USD",
				}
				mockRepo.On("Create", mock.Anything, expectedRestaurant).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful creation with closed status",
			request: &dto.CreateRestaurantRequest{
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
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   restaurant.StatusClosed,
					Currency: "USD",
				}
				mockRepo.On("Create", mock.Anything, expectedRestaurant).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful creation with invalid status defaults to active",
			request: &dto.CreateRestaurantRequest{
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
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
					Name:     "Test Restaurant",
					Phone:    "+1234567890",
					Email:    "test@restaurant.com",
					Address:  "123 Test St",
					City:     "Test City",
					State:    "Test State",
					ZipCode:  "12345",
					Country:  "Test Country",
					Status:   restaurant.StatusActive,
					Currency: "USD",
				}
				mockRepo.On("Create", mock.Anything, expectedRestaurant).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			request: &dto.CreateRestaurantRequest{
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
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
			result, err := service.Create(t.Context(), testCase.request)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.request.Name, result.Name)
				assert.Equal(t, testCase.request.Phone, result.Phone)
				assert.Equal(t, testCase.request.Email, result.Email)
				assert.Equal(t, testCase.request.Currency, result.Currency)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRestaurantService_GetByID(t *testing.T) {
	testId := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name: "successful retrieval",
			id:   testId,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				expectedRestaurant := &ent.Restaurant{
					ID:   testId,
					Name: "Test Restaurant",
				}
				mockRepo.On("GetByID", mock.Anything, testId).Return(expectedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   testId,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetByID", mock.Anything, testId).Return(nil, errors.New("restaurant not found"))
			},
			expectedError: "restaurant not found",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
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

func TestRestaurantService_Update(t *testing.T) {
	id := uuid.New()
	nameNew := "Updated Restaurant"
	descriptionNew := "Updated Description"
	phoneNew := "+0987654321"
	emailNew := "updated@restaurant.com"
	statusNew := "inactive"
	currencyNew := "EUR"

	testCases := []struct {
		name          string
		id            uuid.UUID
		request       *dto.UpdateRestaurantRequest
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name: "successful update with all fields",
			id:   id,
			request: &dto.UpdateRestaurantRequest{
				Name:        &nameNew,
				Description: &descriptionNew,
				Phone:       &phoneNew,
				Email:       &emailNew,
				Status:      &statusNew,
				Currency:    &currencyNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				existingRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        "Old Restaurant",
					Description: "Old Description",
					Phone:       "+1234567890",
					Email:       "old@restaurant.com",
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusActive,
					Currency:    "USD",
				}
				updatedRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        nameNew,
					Description: descriptionNew,
					Phone:       phoneNew,
					Email:       emailNew,
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusInactive,
					Currency:    currencyNew,
				}
				mockRepo.On("GetByID", mock.Anything, id).Return(existingRestaurant, nil)
				mockRepo.On("Update", mock.Anything, updatedRestaurant).Return(updatedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful update with partial fields",
			id:   id,
			request: &dto.UpdateRestaurantRequest{
				Name:  &nameNew,
				Email: &emailNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				existingRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        "Old Restaurant",
					Description: "Old Description",
					Phone:       "+1234567890",
					Email:       "old@restaurant.com",
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusActive,
					Currency:    "USD",
				}
				updatedRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        nameNew,
					Description: "Old Description",
					Phone:       "+1234567890",
					Email:       emailNew,
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusActive,
					Currency:    "USD",
				}
				mockRepo.On("GetByID", mock.Anything, id).Return(existingRestaurant, nil)
				mockRepo.On("Update", mock.Anything, updatedRestaurant).Return(updatedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name: "successful update with status change to closed",
			id:   id,
			request: &dto.UpdateRestaurantRequest{
				Status: func() *string { s := "closed"; return &s }(),
			},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				existingRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        "Test Restaurant",
					Description: "Test Description",
					Phone:       "+1234567890",
					Email:       "test@restaurant.com",
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusActive,
					Currency:    "USD",
				}
				updatedRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        "Test Restaurant",
					Description: "Test Description",
					Phone:       "+1234567890",
					Email:       "test@restaurant.com",
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusClosed,
					Currency:    "USD",
				}
				mockRepo.On("GetByID", mock.Anything, id).Return(existingRestaurant, nil)
				mockRepo.On("Update", mock.Anything, updatedRestaurant).Return(updatedRestaurant, nil)
			},
			expectedError: "",
		},
		{
			name:    "restaurant not found",
			id:      id,
			request: &dto.UpdateRestaurantRequest{Name: &nameNew},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))
			},
			expectedError: "restaurant not found",
		},
		{
			name: "repository update error",
			id:   id,
			request: &dto.UpdateRestaurantRequest{
				Name: &nameNew,
			},
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				existingRestaurant := &ent.Restaurant{
					ID:          id,
					Name:        "Old Restaurant",
					Description: "Old Description",
					Phone:       "+1234567890",
					Email:       "old@restaurant.com",
					Address:     "123 Test St",
					City:        "Test City",
					State:       "Test State",
					ZipCode:     "12345",
					Country:     "Test Country",
					Status:      restaurant.StatusActive,
					Currency:    "USD",
				}
				mockRepo.On("GetByID", mock.Anything, id).Return(existingRestaurant, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
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

func TestRestaurantService_Delete(t *testing.T) {
	id := uuid.New()

	testCases := []struct {
		name          string
		id            uuid.UUID
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
	}{
		{
			name: "successful deletion",
			id:   id,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("Delete", mock.Anything, id).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "repository error",
			id:   id,
			mockSetup: func(mockRepo *MockRestaurantRepository) {
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

func TestRestaurantService_GetAll(t *testing.T) {
	testCases := []struct {
		name          string
		mockSetup     func(*MockRestaurantRepository)
		expectedError string
		expectedCount int
	}{
		{
			name: "successful retrieval with restaurants",
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				restaurants := []*ent.Restaurant{
					{
						ID:       uuid.New(),
						Name:     "Restaurant 1",
						Phone:    "+1234567890",
						Email:    "rest1@example.com",
						Currency: "USD",
						Status:   restaurant.StatusActive,
					},
					{
						ID:       uuid.New(),
						Name:     "Restaurant 2",
						Phone:    "+0987654321",
						Email:    "rest2@example.com",
						Currency: "EUR",
						Status:   restaurant.StatusInactive,
					},
				}
				mockRepo.On("GetAll", mock.Anything).Return(restaurants, nil)
			},
			expectedError: "",
			expectedCount: 2,
		},
		{
			name: "successful retrieval with empty result",
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAll", mock.Anything).Return([]*ent.Restaurant{}, nil)
			},
			expectedError: "",
			expectedCount: 0,
		},
		{
			name: "repository error",
			mockSetup: func(mockRepo *MockRestaurantRepository) {
				mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError: "database error",
			expectedCount: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)
			testCase.mockSetup(mockRepo)

			service := NewRestaurantService(mockRepo)
			result, err := service.GetAll(t.Context())

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
