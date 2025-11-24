package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const userAPIBase = "/api/users"

type UserTestSuite struct {
	IntegrationTestSuite
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

// TestUserAPI tests the user API endpoints
func (s *UserTestSuite) TestCreateUser() {
	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateUser",
			body: dto.RegisterUserRequest{
				Name:     "Test User",
				Email:    "testuser@example.com",
				Password: "securepassword",
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.User]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test User", response.Data.Name)
				s.Equal("testuser@example.com", response.Data.Email)
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

			req := httptest.NewRequest(http.MethodPost, path.Join(userAPIBase, "register"), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *UserTestSuite) TestGetUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("Initial User").
		SetEmail("initialuser@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName          string
		url               string
		mockJWTMiddleware func(secretKey []byte) gin.HandlerFunc
		expected          int
		validate          func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetUserByID_NotFound",
			url:      path.Join(userAPIBase, "profile"),
			mockJWTMiddleware: func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: uuid.New()})
					c.Next()
				}
			},
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetUserByID_Success",
			url:      path.Join(userAPIBase, "profile"),
			mockJWTMiddleware: func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: initialUser.ID})
					c.Next()
				}
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.User]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialUser.ID, response.Data.ID)
				s.Equal("Initial User", response.Data.Name)
				s.Equal("initialuser@example.com", response.Data.Email)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockMiddlewares := DefaultMiddleware()
			mockMiddlewares.JWTMiddleware = tt.mockJWTMiddleware
			server := s.CreateServerWithMiddleware(mockMiddlewares)
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *UserTestSuite) TestUpdateUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("Initial User").
		SetEmail("initialuser2@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName          string
		url               string
		body              dto.UpdateUserRequest
		mockJWTMiddleware func(secretKey []byte) gin.HandlerFunc
		expected          int
		validate          func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateUser",
			url:      path.Join(userAPIBase, "profile"),
			body: dto.UpdateUserRequest{
				Name:  ptr("Updated User"),
				Email: ptr("updateduser@example.com"),
			},
			mockJWTMiddleware: func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: initialUser.ID})
					c.Next()
				}
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedUser utils.APIResponse[dto.User]
				err := json.Unmarshal(w.Body.Bytes(), &updatedUser)
				s.Require().NoError(err)
				s.Equal(initialUser.ID, updatedUser.Data.ID)
				s.Equal("Updated User", updatedUser.Data.Name)
				s.Equal("updateduser@example.com", updatedUser.Data.Email)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			body, err := json.Marshal(tt.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPut, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockMiddlewares := DefaultMiddleware()
			mockMiddlewares.JWTMiddleware = tt.mockJWTMiddleware
			server := s.CreateServerWithMiddleware(mockMiddlewares)
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *UserTestSuite) TestDeleteUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("User To Delete").
		SetEmail("deleteuser@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName          string
		url               string
		mockJWTMiddleware func(secretKey []byte) gin.HandlerFunc
		expected          int
	}{
		{
			testName: "DeleteUser",
			url:      path.Join(userAPIBase, "profile"),
			mockJWTMiddleware: func(secretKey []byte) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Set("claims", utils.JWTClaims{UserID: initialUser.ID})
					c.Next()
				}
			},
			expected: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			mockMiddlewares := DefaultMiddleware()
			mockMiddlewares.JWTMiddleware = tt.mockJWTMiddleware
			server := s.CreateServerWithMiddleware(mockMiddlewares)
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}

func (s *UserTestSuite) TestUserValidation() {
	tests := []struct {
		testName string
		url      string
		body     dto.RegisterUserRequest
		expected int
	}{
		{
			testName: "CreateUser_InvalidData_EmptyName",
			url:      path.Join(userAPIBase, "register"),
			body:     dto.RegisterUserRequest{Name: "", Email: "validemail@example.com"},
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			body, err := json.Marshal(tt.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServerWithMiddleware(DefaultMiddleware())
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}
