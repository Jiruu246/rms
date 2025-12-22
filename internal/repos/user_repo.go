package repos

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *dto.RegisterUserRequest) (*dto.User, error)
	GetByEmail(ctx context.Context, email string) (*dto.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.User, error)
	Update(ctx context.Context, id uuid.UUID, user *dto.UpdateUserRequest) (*dto.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type EntUserRepository struct {
	client *ent.Client
}

func NewEntUserRepository(client *ent.Client) *EntUserRepository {
	return &EntUserRepository{client: client}
}

func (r *EntUserRepository) Create(ctx context.Context, req *dto.RegisterUserRequest) (*dto.User, error) {
	created, err := r.client.User.
		Create().
		SetName(req.Name).
		SetEmail(req.Email).
		SetPasswordHash(req.Password).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       created.ID,
		Name:     created.Name,
		Email:    created.Email,
		Password: created.PasswordHash,
	}, nil
}

func (r *EntUserRepository) GetByEmail(ctx context.Context, email string) (*dto.User, error) {
	user, err := r.client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.PasswordHash,
	}, nil
}

func (r *EntUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.User, error) {
	user, err := r.client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Password: user.PasswordHash,
	}, nil
}

func (r *EntUserRepository) Update(ctx context.Context, id uuid.UUID, user *dto.UpdateUserRequest) (*dto.User, error) {
	updateBuilder := r.client.User.UpdateOneID(id)

	hasUpdates := false

	if user.Name != nil {
		updateBuilder.SetName(*user.Name)
		hasUpdates = true
	}

	if user.Email != nil {
		updateBuilder.SetEmail(*user.Email)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no valid fields to update")
	}

	updated, err := updateBuilder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       updated.ID,
		Name:     updated.Name,
		Email:    updated.Email,
		Password: updated.PasswordHash,
	}, nil
}

func (r *EntUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.User.DeleteOneID(id).Exec(ctx)
}
