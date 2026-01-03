package services

import (
	"context"
	"fmt"

	ds "github.com/Jiruu246/rms/internal/data_structures"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/google/uuid"
)

type CreateOrderInput struct {
	OrderType    dto.OrderType
	RestaurantID uuid.UUID
	OrderItems   []OrderItemInput
}

type ModifierOptionInput struct {
	ModifierOptionID uuid.UUID
	Quantity         int
}

type OrderItemInput struct {
	MenuItemID         int64
	Quantity           int
	SpecialInstruction string
	ModifierOptions    []ModifierOptionInput
}

type OrderService interface {
	Create(ctx context.Context, input CreateOrderInput) (*dto.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Order, error)
	GetAllByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*dto.Order, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderRequest) (*dto.Order, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type orderService struct {
	OrderRepo          repos.OrderRepository
	MenuItemRepo       repos.MenuItemRepository
	ModifierOptionRepo repos.ModifierOptionRepository
}

func NewOrderService(
	orderRepo repos.OrderRepository,
	menuItemRepo repos.MenuItemRepository,
	modifierOptionRepo repos.ModifierOptionRepository,
) OrderService {
	return &orderService{
		OrderRepo:          orderRepo,
		MenuItemRepo:       menuItemRepo,
		ModifierOptionRepo: modifierOptionRepo,
	}
}

func (s *orderService) Create(ctx context.Context, input CreateOrderInput) (*dto.Order, error) {
	if len(input.OrderItems) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}

	uniqueItemIDs := ds.NewSet[int64]()
	uniqueModifierIDs := ds.NewSet[uuid.UUID]()
	for _, item := range input.OrderItems {
		uniqueItemIDs.Add(item.MenuItemID)
		for _, mod := range item.ModifierOptions {
			uniqueModifierIDs.Add(mod.ModifierOptionID)
		}
	}

	menuItemsFromDB, err := s.MenuItemRepo.GetByIDsStrict(ctx, *uniqueItemIDs, repos.WithModifierOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to get menu items: %w", err)
	}

	modifierOptionsFromDB, err := s.ModifierOptionRepo.GetByIDsStrict(ctx, *uniqueModifierIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifier options: %w", err)
	}

	err = s.validateOrderItems(input.OrderItems, input.RestaurantID, menuItemsFromDB, modifierOptionsFromDB)
	if err != nil {
		return nil, err
	}

	var orderItems []repos.OrderItemData
	for _, item := range input.OrderItems {
		var modifiers []repos.ModifierItemData
		for _, m := range item.ModifierOptions {
			modifiers = append(modifiers, repos.ModifierItemData{
				ModifierOptionID: m.ModifierOptionID,
				Quantity:         m.Quantity,
				OptionName:       modifierOptionsFromDB[m.ModifierOptionID].Name,
				OptionPrice:      modifierOptionsFromDB[m.ModifierOptionID].Price,
			})
		}
		orderItems = append(orderItems, repos.OrderItemData{
			MenuItemID:          item.MenuItemID,
			Quantity:            item.Quantity,
			ItemName:            menuItemsFromDB[item.MenuItemID].Name,
			ItemPrice:           menuItemsFromDB[item.MenuItemID].Price,
			SpecialInstructions: item.SpecialInstruction,
			ModifierOptions:     modifiers,
		})
	}
	data := &repos.CreateOrderData{
		OrderType:     input.OrderType,
		OrderStatus:   dto.OrderStatusOPEN,
		PaymentStatus: dto.PaymentStatusUNPAID,
		RestaurantID:  input.RestaurantID,
		OrderItems:    orderItems,
	}
	return s.OrderRepo.Create(ctx, data)
}

func (s *orderService) GetByID(ctx context.Context, id uuid.UUID) (*dto.Order, error) {
	return s.OrderRepo.GetByID(ctx, id, repos.WithOrderItems(repos.WithOrderItemModifierOptions()))
}

func (s *orderService) GetAllByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*dto.Order, error) {
	return s.OrderRepo.GetAllByRestaurant(ctx, restaurantID)
}

func (s *orderService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateOrderRequest) (*dto.Order, error) {
	return s.OrderRepo.Update(ctx, &dto.UpdateOrderData{
		Request: req,
		ID:      id,
	})
}

func (s *orderService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.OrderRepo.Delete(ctx, id)
}

func (s *orderService) validateOrderItems(
	items []OrderItemInput,
	restaurantID uuid.UUID,
	menuItemsFromDB map[int64]*dto.MenuItem,
	modifierOptionsFromDB map[uuid.UUID]*dto.ModifierOption,
) error {
	for _, item := range items {
		menuItem, exists := menuItemsFromDB[item.MenuItemID]
		if !exists {
			return fmt.Errorf("menu item with ID %d does not exist", item.MenuItemID)
		}
		if menuItem.RestaurantID != restaurantID {
			return fmt.Errorf("menu item with ID %d does not belong to restaurant %s", item.MenuItemID, restaurantID)
		}
		if !menuItem.IsAvailable {
			return fmt.Errorf("menu item with ID %d is not available", item.MenuItemID)
		}
		if menuItem.Modifiers == nil {
			return fmt.Errorf("INTERNAL ERROR: menu item with ID %d has no modifiers loaded", item.MenuItemID)
		}

		ModifiersIds := ds.NewSet[uuid.UUID]()
		for _, mod := range menuItem.Modifiers {
			ModifiersIds.Add(mod.ID)
		}

		modifierGroup := make(map[uuid.UUID][]ModifierOptionInput)
		for _, modOpt := range item.ModifierOptions {
			modifierOption, exists := modifierOptionsFromDB[modOpt.ModifierOptionID]
			if !exists {
				return fmt.Errorf("modifier option with ID %s does not exist", modOpt.ModifierOptionID)
			}
			if !modifierOption.Available {
				return fmt.Errorf("modifier option with ID %s is not available", modOpt.ModifierOptionID)
			}
			if modOpt.Quantity < 1 {
				return fmt.Errorf("modifier option with ID %s has invalid quantity %d", modOpt.ModifierOptionID, modOpt.Quantity)
			}
			if !ModifiersIds.Contains(modifierOption.ModifierID) {
				return fmt.Errorf("modifier option with ID %s does not belong to menu item with ID %d", modOpt.ModifierOptionID, item.MenuItemID)
			}
			modifierGroup[modifierOption.ModifierID] = append(modifierGroup[modifierOption.ModifierID], modOpt)
		}

		for _, mod := range menuItem.Modifiers {
			if mod.Required {
				if _, exists := modifierGroup[mod.ID]; !exists {
					return fmt.Errorf("required modifier group with ID %s has no selected options", mod.ID)
				}
			}
		}

		for groupId, modOptions := range modifierGroup {
			modifier, found := utils.FindFirst(menuItem.Modifiers, func(m dto.Modifier) bool {
				return m.ID == groupId
			})
			if !found {
				return fmt.Errorf("INTERNAL ERROR: modifier with ID %s is not loaded correctly", groupId)
			}
			NumSelected := utils.Reduce(
				modOptions,
				func(acc int, modOpt ModifierOptionInput) int {
					return acc + modOpt.Quantity
				},
				0,
			)
			if (modifier.Required && NumSelected < 1) || NumSelected > modifier.Max {
				min := 0
				if modifier.Required {
					min = 1
				}
				return fmt.Errorf("number of selected modifier options for modifier ID %s violates constraints (%d selected, min %d, max %d)",
					groupId,
					NumSelected,
					min,
					modifier.Max,
				)
			}
		}
	}

	return nil
}
