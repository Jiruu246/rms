package dto

import (
	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusOPEN      OrderStatus = "OPEN"
	OrderStatusCONFIRMED OrderStatus = "CONFIRMED"
	OrderStatusCOMPLETED OrderStatus = "COMPLETED"
	OrderStatusCANCELLED OrderStatus = "CANCELLED"
)

type OrderType string

const (
	OrderTypeDINE_IN  OrderType = "DINE_IN"
	OrderTypeTAKEOUT  OrderType = "TAKEOUT"
	OrderTypeDELIVERY OrderType = "DELIVERY"
)

type PaymentStatus string

const (
	PaymentStatusUNPAID   PaymentStatus = "UNPAID"
	PaymentStatusPENDING  PaymentStatus = "PENDING"
	PaymentStatusPAID     PaymentStatus = "PAID"
	PaymentStatusREFUNDED PaymentStatus = "REFUNDED"
)

// UpdateOrderRequest for PATCH (partial update)
type UpdateOrderRequest struct {
	OrderNumber  *string    `json:"order_number"`
	OrderType    *string    `json:"order_type"`
	OrderStatus  *string    `json:"order_status"`
	RestaurantID *uuid.UUID `json:"restaurant_id"`
}

type UpdateOrderData struct {
	Request *UpdateOrderRequest
	ID      uuid.UUID
}

type OrderItemModifierOption struct {
	OrderItemID      uuid.UUID `json:"order_item_id"`
	ModifierOptionID uuid.UUID `json:"modifier_option_id"`
	Quantity         int       `json:"quantity"`
	OptionName       string    `json:"option_name"`
	OptionPrice      float64   `json:"option_price"`
}

type OrderItem struct {
	ID                  uuid.UUID                 `json:"id"`
	Quantity            int                       `json:"quantity"`
	SpecialInstructions string                    `json:"special_instructions"`
	MenuItemID          int64                     `json:"menu_item_id"`
	ItemName            string                    `json:"item_name"`
	ItemPrice           float64                   `json:"item_price"`
	ModifierOptions     []OrderItemModifierOption `json:"modifier_options"`
	OrderID             uuid.UUID                 `json:"order_id"`
}

type Order struct {
	ID           uuid.UUID   `json:"id"`
	OrderNumber  string      `json:"order_number"`
	OrderType    OrderType   `json:"order_type"`
	OrderStatus  OrderStatus `json:"order_status"`
	RestaurantID uuid.UUID   `json:"restaurant_id"`
	OrderItems   []OrderItem `json:"order_items"`
}
