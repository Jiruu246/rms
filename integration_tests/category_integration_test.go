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
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const categoryAPIBase = "/api/categories"

type CategoryTestSuite struct {
	IntegrationTestSuite
}

func (s *CategoryTestSuite) SetupTest() {
	// s.IntegrationTestSuite.SetupTest()
	count := 0
	fmt.Printf("%d", count)
}

// TestCategoryTestSuite runs the category test suite
func TestCategoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}

// TestCategoryAPI tests the category API endpoints
func (s *CategoryTestSuite) TestCreateCategory() {
	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateCategory",
			body: dto.CreateCategoryRequest{
				Name:        "Test Category",
				Description: "A test category description",
				RestaurantID: func() uuid.UUID {
					restaurant, err := SetupRestaurant(s.client, s.T().Context())
					s.Require().NoError(err)
					return restaurant.ID
				}(),
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Category]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Category", response.Data.Name)
				s.Equal("A test category description", response.Data.Description)
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

			req := httptest.NewRequest(http.MethodPost, categoryAPIBase, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *CategoryTestSuite) TestGetCategory() {
	initialCategory1, err := SetupCategory(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = initialCategory1.Update().
		SetName("Initial Category 1").
		SetDescription("Initial Description 1").
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = SetupCategory(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetCategoryByID_NotFound",
			url:      path.Join(categoryAPIBase, uuid.New().String()),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetCategoryByID_InvalidUUID",
			url:      path.Join(categoryAPIBase, "invalid-uuid"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetCategoryByID_Success",
			url:      path.Join(categoryAPIBase, initialCategory1.ID.String()),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Category]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialCategory1.ID, response.Data.ID)
				s.Equal("Initial Category 1", response.Data.Name)
				s.Equal("Initial Description 1", response.Data.Description)
				s.Equal(initialCategory1.DisplayOrder, response.Data.DisplayOrder)
				s.Equal(initialCategory1.IsActive, response.Data.IsActive)
			},
		},
		{
			testName: "GetAllCategories",
			url:      categoryAPIBase,
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[[]dto.Category]
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

func (s *CategoryTestSuite) TestUpdateCategory() {
	initialCategory1, err := SetupCategory(s.client, s.T().Context())
	s.Require().NoError(err)

	_, err = initialCategory1.Update().
		SetName("Initial Category 1").
		SetDescription("Initial Description 1").
		Save(s.T().Context())
	s.Require().NoError(err)

	_, err = SetupCategory(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateCategory_Partial",
			url:      path.Join(categoryAPIBase, initialCategory1.ID.String()),
			body: dto.UpdateCategoryRequest{
				Name:        ptr("Updated Category"),
				Description: ptr("Updated description"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedCategory utils.APIResponse[dto.Category]
				err := json.Unmarshal(w.Body.Bytes(), &updatedCategory)
				s.Require().NoError(err)
				s.Equal(initialCategory1.ID, updatedCategory.Data.ID)
				s.Equal("Updated Category", updatedCategory.Data.Name)
				s.Equal("Updated description", updatedCategory.Data.Description)
				s.Equal(initialCategory1.DisplayOrder, updatedCategory.Data.DisplayOrder)
				s.Equal(initialCategory1.IsActive, updatedCategory.Data.IsActive)
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

func (s *CategoryTestSuite) TestDeleteCategory() {
	initialCategory, err := SetupCategory(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteCategory",
			url:      path.Join(categoryAPIBase, initialCategory.ID.String()),
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

// TestCategoryValidation tests input validation
func (s *CategoryTestSuite) TestCategoryValidation() {
	tests := []struct {
		testName string
		method   string
		url      string
		body     string
		expected int
	}{
		{
			testName: "CreateCategory_InvalidData_EmptyName",
			method:   http.MethodPost,
			url:      "/api/categories",
			body:     `{"Name": "", "Description": "Valid description"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateCategory_MalformedJSON",
			method:   http.MethodPost,
			url:      "/api/categories",
			body:     "{invalid json}",
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}
