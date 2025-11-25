package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
)

type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *ent.Restaurant) (*ent.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Restaurant, error)
	Update(ctx context.Context, restaurant *ent.Restaurant) (*ent.Restaurant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*ent.Restaurant, error)
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

func (r *restaurantRepository) Create(ctx context.Context, restaurant *ent.Restaurant) (*ent.Restaurant, error) {
	create := r.client.Restaurant.Create().
		SetName(restaurant.Name).
		SetDescription(restaurant.Description).
		SetPhone(restaurant.Phone).
		SetEmail(restaurant.Email).
		SetAddress(restaurant.Address).
		SetCity(restaurant.City).
		SetState(restaurant.State).
		SetZipCode(restaurant.ZipCode).
		SetCountry(restaurant.Country).
		SetStatus(restaurant.Status).
		SetCurrency(restaurant.Currency)

	// Set optional fields if they are not empty
	if restaurant.LogoURL != "" {
		create = create.SetLogoURL(restaurant.LogoURL)
	}
	if restaurant.CoverImageURL != "" {
		create = create.SetCoverImageURL(restaurant.CoverImageURL)
	}
	if restaurant.OperatingHours != nil {
		create = create.SetOperatingHours(restaurant.OperatingHours)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create restaurant: %w", err)
	}

	return created, nil
}

func (r *restaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Restaurant, error) {
	restaurant, err := r.client.Restaurant.Query().
		Where(restaurant.IDEQ(id)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("restaurant not found with id %s", id)
		}
		return nil, fmt.Errorf("failed to get restaurant: %w", err)
	}

	return restaurant, nil
}

func (r *restaurantRepository) Update(ctx context.Context, rest *ent.Restaurant) (*ent.Restaurant, error) {
	update := r.client.Restaurant.UpdateOneID(rest.ID).
		SetName(rest.Name).
		SetDescription(rest.Description).
		SetPhone(rest.Phone).
		SetEmail(rest.Email).
		SetAddress(rest.Address).
		SetCity(rest.City).
		SetState(rest.State).
		SetZipCode(rest.ZipCode).
		SetCountry(rest.Country).
		SetStatus(rest.Status).
		SetCurrency(rest.Currency)

	// Set optional fields
	update = update.SetLogoURL(rest.LogoURL)
	update = update.SetCoverImageURL(rest.CoverImageURL)
	if rest.OperatingHours != nil {
		update = update.SetOperatingHours(rest.OperatingHours)
	}

	updated, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("restaurant not found with id %s", rest.ID)
		}
		return nil, fmt.Errorf("failed to update restaurant: %w", err)
	}

	return updated, nil
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

func (r *restaurantRepository) GetAll(ctx context.Context) ([]*ent.Restaurant, error) {
	restaurants, err := r.client.Restaurant.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all restaurants: %w", err)
	}
	return restaurants, nil
}
