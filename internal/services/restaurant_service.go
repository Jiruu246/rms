package services

import (
	"context"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type RestaurantService interface {
	Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantResponse, error)
	GetAll(ctx context.Context) ([]*dto.RestaurantResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.RestaurantResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type restaurantService struct {
	repo repos.RestaurantRepository
}

func NewRestaurantService(repo repos.RestaurantRepository) RestaurantService {
	return &restaurantService{
		repo: repo,
	}
}

func (s *restaurantService) Create(ctx context.Context, data *dto.CreateRestaurantData) (*dto.RestaurantResponse, error) {
	return s.repo.Create(ctx, data)
}

func (s *restaurantService) GetByID(ctx context.Context, id uuid.UUID) (*dto.RestaurantResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *restaurantService) GetAll(ctx context.Context) ([]*dto.RestaurantResponse, error) {
	return s.repo.GetAll(ctx)
}

func (s *restaurantService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*dto.RestaurantResponse, error) {
	return s.repo.Update(ctx, &dto.UpdateRestaurantData{
		Request: req,
		ID:      id,
	})
}

func (s *restaurantService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
