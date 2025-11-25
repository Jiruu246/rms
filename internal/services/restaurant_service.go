package services

import (
	"context"
	"fmt"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/restaurant"
	"github.com/Jiruu246/rms/internal/repos"
	"github.com/google/uuid"
)

type RestaurantService interface {
	Create(ctx context.Context, req *dto.CreateRestaurantRequest) (*ent.Restaurant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Restaurant, error)
	GetAll(ctx context.Context) ([]*ent.Restaurant, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*ent.Restaurant, error)
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

func (s *restaurantService) Create(ctx context.Context, req *dto.CreateRestaurantRequest) (*ent.Restaurant, error) {
	// Create restaurant entity from request
	rest := &ent.Restaurant{
		Name:        req.Name,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		ZipCode:     req.ZipCode,
		Country:     req.Country,
		Currency:    req.Currency,
	}

	// Set optional fields
	rest.LogoURL = req.LogoURL
	rest.CoverImageURL = req.CoverImageURL

	if req.Status != "" {
		switch req.Status {
		case "active":
			rest.Status = restaurant.StatusActive
		case "inactive":
			rest.Status = restaurant.StatusInactive
		case "closed":
			rest.Status = restaurant.StatusClosed
		default:
			rest.Status = restaurant.StatusActive
		}
	} else {
		rest.Status = restaurant.StatusActive
	}

	if req.OperatingHours != nil {
		rest.OperatingHours = req.OperatingHours
	}

	return s.repo.Create(ctx, rest)
}

func (s *restaurantService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Restaurant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *restaurantService) GetAll(ctx context.Context) ([]*ent.Restaurant, error) {
	return s.repo.GetAll(ctx)
}

func (s *restaurantService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateRestaurantRequest) (*ent.Restaurant, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("restaurant not found: %w", err)
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.City != nil {
		existing.City = *req.City
	}
	if req.State != nil {
		existing.State = *req.State
	}
	if req.ZipCode != nil {
		existing.ZipCode = *req.ZipCode
	}
	if req.Country != nil {
		existing.Country = *req.Country
	}
	if req.LogoURL != nil {
		existing.LogoURL = *req.LogoURL
	}
	if req.CoverImageURL != nil {
		existing.CoverImageURL = *req.CoverImageURL
	}
	if req.Status != nil {
		switch *req.Status {
		case "active":
			existing.Status = restaurant.StatusActive
		case "inactive":
			existing.Status = restaurant.StatusInactive
		case "closed":
			existing.Status = restaurant.StatusClosed
		}
	}
	if req.OperatingHours != nil {
		existing.OperatingHours = *req.OperatingHours
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}

	return s.repo.Update(ctx, existing)
}

func (s *restaurantService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
