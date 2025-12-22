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

const modifierOptionAPIBase = "/api/modifiers/options"

type ModifierOptionTestSuite struct {
	IntegrationTestSuite
}

func (s *ModifierOptionTestSuite) SetupTest() {
	count := 0
	fmt.Printf("%d", count)
}

func TestModifierOptionTestSuite(t *testing.T) {
	suite.Run(t, new(ModifierOptionTestSuite))
}

func (s *ModifierOptionTestSuite) TestCreateModifierOption() {
	modifier, err := CreateModifier(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateModifierOption",
			body: dto.CreateModifierOptionRequest{
				Name:       "Test Modifier Option",
				ModifierID: modifier.ID,
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.ModifierOptionResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal("Test Modifier Option", response.Data.Name)
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

			req := httptest.NewRequest(http.MethodPost, modifierOptionAPIBase, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *ModifierOptionTestSuite) TestGetModifierOption() {
	initialOption, err := CreateModifierOption(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = initialOption.Update().
		SetName("Initial Modifier Option").
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetModifierOptionByID_NotFound",
			url:      path.Join(modifierOptionAPIBase, uuid.New().String()),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetModifierOptionByID_InvalidUUID",
			url:      path.Join(modifierOptionAPIBase, "invalid-uuid"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetModifierOptionByID_Success",
			url:      path.Join(modifierOptionAPIBase, initialOption.ID.String()),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.ModifierOptionResponse]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(initialOption.ID, response.Data.ID)
				s.Equal("Initial Modifier Option", response.Data.Name)
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

func (s *ModifierOptionTestSuite) TestUpdateModifierOption() {
	initialOption, err := CreateModifierOption(s.client, s.T().Context())
	s.Require().NoError(err)

	_, err = initialOption.Update().
		SetName("Initial Modifier Option").
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
			testName: "UpdateModifierOption_Partial",
			url:      path.Join(modifierOptionAPIBase, initialOption.ID.String()),
			body: dto.UpdateModifierOptionRequest{
				Name: ptr("Updated Modifier Option"),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var updatedOption utils.APIResponse[dto.ModifierOptionResponse]
				err := json.Unmarshal(w.Body.Bytes(), &updatedOption)
				s.Require().NoError(err)
				s.Equal(initialOption.ID, updatedOption.Data.ID)
				s.Equal("Updated Modifier Option", updatedOption.Data.Name)
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

func (s *ModifierOptionTestSuite) TestDeleteModifierOption() {
	initialOption, err := CreateModifierOption(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteModifierOption",
			url:      path.Join(modifierOptionAPIBase, initialOption.ID.String()),
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
