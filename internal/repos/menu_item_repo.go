package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/menuitem"
	"github.com/google/uuid"
)

type MenuItemRepository interface {
	Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItemResponse, error)
	GetAll(ctx context.Context) ([]*dto.MenuItemResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.MenuItemResponse, error)
	Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItemResponse, error)
	Delete(ctx context.Context, id int64) error
}

type menuItemRepository struct {
	client *ent.Client
}

func NewEntMenuItemRepository(client *ent.Client) MenuItemRepository {
	return &menuItemRepository{client: client}
}

func (r *menuItemRepository) Create(ctx context.Context, req *dto.CreateMenuItemRequest) (*dto.MenuItemResponse, error) {
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
func (r *menuItemRepository) GetAll(ctx context.Context) ([]*dto.MenuItemResponse, error) {
	items, err := r.client.MenuItem.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu items: %w", err)
	}
	var responses []*dto.MenuItemResponse
	for _, item := range items {
		responses = append(responses, mapToMenuItemResponse(item))
	}
	return responses, nil
}

func (r *menuItemRepository) GetByID(ctx context.Context, id int64) (*dto.MenuItemResponse, error) {
	item, err := r.client.MenuItem.Query().Where(menuitem.IDEQ(id)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("menu item not found with id %d", id)
		}
		return nil, fmt.Errorf("failed to get menu item: %w", err)
	}
	return mapToMenuItemResponse(item), nil
}

func (r *menuItemRepository) Update(ctx context.Context, id int64, req *dto.UpdateMenuItemRequest) (*dto.MenuItemResponse, error) {
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

func mapToMenuItemResponse(item *ent.MenuItem) *dto.MenuItemResponse {
	return &dto.MenuItemResponse{
		ID:           item.ID,
		Name:         item.Name,
		Description:  item.Description,
		Price:        item.Price,
		IsAvailable:  item.IsAvailable,
		RestaurantID: item.RestaurantID,
		CategoryID:   item.CategoryID,
	}
}
