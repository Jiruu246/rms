package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/handler"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

const authAPIBase = "/api/auth"

type AuthTestSuite struct {
	IntegrationTestSuite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

// Helper function to find cookie by name
func (s *AuthTestSuite) findCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

// TestRegister tests the register endpoint
func (s *AuthTestSuite) TestRegister() {
	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "Register_Success",
			body: handler.RegisterUserSchema{
				Name:     "New User",
				Email:    "newuser@example.com",
				Password: "securepassword123",
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.User]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("New User", response.Data.Name)
				s.Equal("newuser@example.com", response.Data.Email)
				s.NotEmpty(response.Data.ID)
			},
		},
		{
			testName: "Register_MissingName",
			body: handler.RegisterUserSchema{
				Name:     "",
				Email:    "newuser@example.com",
				Password: "securepassword123",
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Register_MissingEmail",
			body: handler.RegisterUserSchema{
				Name:     "New User",
				Email:    "",
				Password: "securepassword123",
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Register_InvalidEmail",
			body: handler.RegisterUserSchema{
				Name:     "New User",
				Email:    "invalidemail",
				Password: "securepassword123",
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Register_MissingPassword",
			body: handler.RegisterUserSchema{
				Name:     "New User",
				Email:    "newuser@example.com",
				Password: "",
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
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

			req := httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "register"), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code, "Response body: %s", w.Body.String())

			tt.validate(w)
		})
	}
}

// TestLogin tests the login endpoint
func (s *AuthTestSuite) TestLogin() {
	// Create a user for login tests
	const Password = "password"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = user.Update().SetPasswordHash(string(hashedPassword)).Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "Login_Success",
			body: dto.LoginUserRequest{
				Email:    user.Email,
				Password: Password,
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.AccessToken]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.NotEmpty(response.Data.Token)
				s.False(response.Data.ExpiresAt.IsZero())

				// Verify refresh token cookie is set
				cookies := w.Result().Cookies()
				refreshTokenCookie := s.findCookieByName(cookies, "refresh_token")
				s.NotNil(refreshTokenCookie, "refresh_token cookie should be set")
				s.NotEmpty(refreshTokenCookie.Value)
				s.True(refreshTokenCookie.HttpOnly)
				s.Equal("/auth/refresh", refreshTokenCookie.Path)
			},
		},
		{
			testName: "Login_MissingEmail",
			body: dto.LoginUserRequest{
				Email:    "",
				Password: Password,
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Login_InvalidEmail",
			body: dto.LoginUserRequest{
				Email:    "invalidemail",
				Password: Password,
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Login_MissingPassword",
			body: dto.LoginUserRequest{
				Email:    user.Email,
				Password: "",
			},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Login_UserNotFound",
			body: dto.LoginUserRequest{
				Email:    "nonexistent@example.com",
				Password: "anypassword",
			},
			expected: http.StatusUnauthorized,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "Login_WrongPassword",
			body: dto.LoginUserRequest{
				Email:    user.Email,
				Password: "wrongpassword",
			},
			expected: http.StatusUnauthorized,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
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

			req := httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "login"), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code, "Response body: %s", w.Body.String())

			tt.validate(w)
		})
	}
}

// TestRefreshToken tests the refresh token endpoint
func (s *AuthTestSuite) TestRefreshToken() {
	// Create a user and refresh token
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)
	refreshTokenStr, err := utils.GenerateRefreshToken()
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenStr), bcrypt.DefaultCost)
	s.Require().NoError(err)
	refreshToken, err := s.client.RefreshToken.Create().
		SetToken(string(hashedRefreshToken)).
		SetUserID(user.ID).
		SetExpiresAt(time.Now().Add(7 * 24 * time.Hour)).
		Save(s.T().Context())
	s.Require().NoError(err)
	tokenViaCookie := refreshToken.ID.String() + ":" + refreshTokenStr

	tests := []struct {
		testName string
		cookies  []*http.Cookie
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "RefreshToken_Success_WithCookie",
			cookies:  []*http.Cookie{{Name: "refresh_token", Value: tokenViaCookie}},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.AccessToken]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.NotEmpty(response.Data.Token)
				s.False(response.Data.ExpiresAt.IsZero())
			},
		},
		{
			testName: "RefreshToken_MissingToken",
			cookies:  []*http.Cookie{},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
		{
			testName: "RefreshToken_InvalidToken",
			cookies:  []*http.Cookie{{Name: "refresh_token", Value: "invalid-token"}},
			expected: http.StatusUnauthorized,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "refresh"), nil)
			req.Header.Set("Content-Type", "application/json")

			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}

			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code, "Response body: %s", w.Body.String())

			tt.validate(w)
		})
	}
}

