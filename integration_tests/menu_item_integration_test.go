package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/stretchr/testify/suite"
)

const menuItemAPIBase = "/api/menu-items"

type MenuItemTestSuite struct {
	IntegrationTestSuite
}

func (s *MenuItemTestSuite) SetupTest() {
	count := 0
	fmt.Printf("%d", count)
}

func TestMenuItemTestSuite(t *testing.T) {
	suite.Run(t, new(MenuItemTestSuite))
}

func (s *MenuItemTestSuite) TestCreateMenuItem() {
	restaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName       string
		body           any
		expectedStatus int
		validate       func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateMenuItem",
			body: dto.CreateMenuItemRequest{
				Name:         "Test Menu Item",
				Description:  "A test menu item description",
				Price:        9.99,
				RestaurantID: restaurant.ID,
			},
			expectedStatus: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.MenuItemResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Menu Item", response.Data.Name)
				s.Equal("A test menu item description", response.Data.Description)
				s.Equal(9.99, response.Data.Price)
				s.Equal(restaurant.ID, response.Data.RestaurantID)
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

			req := httptest.NewRequest(http.MethodPost, menuItemAPIBase, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expectedStatus, w.Code)

			tt.validate(w)
		})
	}
}

func (s *MenuItemTestSuite) TestGetMenuItem() {
	initialMenuItem, err := CreateMenuItem(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = initialMenuItem.Update().
		SetName("Initial Menu Item").
		SetDescription("Initial Description").
		SetPrice(19.99).
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = CreateMenuItem(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetMenuItemByID_NotFound",
			url:      path.Join(menuItemAPIBase, "999999"),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetMenuItemByID_InvalidID",
			url:      path.Join(menuItemAPIBase, "invalid-id"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetMenuItemByID_Success",
			url:      path.Join(menuItemAPIBase, fmt.Sprintf("%d", initialMenuItem.ID)),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.MenuItemResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialMenuItem.ID, response.Data.ID)
				s.Equal("Initial Menu Item", response.Data.Name)
				s.Equal("Initial Description", response.Data.Description)
				s.Equal(19.99, response.Data.Price)
			},
		},
		{
			testName: "GetAllMenuItems",
			url:      menuItemAPIBase,
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[[]dto.MenuItemResponse]
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

func (s *MenuItemTestSuite) TestUpdateMenuItem() {
	initialMenuItem, err := CreateMenuItem(s.client, s.T().Context())
	s.Require().NoError(err)

	_, err = initialMenuItem.Update().
		SetName("Initial Menu Item").
		SetDescription("Initial Description").
		SetPrice(19.99).
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = CreateMenuItem(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateMenuItem_Partial",
			url:      path.Join(menuItemAPIBase, fmt.Sprintf("%d", initialMenuItem.ID)),
			body: dto.UpdateMenuItemRequest{
				Name:        ptr("Updated Menu Item"),
				Description: ptr("Updated description"),
				Price:       ptrFloat(29.99),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedMenuItem utils.APIResponse[dto.MenuItemResponse]
				err := json.Unmarshal(w.Body.Bytes(), &updatedMenuItem)
				s.Require().NoError(err)
				s.Equal(initialMenuItem.ID, updatedMenuItem.Data.ID)
				s.Equal("Updated Menu Item", updatedMenuItem.Data.Name)
				s.Equal("Updated description", updatedMenuItem.Data.Description)
				s.Equal(29.99, updatedMenuItem.Data.Price)
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

func (s *MenuItemTestSuite) TestDeleteMenuItem() {
	initialMenuItem, err := CreateMenuItem(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteMenuItem",
			url:      path.Join(menuItemAPIBase, fmt.Sprintf("%d", initialMenuItem.ID)),
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

func ptrFloat(f float64) *float64 {
	return &f
}
