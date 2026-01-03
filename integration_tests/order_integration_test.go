package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent/order"
	"github.com/Jiruu246/rms/internal/handler"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

const orderAPIBase = "/api/orders"

type OrderTestSuite struct {
	IntegrationTestSuite
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}

func (s *OrderTestSuite) TestCreateOrder() {
	restaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	menuItem, err := CreateMenuItemForRestaurant(s.client, s.T().Context(), restaurant)
	s.Require().NoError(err)

	modifier, err := CreateModifierForItem(s.client, s.T().Context(), menuItem)
	s.Require().NoError(err)
	modifier, err = s.client.Modifier.UpdateOne(modifier).
		SetMax(3).
		Save(s.T().Context())
	s.Require().NoError(err)

	modifierOption1, err := CreateModifierOptionForModifier(s.client, s.T().Context(), modifier)
	s.Require().NoError(err)

	modifierOption2, err := CreateModifierOptionForModifier(s.client, s.T().Context(), modifier)
	s.Require().NoError(err)

	tests := []struct {
		testName string
		body     handler.CreateOrderSchema
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "CreateOrder_Success",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem.ID,
						Quantity:   2,
						ModifierOptions: []handler.ModifierOption{
							{
								ModifierID: modifierOption1.ID,
								Quantity:   1,
							},
							{
								ModifierID: modifierOption2.ID,
								Quantity:   2,
							},
						},
					},
				},
			},
			expected: http.StatusCreated,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Order]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(dto.OrderTypeDINE_IN, response.Data.OrderType)
				s.Equal(dto.OrderStatusOPEN, response.Data.OrderStatus)
				s.Equal(restaurant.ID, response.Data.RestaurantID)
				s.NotEqual(uuid.Nil, response.Data.ID)

				s.Require().Len(response.Data.OrderItems, 1)
				orderItem := response.Data.OrderItems[0]
				s.Equal(menuItem.ID, orderItem.MenuItemID)
				s.Equal(2, orderItem.Quantity)

				s.Require().Len(orderItem.ModifierOptions, 2)

				var modOpt1, modOpt2 *dto.OrderItemModifierOption
				for i := range orderItem.ModifierOptions {
					switch orderItem.ModifierOptions[i].ModifierOptionID {
					case modifierOption1.ID:
						modOpt1 = &orderItem.ModifierOptions[i]
					case modifierOption2.ID:
						modOpt2 = &orderItem.ModifierOptions[i]
					}
				}

				s.Require().NotNil(modOpt1, "Modifier option 1 not found in response")
				s.Require().NotNil(modOpt2, "Modifier option 2 not found in response")
				s.Equal(1, modOpt1.Quantity)
				s.Equal(2, modOpt2.Quantity)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			var body []byte
			var err error
			body, err = json.Marshal(tt.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/public/order", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.Require().NoError(err)
			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)
			s.Equal(tt.expected, w.Code)

			tt.validate(w)
		})
	}
}