// TestLogout tests the logout endpoint
func (s *AuthTestSuite) TestLogout() {
	// Create a user and get refresh token via login
	user, err := SetupUser(s.client, s.T().Context())
	s.Require().NoError(err)
	refreshTokenStr, err := utils.GenerateRefreshToken()
	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshTokenStr), bcrypt.DefaultCost)
	s.Require().NoError(err)
	refreshToken, err := s.client.RefreshToken.Create().
		SetToken(string(hashedRefreshToken)).
		SetUserID(user.ID).
		SetExpiresAt(time.Now().Add(7 * 24 * time.Hour)).
		Save(s.T().Context())
	s.Require().NoError(err)
	tokenViaCookie := refreshToken.ID.String() + ":" + refreshTokenStr

	tests := []struct {
		testName string
		cookies  []*http.Cookie
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "Logout_Success",
			cookies:  []*http.Cookie{{Name: "refresh_token", Value: tokenViaCookie}},
			expected: http.StatusNoContent,
			validate: func(w *httptest.ResponseRecorder) {
				// Verify that the refresh token cookie is cleared
				cookies := w.Result().Cookies()
				clearedCookie := s.findCookieByName(cookies, "refresh_token")
				if clearedCookie != nil {
					s.Equal("", clearedCookie.Value)
					s.Equal(-1, clearedCookie.MaxAge)
				}
			},
		},
		{
			testName: "Logout_MissingToken",
			cookies:  []*http.Cookie{},
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {
				s.NotEmpty(w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "logout"), nil)
			req.Header.Set("Content-Type", "application/json")

			// Add cookies if provided
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}

			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code, "Response body: %s", w.Body.String())

			tt.validate(w)
		})
	}
}

// TestAuthFlow tests the complete authentication flow
func (s *AuthTestSuite) TestAuthFlow() {
	// Step 1: Register a new user
	registerReq := handler.RegisterUserSchema{
		Name:     "Integration Test User",
		Email:    "integrationtest@example.com",
		Password: "testpassword123",
	}

	registerBody, err := json.Marshal(registerReq)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "register"), bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server := s.CreateServer()
	server.Engine().ServeHTTP(w, req)
	s.Equal(http.StatusCreated, w.Code)

	var registerResp utils.APIResponse[dto.User]
	err = json.Unmarshal(w.Body.Bytes(), &registerResp)
	s.Require().NoError(err)
	s.Equal("Integration Test User", registerResp.Data.Name)
	s.Equal("integrationtest@example.com", registerResp.Data.Email)

	// Step 2: Login with the registered user
	loginReq := dto.LoginUserRequest{
		Email:    "integrationtest@example.com",
		Password: "testpassword123",
	}

	loginBody, err := json.Marshal(loginReq)
	s.Require().NoError(err)

	req = httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "login"), bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	server.Engine().ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var loginResp utils.APIResponse[dto.AccessToken]
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	s.Require().NoError(err)
	s.NotEmpty(loginResp.Data.Token)
	s.False(loginResp.Data.ExpiresAt.IsZero())

	// Extract refresh token from the login response
	// Note: If refresh token is not in the response, we may need to check cookies or headers
	// For now, we'll assume it's in the response body

	// Step 3: Test refresh token using cookie
	cookies := w.Result().Cookies()
	refreshTokenCookie := s.findCookieByName(cookies, "refresh_token")
	s.Require().NotNil(refreshTokenCookie, "refresh_token cookie should be set")
	s.NotEmpty(refreshTokenCookie.Value)

	// Test refresh token endpoint
	req = httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "refresh"), bytes.NewBuffer([]byte(`{"refresh_token":""}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(refreshTokenCookie)
	w = httptest.NewRecorder()

	server.Engine().ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	var refreshResp utils.APIResponse[dto.AccessToken]
	err = json.Unmarshal(w.Body.Bytes(), &refreshResp)
	s.Require().NoError(err)
	s.NotEmpty(refreshResp.Data.Token)
	s.False(refreshResp.Data.ExpiresAt.IsZero())

	// Step 4: Test logout
	req = httptest.NewRequest(http.MethodPost, path.Join(authAPIBase, "logout"), nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(refreshTokenCookie)
	w = httptest.NewRecorder()

	server.Engine().ServeHTTP(w, req)
	s.Equal(http.StatusNoContent, w.Code)

	// Verify that the refresh token cookie is cleared
	logoutCookies := w.Result().Cookies()
	clearedCookie := s.findCookieByName(logoutCookies, "refresh_token")
	if clearedCookie != nil {
		s.Equal("", clearedCookie.Value)
		s.Equal(-1, clearedCookie.MaxAge)
	}
}

// TestAuthValidation tests field validation for auth endpoints
func (s *AuthTestSuite) TestAuthValidation() {
	tests := []struct {
		testName string
		endpoint string
		body     any
		expected int
	}{
		{
			testName: "Register_InvalidEmailFormat",
			endpoint: path.Join(authAPIBase, "register"),
			body: handler.RegisterUserSchema{
				Name:     "Test User",
				Email:    "notanemail",
				Password: "password123",
			},
			expected: http.StatusBadRequest,
		},
		{
			testName: "Login_InvalidEmailFormat",
			endpoint: path.Join(authAPIBase, "login"),
			body: dto.LoginUserRequest{
				Email:    "notanemail",
				Password: "password123",
			},
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			body, err := json.Marshal(tt.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, tt.endpoint, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code, "Response body: %s", w.Body.String())
		})
	}
}
