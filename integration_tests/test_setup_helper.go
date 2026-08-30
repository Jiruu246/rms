package integration_tests

import (
	"context"
	"fmt"
	"time"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/mediaupload"
	"github.com/Jiruu246/rms/internal/ent/order"
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

	return CreateMenuItemForRestaurant(client, ctx, restaurant)
}

func CreateMenuItemForRestaurant(client *ent.Client, ctx context.Context, restaurant *ent.Restaurant) (*ent.MenuItem, error) {
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

	return CreateModifierForRestaurant(client, ctx, restaurant)
}

func CreateModifierForRestaurant(client *ent.Client, ctx context.Context, restaurant *ent.Restaurant) (*ent.Modifier, error) {
	modifier, err := client.Modifier.Create().
		SetName("Test Modifier").
		SetRestaurant(restaurant).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return modifier, nil
}

func CreateModifierForItem(client *ent.Client, ctx context.Context, menuItem *ent.MenuItem) (*ent.Modifier, error) {
	modifier, err := client.Modifier.Create().
		SetName("Test Modifier").
		SetRestaurantID(menuItem.RestaurantID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_, err = client.MenuItem.UpdateOne(menuItem).
		AddModifiers(modifier).
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

	return CreateModifierOptionForModifier(client, ctx, modifier)
}

func CreateModifierOptionForModifier(client *ent.Client, ctx context.Context, modifier *ent.Modifier) (*ent.ModifierOption, error) {
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

func SetupOrder(client *ent.Client, ctx context.Context) (*ent.Order, error) {
	restaurant, err := SetupRestaurant(client, ctx)
	if err != nil {
		return nil, err
	}

	return client.Order.Create().
		SetOrderType(order.OrderTypeDINE_IN).
		SetRestaurant(restaurant).
		Save(ctx)
}

func SetupMediaUpload(client *ent.Client, ctx context.Context) (*ent.MediaUpload, error) {
	user, err := SetupUser(client, ctx)
	if err != nil {
		return nil, err
	}

	return SetupMediaUploadForOwner(client, ctx, user.ID)
}

func SetupMediaUploadForOwner(client *ent.Client, ctx context.Context, ownerID uuid.UUID) (*ent.MediaUpload, error) {
	return client.MediaUpload.Create().
		SetOwnerID(ownerID).
		SetPurpose(mediaupload.PurposeMenuItemImage).
		SetObjectKey(fmt.Sprintf("uploads/%s", uuid.NewString())).
		SetExpiresAt(time.Now().Add(time.Hour)).
		Save(ctx)
}

// SetupMediaAsset creates a consumed upload and its resulting asset directly
// (bypassing MediaUploadRepository.Consume), for tests that only need an
// existing asset fixture rather than exercising the finalize flow itself.
func SetupMediaAsset(client *ent.Client, ctx context.Context) (*ent.MediaAsset, error) {
	upload, err := SetupMediaUpload(client, ctx)
	if err != nil {
		return nil, err
	}

	upload, err = upload.Update().SetStatus(mediaupload.StatusConsumed).Save(ctx)
	if err != nil {
		return nil, err
	}

	return client.MediaAsset.Create().
		SetUploadedByUserID(upload.OwnerID).
		SetUploadID(upload.ID).
		SetStorageKey(upload.ObjectKey).
		SetContentType("image/png").
		SetSizeBytes(1024).
		Save(ctx)
}

func ptrString(s string) *string {
	return &s
}

func ptr(s string) *string {
	return &s
}