func (s *OrderTestSuite) TestGetOrder() {
	order, err := SetupOrder(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetOrderByID_NotFound",
			url:      path.Join(orderAPIBase, uuid.New().String()),
			expected: http.StatusNotFound,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetOrderByID_InvalidUUID",
			url:      path.Join(orderAPIBase, "invalid-uuid"),
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetOrderByID_Success",
			url:      path.Join(orderAPIBase, order.ID.String()),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Order]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(order.ID, response.Data.ID)
				s.Equal(order.RestaurantID, response.Data.RestaurantID)
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

func (s *OrderTestSuite) TestGetOrders() {
	restaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)
	_, err = s.client.Order.Create().
		SetOrderType(order.OrderTypeDINE_IN).
		SetRestaurant(restaurant).
		Save(s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "GetOrdersByRestaurantID_Success",
			url:      orderAPIBase + "?restaurant_id=" + restaurant.ID.String(),
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[[]dto.Order]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.True(response.Success)
				s.True(len(response.Data) >= 1)
			},
		},
		{
			testName: "GetOrdersByRestaurantID_InvalidUUID",
			url:      orderAPIBase + "?restaurant_id=invalid-uuid",
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
		},
		{
			testName: "GetOrdersByRestaurantID_MissingParam",
			url:      orderAPIBase,
			expected: http.StatusBadRequest,
			validate: func(w *httptest.ResponseRecorder) {},
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

func (s *OrderTestSuite) TestUpdateOrder() {
	order, err := SetupOrder(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		body     any
		expected int
		validate func(*httptest.ResponseRecorder)
	}{
		{
			testName: "UpdateOrder_Success",
			url:      path.Join(orderAPIBase, order.ID.String()),
			body: dto.UpdateOrderRequest{
				OrderStatus: ptrString(string(dto.OrderStatusCOMPLETED)),
			},
			expected: http.StatusOK,
			validate: func(w *httptest.ResponseRecorder) {
				var response utils.APIResponse[dto.Order]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				s.Require().NoError(err)
				s.Equal(order.ID, response.Data.ID)
				s.Equal(dto.OrderStatusCOMPLETED, response.Data.OrderStatus)
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

func (s *OrderTestSuite) TestDeleteOrder() {
	order, err := SetupOrder(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		url      string
		expected int
	}{
		{
			testName: "DeleteOrder_Success",
			url:      path.Join(orderAPIBase, order.ID.String()),
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

func (s *OrderTestSuite) TestOrderValidation() {
	restaurant, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	tests := []struct {
		testName string
		method   string
		url      string
		body     string
		expected int
	}{
		{
			testName: "CreateOrder_InvalidData_EmptyOrderType",
			method:   http.MethodPost,
			url:      orderAPIBase,
			body:     `{"order_type": "", "restaurant_id": "` + restaurant.ID.String() + `"}`,
			expected: http.StatusBadRequest,
		},
		{
			testName: "CreateOrder_InvalidData_InvalidRestaurantID",
			method:   http.MethodPost,
			url:      orderAPIBase,
			body:     `{"order_type": "dine_in", "restaurant_id": "invalid-uuid"}`,
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

func (s *OrderTestSuite) TestCreateOrderItemValidations() {
	// Setup test data
	restaurant1, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	restaurant2, err := SetupRestaurant(s.client, s.T().Context())
	s.Require().NoError(err)

	// Available menu item in restaurant1
	menuItem1, err := CreateMenuItemForRestaurant(s.client, s.T().Context(), restaurant1)
	s.Require().NoError(err)

	// Unavailable menu item in restaurant1
	unAvailMenuItem, err := CreateMenuItemForRestaurant(s.client, s.T().Context(), restaurant1)
	s.Require().NoError(err)
	unAvailMenuItem, err = s.client.MenuItem.UpdateOne(unAvailMenuItem).
		SetIsAvailable(false).
		Save(s.T().Context())
	s.Require().NoError(err)

	// Menu item in restaurant2
	menuItem2, err := CreateMenuItemForRestaurant(s.client, s.T().Context(), restaurant2)
	s.Require().NoError(err)

	// Create modifiers for menuItem1
	requiredModifier, err := CreateModifierForItem(s.client, s.T().Context(), menuItem1)
	s.Require().NoError(err)
	requiredModifier, err = s.client.Modifier.UpdateOne(requiredModifier).
		SetRequired(true).
		SetMax(2).
		Save(s.T().Context())
	s.Require().NoError(err)

	optionalModifier, err := CreateModifierForItem(s.client, s.T().Context(), menuItem1)
	s.Require().NoError(err)
	optionalModifier, err = s.client.Modifier.UpdateOne(optionalModifier).
		SetRequired(false).
		SetMax(3).
		Save(s.T().Context())
	s.Require().NoError(err)

	// Create modifier options
	requiredOption1, err := CreateModifierOptionForModifier(s.client, s.T().Context(), requiredModifier)
	s.Require().NoError(err)

	requiredOption2, err := CreateModifierOptionForModifier(s.client, s.T().Context(), requiredModifier)
	s.Require().NoError(err)

	optionalOption1, err := CreateModifierOptionForModifier(s.client, s.T().Context(), optionalModifier)
	s.Require().NoError(err)

	// Unavailable modifier option
	unavailableOption, err := CreateModifierOptionForModifier(s.client, s.T().Context(), requiredModifier)
	s.Require().NoError(err)
	unavailableOption, err = s.client.ModifierOption.UpdateOne(unavailableOption).
		SetAvailable(false).
		Save(s.T().Context())
	s.Require().NoError(err)

	// Modifier option for menuItem2 (different restaurant)
	modifierForItem2, err := CreateModifierForItem(s.client, s.T().Context(), menuItem2)
	s.Require().NoError(err)
	optionForItem2, err := CreateModifierOptionForModifier(s.client, s.T().Context(), modifierForItem2)
	s.Require().NoError(err)

	nonExistentMenuItemID := int64(-1)
	nonExistentModifierOptionID := uuid.New()

	tests := []struct {
		testName string
		body     handler.CreateOrderSchema
		expected int
	}{
		// Menu item scenarios
		{
			testName: "EmptyOrderItems",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems:   []handler.OrderItemSchema{},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "DuplicateMenuItems",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1},
						},
					},
					{
						MenuItemID: menuItem1.ID,
						Quantity:   2,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption2.ID, Quantity: 1},
						},
					},
				},
			},
			expected: http.StatusCreated,
		},
		{
			testName: "MenuItemDoesNotExist",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: nonExistentMenuItemID,
						Quantity:   1,
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "MenuItemFromAnotherRestaurant",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem2.ID, // This belongs to restaurant2
						Quantity:   1,
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "MenuItemUnavailable",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: unAvailMenuItem.ID, // This is unavailable
						Quantity:   1,
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "ZeroQuantityItem",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   0,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1},
						},
					},
				},
			},
			expected: http.StatusBadRequest, // Should be invalid due to zero quantity
		},
		{
			testName: "NegativeQuantityItem",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   -1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1},
						},
					},
				},
			},
			expected: http.StatusBadRequest, // Should be invalid due to negative quantity
		},

		// Modifier option scenarios
		{
			testName: "ModifierOptionDoesNotExist",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: nonExistentModifierOptionID, Quantity: 1},
						},
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "ModifierOptionUnavailable",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: unavailableOption.ID, Quantity: 1},
						},
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "ModifierOptionBelongsToDifferentMenuItem",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: optionForItem2.ID, Quantity: 1}, // This belongs to menuItem3
						},
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "ModifierQuantityZero",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 0},
						},
					},
				},
			},
			expected: http.StatusBadRequest,
		},
		{
			testName: "ModifierQuantityNegative",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: -1},
						},
					},
				},
			},
			expected: http.StatusBadRequest,
		},
		{
			testName: "RequiredModifierGroupZeroSelections",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID:      menuItem1.ID,
						Quantity:        1,
						ModifierOptions: []handler.ModifierOption{}, // No modifiers selected
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "SumOfQuantitiesExceedsMax",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 2},
							{ModifierID: requiredOption2.ID, Quantity: 2}, // Total: 4, Max: 2
						},
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "OptionalModifierGroupZeroSelections",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1}, // Satisfy required
							// No optional modifiers - this should be valid
						},
					},
				},
			},
			expected: http.StatusCreated, // Should succeed
		},
		{
			testName: "SameModifierOptionIDMultipleTimes",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1},
							{ModifierID: requiredOption1.ID, Quantity: 1}, // Same ID appears twice
						},
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
		{
			testName: "ValidOrderWithOptionalModifiers",
			body: handler.CreateOrderSchema{
				OrderType:    dto.OrderTypeDINE_IN,
				RestaurantID: restaurant1.ID,
				OrderItems: []handler.OrderItemSchema{
					{
						MenuItemID: menuItem1.ID,
						Quantity:   1,
						ModifierOptions: []handler.ModifierOption{
							{ModifierID: requiredOption1.ID, Quantity: 1},
							{ModifierID: optionalOption1.ID, Quantity: 2},
						},
					},
				},
			},
			expected: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		s.Run(tt.testName, func() {
			var body []byte
			var err error
			body, err = json.Marshal(tt.body)
			s.Require().NoError(err)

			req := httptest.NewRequest(http.MethodPost, "/api/public/order", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server := s.CreateServer()
			server.Engine().ServeHTTP(w, req)

			s.Equal(tt.expected, w.Code, "Test: %s, Response: %s", tt.testName, w.Body.String())
		})
	}
}
