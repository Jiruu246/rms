package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/google/uuid"
)

// TestCategoryAPI tests the category API endpoints
func (s *IntegrationTestSuite) TestCategoryAPI() {
	var createdCategory dto.CategoryResponse

	tests := []struct {
		testName string
		method   string
		url      string
		body     interface{}
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateCategory",
			method:   http.MethodPost,
			url:      "/api/categories",
			body: dto.CreateCategoryRequest{
				Name:        "Test Category",
				Description: "A test category description",
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Category", response.Name)
				s.Equal("A test category description", response.Description)
				s.NotEqual(uuid.Nil, response.ID)

				// Save the created category for later tests
				createdCategory = response
			},
		},
		{
			testName: "GetCategoryByID_NotFound",
			method:   http.MethodGet,
			url:      "/api/categories/" + uuid.New().String(),
			body:     nil,
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetCategoryByID_InvalidUUID",
			method:   http.MethodGet,
			url:      "/api/categories/invalid-uuid",
			body:     nil,
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "UpdateCategory",
			method:   http.MethodPut,
			url:      "/api/categories/" + createdCategory.ID.String(),
			body: dto.UpdateCategoryRequest{
				Name:        ptr("Updated Category"),
				Description: ptr("Updated description"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedCategory dto.CategoryResponse
				err := json.Unmarshal(w.Body.Bytes(), &updatedCategory)
				s.Require().NoError(err)
				s.Equal(createdCategory.ID, updatedCategory.ID)
				s.Equal("Updated Category", updatedCategory.Name)
				s.Equal("Updated description", updatedCategory.Description)
			},
		},
		{
			testName: "DeleteCategory",
			method:   http.MethodDelete,
			url:      "/api/categories/" + createdCategory.ID.String(),
			body:     nil,
			expected: http.StatusNoContent,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "ListCategories",
			method:   http.MethodGet,
			url:      "/api/categories",
			body:     nil,
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response struct {
					Categories []dto.CategoryResponse `json:"categories"`
					Total      int                    `json:"total"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.True(response.Total > 0)
				s.Len(response.Categories, response.Total)
			},
		},
		{
			testName: "ListCategoriesWithSearch",
			method:   http.MethodGet,
			url:      "/api/categories?search=Electronic",
			body:     nil,
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response struct {
					Categories []dto.CategoryResponse `json:"categories"`
					Total      int                    `json:"total"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(2, response.Total)
				s.Len(response.Categories, 2)
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

			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			if tt.validate != nil {
				tt.validate(w)
			}
		})
	}
}

// TestCategoryValidation tests input validation
func (s *IntegrationTestSuite) TestCategoryValidation() {
	tests := []struct {
		testName string
		body     string
		expected int
	}{
		{
			testName: "CreateCategory_InvalidData_EmptyName",
			body:     `{"Name": "", "Description": "Valid description"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateCategory_InvalidData_EmptyDescription",
			body:     `{"Name": "Valid Name", "Description": ""}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateCategory_MalformedJSON",
			body:     "{invalid json}",
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)
		})
	}
}

func ptr(s string) *string {
	return &s
}
