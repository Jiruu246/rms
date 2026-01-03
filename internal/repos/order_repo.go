package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/order"
	"github.com/google/uuid"
)

type OrderQueryOptions func(*ent.OrderQuery)

type OrderItemQueryOptions func(*ent.OrderItemQuery)

func WithOrderItems(opts ...OrderItemQueryOptions) OrderQueryOptions {
	return func(query *ent.OrderQuery) {
		query.WithOrderItems(func(oq *ent.OrderItemQuery) {
			for _, opt := range opts {
				opt(oq)
			}
		})
	}
}

func WithOrderItemModifierOptions() OrderItemQueryOptions {
	return func(query *ent.OrderItemQuery) {
		query.WithOrderItemModifierOptions()
	}
}

type CreateOrderData struct {
	OrderType     dto.OrderType
	OrderStatus   dto.OrderStatus
	PaymentStatus dto.PaymentStatus
	RestaurantID  uuid.UUID
	OrderItems    []OrderItemData
}

type ModifierItemData struct {
	ModifierOptionID uuid.UUID
	Quantity         int
	OptionName       string
	OptionPrice      float64
}

type OrderItemData struct {
	MenuItemID          int64
	Quantity            int
	ItemName            string
	ItemPrice           float64
	SpecialInstructions string
	ModifierOptions     []ModifierItemData
}

type OrderRepository interface {
	Create(ctx context.Context, data *CreateOrderData) (*dto.Order, error)
	GetByID(ctx context.Context, id uuid.UUID, opts ...OrderQueryOptions) (*dto.Order, error)
	Update(ctx context.Context, data *dto.UpdateOrderData) (*dto.Order, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAllByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*dto.Order, error)
}

type orderRepository struct {
	client *ent.Client
}

func NewEntOrderRepository(client *ent.Client) OrderRepository {
	return &orderRepository{client: client}
}

func (r *orderRepository) Create(ctx context.Context, data *CreateOrderData) (*dto.Order, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	createOrder := tx.Order.Create().
		SetOrderType(order.OrderType(data.OrderType)).
		SetOrderStatus(order.OrderStatus(data.OrderStatus)).
		SetPaymentStatus(order.PaymentStatus(data.PaymentStatus)).
		SetRestaurantID(data.RestaurantID)
	ord, err := createOrder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range data.OrderItems {
		orderItemCreate := tx.OrderItem.Create().
			SetOrderID(ord.ID).
			SetQuantity(item.Quantity).
			SetSpecialInstructions(item.SpecialInstructions).
			SetMenuItemID(item.MenuItemID).
			SetItemName(item.ItemName).
			SetItemPrice(item.ItemPrice)
		orderItem, err := orderItemCreate.Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
		for _, mod := range item.ModifierOptions {
			_, err := tx.OrderItemModifierOption.Create().
				SetOrderItemID(orderItem.ID).
				SetModifierOptionID(mod.ModifierOptionID).
				SetQuantity(mod.Quantity).
				SetOptionName(mod.OptionName).
				SetOptionPrice(mod.OptionPrice).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create order item modifier: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, ord.ID, WithOrderItems(WithOrderItemModifierOptions()))
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID, opts ...OrderQueryOptions) (*dto.Order, error) {
	query := r.client.Order.Query().Where(order.IDEQ(id))
	for _, opt := range opts {
		opt(query)
	}
	ord, err := query.First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("order not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return mapToOrderResponse(ord), nil
}

func (r *orderRepository) Update(ctx context.Context, data *dto.UpdateOrderData) (*dto.Order, error) {
	update := r.client.Order.UpdateOneID(data.ID)
	if data.Request.OrderType != nil {
		update.SetOrderType(order.OrderType(*data.Request.OrderType))
	}
	if data.Request.OrderStatus != nil {
		update.SetOrderStatus(order.OrderStatus(*data.Request.OrderStatus))
	}
	if data.Request.RestaurantID != nil {
		update.SetRestaurantID(*data.Request.RestaurantID)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}
	return mapToOrderResponse(updated), nil
}

func (r *orderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Order.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("order not found with id %s", id)
		}
		return fmt.Errorf("failed to delete order: %w", err)
	}
	return nil
}

func (r *orderRepository) GetAllByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*dto.Order, error) {
	orders, err := r.client.Order.Query().Where(order.RestaurantIDEQ(restaurantID)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	var responses []*dto.Order
	for _, o := range orders {
		responses = append(responses, mapToOrderResponse(o))
	}
	return responses, nil
}

func mapToOrderResponse(order *ent.Order) *dto.Order {
	var orderItems []dto.OrderItem
	orderItems = nil

	if order.Edges.OrderItems != nil {
		for _, oi := range order.Edges.OrderItems {
			var modifierOptions []dto.OrderItemModifierOption
			if oi.Edges.OrderItemModifierOptions != nil {
				for _, mo := range oi.Edges.OrderItemModifierOptions {
					modifierOptions = append(modifierOptions, dto.OrderItemModifierOption{
						OrderItemID:      mo.OrderItemID,
						ModifierOptionID: mo.ModifierOptionID,
						Quantity:         mo.Quantity,
						OptionName:       mo.OptionName,
						OptionPrice:      mo.OptionPrice,
					})
				}
			}
			orderItems = append(orderItems, dto.OrderItem{
				ID:                  oi.ID,
				MenuItemID:          oi.MenuItemID,
				Quantity:            oi.Quantity,
				ItemName:            oi.ItemName,
				ItemPrice:           oi.ItemPrice,
				SpecialInstructions: oi.SpecialInstructions,
				ModifierOptions:     modifierOptions,
				OrderID:             oi.OrderID,
			})
		}
	}

	return &dto.Order{
		ID:           order.ID,
		OrderType:    dto.OrderType(order.OrderType),
		OrderStatus:  dto.OrderStatus(order.OrderStatus),
		RestaurantID: order.RestaurantID,
		OrderItems:   orderItems,
	}
}
