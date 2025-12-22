package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const restaurantAPIBase = "/api/restaurants"

type RestaurantTestSuite struct {
	IntegrationTestSuite
}

func TestRestaurantTestSuite(t *testing.T) {
	suite.Run(t, new(RestaurantTestSuite))
}

// TestRestaurantAPI tests the restaurant API endpoints
func (s *RestaurantTestSuite) TestCreateRestaurant() {
	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateRestaurant",
			body: dto.CreateRestaurantRequest{
				Name:        "Test Restaurant",
				Description: "A test restaurant description",
				Phone:       "+1234567890",
				Email:       "test@restaurant.com",
				Address:     "123 Test Street",
				City:        "Test City",
				State:       "Test State",
				ZipCode:     "12345",
				Country:     "Test Country",
				Currency:    "USD",
				Status:      "active",
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.RestaurantResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Restaurant", response.Data.Name)
				s.Equal("A test restaurant description", response.Data.Description)
				s.Equal("+1234567890", response.Data.Phone)
				s.Equal("test@restaurant.com", response.Data.Email)
				s.Equal("123 Test Street", response.Data.Address)
				s.Equal("Test City", response.Data.City)
				s.Equal("Test State", response.Data.State)
				s.Equal("12345", response.Data.ZipCode)
				s.Equal("Test Country", response.Data.Country)
				s.Equal("USD", response.Data.Currency)
				s.Equal("active", response.Data.Status)
				s.NotEqual(uuid.Nil, response.Data.ID)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				s.Require().NoError(err)
			}

			req := httptest.NewRequest(http.MethodPost, restaurantAPIBase, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user, err := SetupUser(s.client, s.T().Context())
			s.Require().NoError(err)

			mockMiddlewares := DefaultMiddleware()
			mockMiddlewares.JWTMiddleware = func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: user.ID})
					c.Next()
				}
			}
			server := s.CreateServerWithMiddleware(mockMiddlewares)
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *RestaurantTestSuite) TestGetRestaurant() {
	initialRestaurant1, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)
	initialRestaurant1, err = s.client.Restaurant.UpdateOne(initialRestaurant1).
		SetName("Initial Restaurant 1").
		SetDescription("Initial Description 1").
		SetPhone("+1111111111").
		SetEmail("restaurant1@test.com").
		SetAddress("111 Initial Street").
		SetCity("Initial City 1").
		SetState("Initial State 1").
		SetZipCode("11111").
		SetCountry("Initial Country 1").
		SetCurrency("USD").
		SetStatus(restaurant.StatusActive).
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetRestaurantByID_NotFound",
			url:      path.Join(restaurantAPIBase, uuid.New().String()),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetRestaurantByID_InvalidUUID",
			url:      path.Join(restaurantAPIBase, "invalid-uuid"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetRestaurantByID_Success",
			url:      path.Join(restaurantAPIBase, initialRestaurant1.ID.String()),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.RestaurantResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialRestaurant1.ID, response.Data.ID)
				s.Equal("Initial Restaurant 1", response.Data.Name)
				s.Equal("Initial Description 1", response.Data.Description)
				s.Equal("+1111111111", response.Data.Phone)
				s.Equal("restaurant1@test.com", response.Data.Email)
				s.Equal("111 Initial Street", response.Data.Address)
				s.Equal("Initial City 1", response.Data.City)
				s.Equal("Initial State 1", response.Data.State)
				s.Equal("11111", response.Data.ZipCode)
				s.Equal("Initial Country 1", response.Data.Country)
				s.Equal("USD", response.Data.Currency)
				s.Equal(restaurant.StatusActive.String(), response.Data.Status)
			},
		},
		{
			testName: "GetAllRestaurants",
			url:      restaurantAPIBase,
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[[]dto.RestaurantResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.True(response.Success)
				s.True(len(response.Data) >= 2)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *RestaurantTestSuite) TestUpdateRestaurant() {
	initialRestaurant1, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)
	initialRestaurant1, err = s.client.Restaurant.UpdateOne(initialRestaurant1).
		SetName("Initial Restaurant 1").
		SetDescription("Initial Description 1").
		SetPhone("+1111111111").
		SetEmail("restaurant1@test.com").
		SetAddress("111 Initial Street").
		SetCity("Initial City 1").
		SetState("Initial State 1").
		SetZipCode("11111").
		SetCountry("Initial Country 1").
		SetCurrency("USD").
		SetStatus(restaurant.StatusActive).
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateRestaurant_Partial",
			url:      path.Join(restaurantAPIBase, initialRestaurant1.ID.String()),
			body: dto.UpdateRestaurantRequest{
				Name:        ptrString("Updated Restaurant"),
				Description: ptrString("Updated description"),
				Phone:       ptrString("+9999999999"),
				Email:       ptrString("updated@restaurant.com"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedRestaurant utils.APIResponse[dto.RestaurantResponse]
				err := json.Unmarshal(w.Body.Bytes(), &updatedRestaurant)
				s.Require().NoError(err)
				s.Equal(initialRestaurant1.ID, updatedRestaurant.Data.ID)
				s.Equal("Updated Restaurant", updatedRestaurant.Data.Name)
				s.Equal("Updated description", updatedRestaurant.Data.Description)
				s.Equal("+9999999999", updatedRestaurant.Data.Phone)
				s.Equal("updated@restaurant.com", updatedRestaurant.Data.Email)
				// Check that unchanged fields remain the same
				s.Equal("111 Initial Street", updatedRestaurant.Data.Address)
				s.Equal("Initial City 1", updatedRestaurant.Data.City)
				s.Equal("Initial State 1", updatedRestaurant.Data.State)
				s.Equal("11111", updatedRestaurant.Data.ZipCode)
				s.Equal("Initial Country 1", updatedRestaurant.Data.Country)
				s.Equal("USD", updatedRestaurant.Data.Currency)
				s.Equal("active", updatedRestaurant.Data.Status)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				s.Require().NoError(err)
			}

			req := httptest.NewRequest(http.MethodPut, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *RestaurantTestSuite) TestDeleteRestaurant() {
	initialRestaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteRestaurant",
			url:      path.Join(restaurantAPIBase, initialRestaurant.ID.String()),
			expected: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}

// TestRestaurantValidation tests input validation
func (s *RestaurantTestSuite) TestRestaurantValidation() {
	tests := []struct {
		testName string
		method   string
		url      string
		body     string
		expected int
	}{
		{
			testName: "CreateRestaurant_InvalidData_EmptyName",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "", "phone": "+1234567890", "email": "test@test.com", "address": "123 Test St", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": "USD"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_InvalidData_EmptyPhone",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "Test Restaurant", "phone": "", "email": "test@test.com", "address": "123 Test St", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": "USD"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_InvalidData_InvalidEmail",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "Test Restaurant", "phone": "+1234567890", "email": "invalid-email", "address": "123 Test St", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": "USD"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_InvalidData_EmptyAddress",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "Test Restaurant", "phone": "+1234567890", "email": "test@test.com", "address": "", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": "USD"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_InvalidData_EmptyCurrency",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "Test Restaurant", "phone": "+1234567890", "email": "test@test.com", "address": "123 Test St", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": ""}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_InvalidData_InvalidStatus",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     `{"name": "Test Restaurant", "phone": "+1234567890", "email": "test@test.com", "address": "123 Test St", "city": "Test", "state": "Test", "zip_code": "12345", "country": "Test", "currency": "USD", "status": "invalid_status"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateRestaurant_MalformedJSON",
			method:   http.MethodPost,
			url:      "/api/restaurants",
			body:     "{invalid json}",
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			user, err := SetupUser(s.client, s.T().Context())
			s.Require().NoError(err)
			mockMiddlewares := DefaultMiddleware()
			mockMiddlewares.JWTMiddleware = func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: user.ID})
					c.Next()
				}
			}
			server := s.CreateServerWithMiddleware(mockMiddlewares)
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}
