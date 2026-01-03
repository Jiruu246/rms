package handler

import (
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ModifierOption struct {
	ModifierID uuid.UUID `json:"modifier_id" validate:"required" binding:"required"`
	Quantity   int       `json:"quantity" validate:"required,min=1" binding:"required"`
}

type OrderItemSchema struct {
	MenuItemID      int64            `json:"menu_item_id" validate:"required" binding:"required"`
	Quantity        int              `json:"quantity" validate:"required,min=1" binding:"required"`
	Notes           string           `json:"notes,omitempty" validate:"max=255" binding:"omitempty"`
	ModifierOptions []ModifierOption `json:"modifiers,omitempty" validate:"dive" binding:"omitempty"`
}
type CreateOrderSchema struct {
	OrderType    dto.OrderType     `json:"order_type" validate:"required" binding:"required"`
	RestaurantID uuid.UUID         `json:"restaurant_id" validate:"required" binding:"required"`
	OrderItems   []OrderItemSchema `json:"order_items" validate:"dive"`
}

type OrderHandler struct {
	service services.OrderService
}

func NewOrderHandler(service services.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrderPub handles POST /api/orders
func (h *OrderHandler) CreateOrderPub(c *gin.Context) {
	var req CreateOrderSchema
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	var orderItems []services.OrderItemInput
	for _, item := range req.OrderItems {
		var modifiers []services.ModifierOptionInput
		for _, m := range item.ModifierOptions {
			modifiers = append(modifiers, services.ModifierOptionInput{
				ModifierOptionID: m.ModifierID,
				Quantity:         m.Quantity,
			})
		}
		orderItems = append(orderItems, services.OrderItemInput{
			MenuItemID:         item.MenuItemID,
			Quantity:           item.Quantity,
			SpecialInstruction: item.Notes,
			ModifierOptions:    modifiers,
		})
	}
	input := services.CreateOrderInput{
		OrderType:    req.OrderType,
		RestaurantID: req.RestaurantID,
		OrderItems:   orderItems,
	}
	created, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create order")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetOrder handles GET /api/orders/{id}
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid order ID format")
		return
	}
	order, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.WriteNotFound(c.Writer, "Order not found")
		return
	}
	utils.WriteSuccess(c.Writer, order)
}

// GetOrders handles GET /api/orders?restaurant_id=xxx
func (h *OrderHandler) GetOrders(c *gin.Context) {
	restaurantIDStr := c.Query("restaurant_id")
	if restaurantIDStr == "" {
		utils.WriteBadRequest(c.Writer, "restaurant_id is required")
		return
	}
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid restaurant_id format")
		return
	}
	orders, err := h.service.GetAllByRestaurant(c.Request.Context(), restaurantID)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to fetch orders")
		return
	}
	utils.WriteSuccess(c.Writer, orders)
}

// UpdateOrder handles PATCH /api/orders/{id}
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid order ID format")
		return
	}
	var req dto.UpdateOrderRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to update order")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteOrder handles DELETE /api/orders/{id}
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid order ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.WriteNotFound(c.Writer, "Order not found")
		return
	}
	utils.WriteNoContent(c.Writer)
}
