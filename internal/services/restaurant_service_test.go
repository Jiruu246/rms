package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.Restaurant, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.Restaurant, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRestaurantRepository) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]*dto.Restaurant, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) GetAuthorizationResource(ctx context.Context, id uuid.UUID) (authz.Resource, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(authz.Resource), args.Error(1)
}

func (m *MockRestaurantRepository) SetImage(ctx context.Context, data *dto.SetRestaurantImageData) (*dto.Restaurant, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) ClearImage(ctx context.Context, data *dto.ClearRestaurantImageData) (*dto.Restaurant, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Restaurant), args.Error(1)
}

type MockMediaService struct {
	mock.Mock
}

func (m *MockMediaService) CreateUpload(ctx context.Context, actor authz.Actor, purpose mediaupload.Purpose) (*dto.CreateUploadResult, error) {
	args := m.Called(ctx, actor, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CreateUploadResult), args.Error(1)
}

func (m *MockMediaService) ConsumeUpload(ctx context.Context, actor authz.Actor, uploadID uuid.UUID, expectedPurpose mediaupload.Purpose) (*dto.MediaAsset, error) {
	args := m.Called(ctx, actor, uploadID, expectedPurpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.MediaAsset), args.Error(1)
}

func (m *MockMediaService) DeleteMedia(ctx context.Context, actor authz.Actor, mediaID uuid.UUID) error {
	args := m.Called(ctx, actor, mediaID)
	return args.Error(0)
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
		expected      *dto.Restaurant
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
					Status:         restaurant.StatusActive.String(),
					OperatingHours: operatingHours,
					Currency:       "USD",
				},
				UserID: uuid1,
			},
			expected: &dto.Restaurant{
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
			expected: &dto.Restaurant{
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
			expected: &dto.Restaurant{
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
			expected: &dto.Restaurant{
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
			expected: &dto.Restaurant{
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

			mockMediaService := new(MockMediaService)
			service := NewRestaurantService(mockRepo, mockMediaService)
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
				expectedRestaurant := &dto.Restaurant{
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

			mockMediaService := new(MockMediaService)
			service := NewRestaurantService(mockRepo, mockMediaService)
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
		expected      *dto.Restaurant
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
					Return(&dto.Restaurant{
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
			expected: &dto.Restaurant{
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
					Return(&dto.Restaurant{
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
			expected: &dto.Restaurant{
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

			mockMediaService := new(MockMediaService)
			service := NewRestaurantService(mockRepo, mockMediaService)
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

			mockMediaService := new(MockMediaService)
			service := NewRestaurantService(mockRepo, mockMediaService)
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
		expected      []*dto.Restaurant
		expectedError string
	}{
		{
			name:  "scoped to actor's own restaurants",
			actor: authz.Actor{UserID: ownerID, Role: "owner"},
			expected: []*dto.Restaurant{
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
			expected: []*dto.Restaurant{
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
			expected:      []*dto.Restaurant{},
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

			mockMediaService := new(MockMediaService)
			service := NewRestaurantService(mockRepo, mockMediaService)
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

// The following cover the Restaurant image-slot subresource described in
// documentation/MediaUploadFramework.md: Create never touches MediaService,
// and the general Update never touches an image slot — both are handled by
// UpdateImage/ClearImage, which live on their own failure boundary
// (PUT/DELETE /restaurants/{id}/images/{slot}).

func TestRestaurantService_CreateImageUpload_Succeeds(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	expected := &dto.CreateUploadResult{UploadID: uuid.New()}

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockMediaService.On("CreateUpload", mock.Anything, owner, mediaupload.PurposeRestaurantLogo).
		Return(expected, nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.CreateImageUpload(t.Context(), owner, restaurantID, dto.RestaurantImageSlotLogo)

	assert.NoError(t, err)
	assert.Same(t, expected, result)
	mockRepo.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_CreateImageUpload_Forbidden(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	stranger := authz.Actor{UserID: uuid.New()}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.CreateImageUpload(t.Context(), stranger, restaurantID, dto.RestaurantImageSlotCover)

	assert.ErrorIs(t, err, apperr.ErrForbidden)
	assert.Nil(t, result)
	mockMediaService.AssertNotCalled(t, "CreateUpload")
	mockRepo.AssertExpectations(t)
}

func TestRestaurantService_UpdateImage_Succeeds(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	uploadID := uuid.New()
	newAssetID := uuid.New()

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockMediaService.On("ConsumeUpload", mock.Anything, owner, uploadID, mediaupload.PurposeRestaurantLogo).
		Return(&dto.MediaAsset{ID: newAssetID}, nil)
	mockRepo.On("SetImage", mock.Anything, &dto.SetRestaurantImageData{
		RestaurantID: restaurantID,
		Slot:         dto.RestaurantImageSlotLogo,
		MediaAssetID: newAssetID,
	}).Return(&dto.Restaurant{ID: restaurantID, LogoMediaAssetID: &newAssetID}, nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.UpdateImage(t.Context(), owner, restaurantID, dto.RestaurantImageSlotLogo, uploadID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
	// Whatever was previously in the slot (if anything) is left completely
	// alone — no read of the current row, no compensating delete.
	mockRepo.AssertNotCalled(t, "GetByID")
	mockMediaService.AssertNotCalled(t, "DeleteMedia")
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_UpdateImage_ConsumeFails(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	uploadID := uuid.New()

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockMediaService.On("ConsumeUpload", mock.Anything, owner, uploadID, mediaupload.PurposeRestaurantCoverImage).
		Return(nil, apperr.Invalid("uploaded object content type %q does not match expected %q", "application/pdf", "image/jpeg"))

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.UpdateImage(t.Context(), owner, restaurantID, dto.RestaurantImageSlotCover, uploadID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, apperr.ErrInvalid)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "SetImage")
	mockRepo.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_UpdateImage_SetImageFailsCompensatesNewAsset(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	uploadID := uuid.New()
	newAssetID := uuid.New()

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockMediaService.On("ConsumeUpload", mock.Anything, owner, uploadID, mediaupload.PurposeRestaurantLogo).
		Return(&dto.MediaAsset{ID: newAssetID}, nil)
	mockRepo.On("SetImage", mock.Anything, mock.Anything).
		Return(nil, errors.New("database error"))
	mockMediaService.On("DeleteMedia", mock.Anything, owner, newAssetID).Return(nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.UpdateImage(t.Context(), owner, restaurantID, dto.RestaurantImageSlotLogo, uploadID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_UpdateImage_ConsumeConflictIsSurfaced(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	uploadID := uuid.New()
	conflict := apperr.Conflict("media upload %s has already been consumed", uploadID)

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockMediaService.On("ConsumeUpload", mock.Anything, owner, uploadID, mediaupload.PurposeRestaurantLogo).
		Return(nil, conflict)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.UpdateImage(t.Context(), owner, restaurantID, dto.RestaurantImageSlotLogo, uploadID)

	assert.ErrorIs(t, err, apperr.ErrConflict)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "SetImage")
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_ClearImage_Succeeds(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	owner := authz.Actor{UserID: ownerID}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)
	mockRepo.On("ClearImage", mock.Anything, &dto.ClearRestaurantImageData{
		RestaurantID: restaurantID,
		Slot:         dto.RestaurantImageSlotCover,
	}).Return(&dto.Restaurant{ID: restaurantID}, nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.ClearImage(t.Context(), owner, restaurantID, dto.RestaurantImageSlotCover)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
	mockMediaService.AssertNotCalled(t, "DeleteMedia")
	mockMediaService.AssertExpectations(t)
}

func TestRestaurantService_ClearImage_Forbidden(t *testing.T) {
	restaurantID := uuid.New()
	ownerID := uuid.New()
	stranger := authz.Actor{UserID: uuid.New()}
	resource := authz.Resource{ID: restaurantID, RestaurantID: restaurantID, OwnerUserID: ownerID}

	mockRepo := new(MockRestaurantRepository)
	mockMediaService := new(MockMediaService)

	mockRepo.On("GetAuthorizationResource", mock.Anything, restaurantID).Return(resource, nil)

	service := NewRestaurantService(mockRepo, mockMediaService)
	result, err := service.ClearImage(t.Context(), stranger, restaurantID, dto.RestaurantImageSlotCover)

	assert.ErrorIs(t, err, apperr.ErrForbidden)
	assert.Nil(t, result)
	mockRepo.AssertNotCalled(t, "ClearImage")
	mockRepo.AssertExpectations(t)
}
