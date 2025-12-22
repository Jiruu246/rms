package services

import (
	"context"
	"errors"
	"testing"

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

func (m *MockRestaurantRepository) GetAll(ctx context.Context) ([]*dto.RestaurantResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.RestaurantResponse), args.Error(1)
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
				expectedRestaurant := &dto.RestaurantResponse{
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
		request       *dto.UpdateRestaurantRequest
		expected      *dto.RestaurantResponse
		expectedError string
	}{
		{
			name: "successful update with all fields",
			request: &dto.UpdateRestaurantRequest{
				Name:        &nameNew,
				Description: &descriptionNew,
				Phone:       &phoneNew,
				Email:       &emailNew,
				Status:      &statusNew,
				Currency:    &currencyNew,
			},
			expected: &dto.RestaurantResponse{
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
				Status:      restaurant.StatusInactive.String(),
				Currency:    currencyNew,
			},
			expectedError: "",
		},
		{
			name: "successful update with partial fields",
			request: &dto.UpdateRestaurantRequest{
				Name:  &nameNew,
				Email: &emailNew,
			},
			expected: &dto.RestaurantResponse{
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
				Status:      restaurant.StatusActive.String(),
				Currency:    "USD",
			},
			expectedError: "",
		},
		{
			name: "successful update with status change to closed",
			request: &dto.UpdateRestaurantRequest{
				Status: func() *string { s := "closed"; return &s }(),
			},
			expected: &dto.RestaurantResponse{
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
				Status:      restaurant.StatusClosed.String(),
				Currency:    "USD",
			},

			expectedError: "",
		},
		{
			name:          "restaurant not found",
			request:       &dto.UpdateRestaurantRequest{Name: &nameNew},
			expected:      nil,
			expectedError: "restaurant not found",
		},
		{
			name: "repository update error",
			request: &dto.UpdateRestaurantRequest{
				Name: &nameNew,
			},
			expected: &dto.RestaurantResponse{
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
				Status:      restaurant.StatusActive.String(),
				Currency:    "USD",
			},
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)

			userId := uuid.New()

			if testCase.expectedError != "" {
				mockRepo.On("Update", mock.Anything, &dto.UpdateRestaurantData{
					Request: testCase.request,
					ID:      userId,
				}).Return(nil, errors.New(testCase.expectedError))
			} else {
				mockRepo.On("Update", mock.Anything, &dto.UpdateRestaurantData{
					Request: testCase.request,
					ID:      userId,
				}).Return(testCase.expected, nil)
			}

			service := NewRestaurantService(mockRepo)
			result, err := service.Update(t.Context(), userId, testCase.request)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testCase.expected, result)
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
		expected      []*dto.RestaurantResponse
		expectedError string
	}{
		{
			name: "successful retrieval with restaurants",
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
			name:          "successful retrieval with empty result",
			expected:      []*dto.RestaurantResponse{},
			expectedError: "",
		},
		{
			name:          "repository error",
			expected:      nil,
			expectedError: "database error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockRepo := new(MockRestaurantRepository)

			if testCase.expectedError != "" {
				mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New(testCase.expectedError))
			} else {
				mockRepo.On("GetAll", mock.Anything).Return(testCase.expected, nil)
			}

			service := NewRestaurantService(mockRepo)
			result, err := service.GetAll(t.Context())

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
