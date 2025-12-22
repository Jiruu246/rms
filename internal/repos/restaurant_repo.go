package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
)

type RestaurantRepository interface {
	Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantResponse, error)
	Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.RestaurantResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*dto.RestaurantResponse, error)
}

type restaurantRepository struct {
	client *ent.Client
}

// NewEntRestaurantRepository creates a new Ent-based restaurant repository
func NewEntRestaurantRepository(client *ent.Client) RestaurantRepository {
	return &restaurantRepository{
		client: client,
	}
}

func (r *restaurantRepository) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error) {
	create, err := r.client.Restaurant.Create().
		SetName(data.Request.Name).
		SetDescription(data.Request.Description).
		SetPhone(data.Request.Phone).
		SetEmail(data.Request.Email).
		SetAddress(data.Request.Address).
		SetCity(data.Request.City).
		SetState(data.Request.State).
		SetZipCode(data.Request.ZipCode).
		SetCountry(data.Request.Country).
		SetStatus(restaurant.StatusActive).
		SetCurrency(data.Request.Currency).
		SetLogoURL(data.Request.LogoURL).
		SetCoverImageURL(data.Request.CoverImageURL).
		SetOperatingHours(data.Request.OperatingHours).
		SetUserID(data.UserID).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	return mapToRestaurantResponse(create), nil
}

func (r *restaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantResponse, error) {
	restaurant, err := r.client.Restaurant.Query().
		Where(restaurant.IDEQ(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("restaurant not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get restaurant: %w", err)
	}

	return mapToRestaurantResponse(restaurant), nil
}

func (r *restaurantRepository) Update(ctx context.Context, data *dto.UpdateRestaurantData) (*dto.RestaurantResponse, error) {
	update := r.client.Restaurant.UpdateOneID(data.ID)

	if data.Request.Name != nil {
		update.SetName(*data.Request.Name)
	}

	if data.Request.Description != nil {
		update.SetDescription(*data.Request.Description)
	}

	if data.Request.Phone != nil {
		update.SetPhone(*data.Request.Phone)
	}

	if data.Request.Email != nil {
		update.SetEmail(*data.Request.Email)
	}

	if data.Request.Address != nil {
		update.SetAddress(*data.Request.Address)
	}

	if data.Request.City != nil {
		update.SetCity(*data.Request.City)
	}

	if data.Request.State != nil {
		update.SetState(*data.Request.State)
	}

	if data.Request.ZipCode != nil {
		update.SetZipCode(*data.Request.ZipCode)
	}

	if data.Request.Country != nil {
		update.SetCountry(*data.Request.Country)
	}

	if data.Request.LogoURL != nil {
		update.SetLogoURL(*data.Request.LogoURL)
	}

	if data.Request.CoverImageURL != nil {
		update.SetCoverImageURL(*data.Request.CoverImageURL)
	}

	if data.Request.Status != nil {
		update.SetStatus(restaurant.Status(*data.Request.Status))
	}

	if data.Request.OperatingHours != nil {
		update.SetOperatingHours(*data.Request.OperatingHours)
	}

	if data.Request.Currency != nil {
		update.SetCurrency(*data.Request.Currency)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update restaurant: %w", err)
	}

	return mapToRestaurantResponse(updated), nil
}

func (r *restaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Restaurant.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("restaurant not found with id %s", id)
		}
		return fmt.Errorf("failed to delete restaurant: %w", err)
	}
	return nil
}

func (r *restaurantRepository) GetAll(ctx context.Context) ([]*dto.RestaurantResponse, error) {
	restaurants, err := r.client.Restaurant.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all restaurants: %w", err)
	}

	var responses []*dto.RestaurantResponse
	for _, res := range restaurants {
		responses = append(responses, mapToRestaurantResponse(res))
	}

	return responses, nil
}

func mapToRestaurantResponse(restaurant *ent.Restaurant) *dto.RestaurantResponse {
	return &dto.RestaurantResponse{
		ID:             restaurant.ID,
		Name:           restaurant.Name,
		Description:    restaurant.Description,
		Phone:          restaurant.Phone,
		Email:          restaurant.Email,
		Address:        restaurant.Address,
		City:           restaurant.City,
		State:          restaurant.State,
		ZipCode:        restaurant.ZipCode,
		Country:        restaurant.Country,
		LogoURL:        restaurant.LogoURL,
		CoverImageURL:  restaurant.CoverImageURL,
		Status:         restaurant.Status.String(),
		OperatingHours: restaurant.OperatingHours,
		Currency:       restaurant.Currency,
	}
}
