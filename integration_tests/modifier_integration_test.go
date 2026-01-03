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

const modifierAPIBase = "/api/modifiers"

type ModifierTestSuite struct {
	IntegrationTestSuite
}

func (s *ModifierTestSuite) SetupTest() {
	count := 0
	fmt.Printf("%d", count)
}

func TestModifierTestSuite(t *testing.T) {
	suite.Run(t, new(ModifierTestSuite))
}

func (s *ModifierTestSuite) TestCreateModifier() {
	restaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateModifier",
			body: dto.CreateModifierRequest{
				Name:         "Test Modifier",
				RestaurantID: restaurant.ID,
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Modifier]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Modifier", response.Data.Name)
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

			req := httptest.NewRequest(http.MethodPost, modifierAPIBase, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *ModifierTestSuite) TestGetModifier() {
	initialModifier, err := CreateModifier(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = initialModifier.Update().
		SetName("Initial Modifier").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetModifierByID_NotFound",
			url:      path.Join(modifierAPIBase, uuid.New().String()),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetModifierByID_InvalidUUID",
			url:      path.Join(modifierAPIBase, "invalid-uuid"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetModifierByID_Success",
			url:      path.Join(modifierAPIBase, initialModifier.ID.String()),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Modifier]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialModifier.ID, response.Data.ID)
				s.Equal("Initial Modifier", response.Data.Name)
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

func (s *ModifierTestSuite) TestUpdateModifier() {
	initialModifier, err := CreateModifier(s.client, s.T().Context())
	s.Require().NoError(err)

	_, err = initialModifier.Update().
		SetName("Initial Modifier").
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
			testName: "UpdateModifier_Partial",
			url:      path.Join(modifierAPIBase, initialModifier.ID.String()),
			body: dto.UpdateModifierRequest{
				Name: ptr("Updated Modifier"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedModifier utils.APIResponse[dto.Modifier]
				err := json.Unmarshal(w.Body.Bytes(), &updatedModifier)
				s.Require().NoError(err)
				s.Equal(initialModifier.ID, updatedModifier.Data.ID)
				s.Equal("Updated Modifier", updatedModifier.Data.Name)
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

			req := httptest.NewRequest(http.MethodPatch, tt.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *ModifierTestSuite) TestDeleteModifier() {
	initialModifier, err := CreateModifier(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteModifier",
			url:      path.Join(modifierAPIBase, initialModifier.ID.String()),
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
