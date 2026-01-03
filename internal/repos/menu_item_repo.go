package repos

import (
	"context"
	"fmt"

	ds "github.com/Jiruu246/rms/internal/data_structures"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/menuitem"
	"github.com/google/uuid"
)

type ItemModifierPair struct {
	MenuItemID       int64
	ModifierOptionID uuid.UUID
}

type MenuItemQueryOptions func(*ent.MenuItemQuery)

func WithModifierOptions() MenuItemQueryOptions {
	return func(query *ent.MenuItemQuery) {
		query.WithModifiers()
	}
}

type MenuItemRepository interface {
	Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItem, error)
	GetAll(ctx context.Context) ([]*dto.MenuItem, error)
	GetByID(ctx context.Context, id int64) (*dto.MenuItem, error)
	GetByIDsStrict(ctx context.Context, ids ds.Set[int64], opts ...MenuItemQueryOptions) (map[int64]*dto.MenuItem, error)
	Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItem, error)
	Delete(ctx context.Context, id int64) error
	// FindUnavailableMenuItemsByIDsAndRestaurant(ctx context.Context, ids []int64, restaurantID uuid.UUID) ([]int64, error)
	// FindUnavailableModifierOptionsForMenuItems(ctx context.Context, pairs []ItemModifierPair) ([]ItemModifierPair, error)
}

type menuItemRepository struct {
	client *ent.Client
}

func NewEntMenuItemRepository(client *ent.Client) MenuItemRepository {
	return &menuItemRepository{client: client}
}

func (r *menuItemRepository) Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItem, error) {
	create := r.client.MenuItem.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetPrice(req.Price).
		SetIsAvailable(req.IsAvailable).
		SetRestaurantID(req.RestaurantID)
	if req.CategoryID != uuid.Nil {
		create = create.SetCategoryID(req.CategoryID)
	}
	item, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}
	return mapToMenuItemResponse(item), nil
}

// TODO: Implement pagination, filtering, sorting
func (r *menuItemRepository) GetAll(ctx context.Context) ([]*dto.MenuItem, error) {
	items, err := r.client.MenuItem.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu items: %w", err)
	}
	var responses []*dto.MenuItem
	for _, item := range items {
		responses = append(responses, mapToMenuItemResponse(item))
	}
	return responses, nil
}

func (r *menuItemRepository) GetByID(ctx context.Context, id int64) (*dto.MenuItem, error) {
	item, err := r.client.MenuItem.Query().Where(menuitem.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("menu item not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get menu item: %w", err)
	}
	return mapToMenuItemResponse(item), nil
}

func (r *menuItemRepository) GetByIDsStrict(ctx context.Context, ids ds.Set[int64], opts ...MenuItemQueryOptions) (map[int64]*dto.MenuItem, error) {
	query := r.client.MenuItem.
		Query().
		Where(menuitem.IDIn(ids.Items()...))

	for _, opt := range opts {
		opt(query)
	}

	items, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu items: %w", err)
	}
	if len(items) != ids.Size() {
		return nil, fmt.Errorf("one or more menu items not found")
	}
	responses := make(map[int64]*dto.MenuItem, len(items))
	for _, item := range items {
		responses[item.ID] = mapToMenuItemResponse(item)
	}
	return responses, nil
}

// func (r *menuItemRepository) FindUnavailableMenuItemsByIDsAndRestaurant(ctx context.Context, ids []int64, restaurantID uuid.UUID) ([]int64, error) {
// 	menuitem, err := r.client.MenuItem.Query().
// 		Select(
// 			menuitem.FieldID,
// 		).
// 		Where(
// 			menuitem.IDIn(ids...),
// 			menuitem.Not(
// 				menuitem.And(
// 					menuitem.RestaurantIDEQ(restaurantID),
// 					menuitem.IsAvailableEQ(true),
// 				),
// 			),
// 		).All(ctx)

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to find unavailable menu items: %w", err)
// 	}

// 	var unavailableIDs []int64
// 	for _, item := range menuitem {
// 		unavailableIDs = append(unavailableIDs, item.ID)
// 	}
// 	return unavailableIDs, nil
// }

func (r *menuItemRepository) Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItem, error) {
	update := r.client.MenuItem.UpdateOneID(id)
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.Description != nil {
		update.SetDescription(*req.Description)
	}
	if req.Price != nil {
		update.SetPrice(*req.Price)
	}
	if req.IsAvailable != nil {
		update.SetIsAvailable(*req.IsAvailable)
	}
	if req.CategoryID != nil {
		update.SetCategoryID(*req.CategoryID)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}
	return mapToMenuItemResponse(updated), nil
}

func (r *menuItemRepository) Delete(ctx context.Context, id int64) error {
	err := r.client.MenuItem.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("menu item not found with id %d", id)
		}
		return fmt.Errorf("failed to delete menu item: %w", err)
	}
	return nil
}

func mapToMenuItemResponse(item *ent.MenuItem) *dto.MenuItem {
	var modifiers []dto.Modifier

	if item.Edges.Modifiers != nil {
		modifiers = make([]dto.Modifier, len(item.Edges.Modifiers))
		for i, mod := range item.Edges.Modifiers {
			modifiers[i] = *mapToModifier(mod)
		}
	}

	return &dto.MenuItem{
		ID:           item.ID,
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		IsAvailable:  item.IsAvailable,
		RestaurantID: item.RestaurantID,
		CategoryID:   item.CategoryID,
		Modifiers:    modifiers,
	}
}
