package integration_tests

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/google/uuid"
)

func SetupUser(client *ent.Client, ctx context.Context) (*ent.User, error) {
	return client.User.Create().
		SetName("Test User").
		SetEmail(fmt.Sprintf("testuser_%s@example.com", uuid.NewString())).
		SetPasswordHash("hashedpassword").
		Save(ctx)
}

func SetupRestaurant(client *ent.Client, ctx context.Context) (*ent.Restaurant, error) {
	user, err := SetupUser(client, ctx)
	if err != nil {
		return nil, err
	}

	return client.Restaurant.Create().
		SetName("Test Restaurant").
		SetPhone("123-456-7890").
		SetEmail(fmt.Sprintf("testrestaurant_%s@example.com", uuid.NewString())).
		SetAddress("123 Main St").
		SetCity("Test City").
		SetState("TS").
		SetZipCode("12345").
		SetCountry("Test Country").
		SetCurrency("USD").
		SetStatus(restaurant.StatusActive).
		SetUser(user).
		Save(ctx)
}

func SetupCategory(client *ent.Client, ctx context.Context) (*ent.Category, error) {
	restaurant, err := SetupRestaurant(client, ctx)
	if err != nil {
		return nil, err
	}

	return client.Category.Create().
		SetName("Test Category").
		SetDescription("A test category description").
		SetRestaurant(restaurant).
		Save(ctx)
}

func CreateMenuItem(client *ent.Client, ctx context.Context) (*ent.MenuItem, error) {
	restaurant, err := SetupRestaurant(client, ctx)
	if err != nil {
		return nil, err
	}

	menuitem, err := client.MenuItem.Create().
		SetName("Test Menu Item").
		SetDescription("A test menu item description").
		SetPrice(9.99).
		SetIsAvailable(true).
		SetRestaurant(restaurant).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return menuitem, nil
}

func CreateModifier(client *ent.Client, ctx context.Context) (*ent.Modifier, error) {
	restaurant, err := SetupRestaurant(client, ctx)

	if err != nil {
		return nil, err
	}
	modifier, err := client.Modifier.Create().
		SetName("Test Modifier").
		SetRestaurant(restaurant).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return modifier, nil
}

func CreateModifierOption(client *ent.Client, ctx context.Context) (*ent.ModifierOption, error) {
	modifier, err := CreateModifier(client, ctx)
	if err != nil {
		return nil, err
	}
	modifierOption, err := client.ModifierOption.Create().
		SetName("Test Modifier Option").
		SetPrice(1.99).
		SetModifier(modifier).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return modifierOption, nil
}

func ptrString(s string) *string {
	return &s
}

func ptr(s string) *string {
	return &s
}
