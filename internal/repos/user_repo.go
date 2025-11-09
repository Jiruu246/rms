package repos

import (
	"context"

	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/customer"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *ent.Customer) (*ent.Customer, error)
	GetByEmail(ctx context.Context, email string) (*ent.Customer, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Customer, error)
	Update(ctx context.Context, user *ent.Customer) (*ent.Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type EntUserRepository struct {
	client *ent.Client
}

func NewEntUserRepository(client *ent.Client) *EntUserRepository {
	return &EntUserRepository{client: client}
}

func (r *EntUserRepository) Create(ctx context.Context, user *ent.Customer) (*ent.Customer, error) {
	created, err := r.client.Customer.
		Create().
		SetName(user.Name).
		SetEmail(user.Email).
		SetPasswordHash(user.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *EntUserRepository) GetByEmail(ctx context.Context, email string) (*ent.Customer, error) {
	return r.client.Customer.Query().
		Where(customer.EmailEQ(email)).
		Only(ctx)
}

func (r *EntUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Customer, error) {
	return r.client.Customer.Get(ctx, id)
}

func (r *EntUserRepository) Update(ctx context.Context, user *ent.Customer) (*ent.Customer, error) {
	updated, err := r.client.Customer.
		UpdateOneID(user.ID).
		SetName(user.Name).
		SetEmail(user.Email).
		SetPasswordHash(user.PasswordHash).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *EntUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.Customer.DeleteOneID(id).Exec(ctx)
}
