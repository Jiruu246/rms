package repos

import (
	"context"
	"time"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/refreshtoken"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, hashedToken string, expiresAt time.Time) (*ent.RefreshToken, error)
	GetByID(ctx context.Context, tokenID string) (*ent.RefreshToken, error)
	GetActiveTokensByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenID uuid.UUID) error
	//TDN TODO: What's this method? is it used?
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) error
}

type entRefreshTokenRepository struct {
	client *ent.Client
}

func NewEntRefreshTokenRepository(client *ent.Client) RefreshTokenRepository {
	return &entRefreshTokenRepository{client: client}
}

func (r *entRefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, hashedToken string, expiresAt time.Time) (*ent.RefreshToken, error) {
	return r.client.RefreshToken.
		Create().
		SetUserID(userID).
		SetToken(hashedToken).
		SetExpiresAt(expiresAt).
		SetRevoked(false).
		Save(ctx)
}

func (r *entRefreshTokenRepository) GetByID(ctx context.Context, tokenID string) (*ent.RefreshToken, error) {
	id, err := uuid.Parse(tokenID)
	if err != nil {
		return nil, err
	}
	return r.client.RefreshToken.
		Query().
		Where(
			refreshtoken.ID(id),
			refreshtoken.Revoked(false),
			refreshtoken.ExpiresAtGT(time.Now()),
		).
		WithUser().
		First(ctx)
}

func (r *entRefreshTokenRepository) GetActiveTokensByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.RefreshToken, error) {
	return r.client.RefreshToken.
		Query().
		Where(
			refreshtoken.UserID(userID),
			refreshtoken.Revoked(false),
			refreshtoken.ExpiresAtGT(time.Now()),
		).
		All(ctx)
}

func (r *entRefreshTokenRepository) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	return r.client.RefreshToken.
		UpdateOneID(tokenID).
		SetRevoked(true).
		SetRevokedAt(time.Now()).
		Exec(ctx)
}

func (r *entRefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := r.client.RefreshToken.
		Update().
		Where(
			refreshtoken.UserID(userID),
			refreshtoken.Revoked(false),
		).
		SetRevoked(true).
		SetRevokedAt(time.Now()).
		Save(ctx)
	return err
}

func (r *entRefreshTokenRepository) UpdateLastUsed(ctx context.Context, tokenID uuid.UUID) error {
	return r.client.RefreshToken.
		UpdateOneID(tokenID).
		SetLastUsedAt(time.Now()).
		Exec(ctx)
}

func (r *entRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.client.RefreshToken.
		Delete().
		Where(
			refreshtoken.Or(
				refreshtoken.ExpiresAtLT(time.Now()),
				refreshtoken.Revoked(true),
			),
		).
		Exec(ctx)
	return err
}
