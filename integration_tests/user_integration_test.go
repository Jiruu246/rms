package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userAPIBase = "/api/users"

// TestUserAPI tests the user API endpoints
func (s *IntegrationTestSuite) TestCreateUser() {
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
				var response utils.APIResponse[dto.UserProfileResponse]
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

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *IntegrationTestSuite) TestGetUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("Initial User").
		SetEmail("initialuser@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName   string
		url        string
		addContext func(req *http.Request) *http.Request
		expected   int
		validate   func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetUserByID_NotFound",
			url:      path.Join(userAPIBase, "profile"),
			addContext: func(req *http.Request) *http.Request {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = req
				c.Set("userID", uuid.New().String())
				return c.Request
			},
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetUserByID_InvalidUUID",
			url:      path.Join(userAPIBase, "profile"),
			addContext: func(req *http.Request) *http.Request {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = req
				c.Set("userID", "invalid-uuid")
				return c.Request
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetUserByID_Success",
			url:      path.Join(userAPIBase, "profile"),
			addContext: func(req *http.Request) *http.Request {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = req
				c.Set("userID", initialUser.ID.String())
				return c.Request
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.UserProfileResponse]
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

			req = tt.addContext(req)

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("Initial User").
		SetEmail("initialuser@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateUser",
			url:      path.Join(userAPIBase, initialUser.ID.String()),
			body: dto.UpdateUserRequest{
				Name:  ptr("Updated User"),
				Email: ptr("updateduser@example.com"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedUser utils.APIResponse[dto.UserProfileResponse]
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
			var body []byte
			var err error
			if tt.body != nil {
				body, err = json.Marshal(tt.body)
				s.Require().NoError(err)
			}

			req := httptest.NewRequest(http.MethodPut, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *IntegrationTestSuite) TestDeleteUser() {
	initialUser, err := s.client.Customer.Create().
		SetName("User To Delete").
		SetEmail("deleteuser@example.com").
		SetPasswordHash("someHash").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteUser",
			url:      path.Join(userAPIBase, initialUser.ID.String()),
			expected: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodDelete, tt.url, nil)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}

func (s *IntegrationTestSuite) TestUserValidation() {
	tests := []struct {
		testName string
		method   string
		url      string
		body     string
		expected int
	}{
		{
			testName: "CreateUser_InvalidData_EmptyName",
			method:   http.MethodPost,
			url:      userAPIBase,
			body:     `{"Name": "", "Email": "validemail@example.com"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateUser_MalformedJSON",
			method:   http.MethodPost,
			url:      userAPIBase,
			body:     "{invalid json}",
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}
