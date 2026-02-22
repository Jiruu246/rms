package repos

import (
	"context"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/userauthprovider"
)

type UserAuthProviderRepository interface {
	Create(ctx context.Context, userAuthProvider *dto.UserAuthProvider) error
	GetByUserIDAndProvider(ctx context.Context, userID string, provider string) (*dto.UserAuthProvider, error)
}

type EntUserAuthProviderRepository struct {
	client *ent.Client
}

func NewEntUserAuthProviderRepository(client *ent.Client) *EntUserAuthProviderRepository {
	return &EntUserAuthProviderRepository{client: client}
}

func (r *EntUserAuthProviderRepository) Create(ctx context.Context, userAuthProvider *dto.UserAuthProvider) error {
	_, err := r.client.UserAuthProvider.
		Create().
		SetUserID(userAuthProvider.UserID).
		SetProviderUserID(userAuthProvider.ProviderUserID).
		SetProvider(userAuthProvider.Provider).
		Save(ctx)
	return err
}

func (r *EntUserAuthProviderRepository) GetByUserIDAndProvider(ctx context.Context, userID string, provider string) (*dto.UserAuthProvider, error) {
	user, err := r.client.UserAuthProvider.Query().
		Where(
			userauthprovider.ProviderUserIDEQ(userID),
			userauthprovider.ProviderEQ(provider),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.UserAuthProvider{
		UserID:         user.UserID,
		ProviderUserID: user.ProviderUserID,
		Provider:       user.Provider,
	}, nil
}
