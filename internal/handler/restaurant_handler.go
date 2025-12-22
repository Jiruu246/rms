package handler

import (
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RestaurantHandler struct {
	service services.RestaurantService
}

func NewRestaurantHandler(service services.RestaurantService) *RestaurantHandler {
	return &RestaurantHandler{
		service: service,
	}
}

// CreateRestaurant handles POST /api/restaurants
func (h *RestaurantHandler) CreateRestaurant(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	var req dto.CreateRestaurantRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	data := &dto.CreateRestaurantData{
		Request: &req,
		UserID:  claims.UserID,
	}

	created, err := h.service.Create(c.Request.Context(), data)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create restaurant")
		return
	}

	utils.WriteCreated(c.Writer, created)
}

// GetRestaurant handles GET /api/restaurants/{id}
func (h *RestaurantHandler) GetRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid restaurant ID format")
		return
	}

	restaurant, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.WriteNotFound(c.Writer, "Restaurant not found")
		return
	}

	utils.WriteSuccess(c.Writer, restaurant)
}

// GetRestaurants handles GET /api/restaurants
func (h *RestaurantHandler) GetRestaurants(c *gin.Context) {
	restaurants, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to fetch restaurants")
		return
	}

	utils.WriteSuccess(c.Writer, restaurants)
}

// UpdateRestaurant handles PUT /api/restaurants/{id}
func (h *RestaurantHandler) UpdateRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid restaurant ID format")
		return
	}

	var req dto.UpdateRestaurantRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to update restaurant")
		return
	}

	utils.WriteSuccess(c.Writer, updated)
}

// DeleteRestaurant handles DELETE /api/restaurants/{id}
func (h *RestaurantHandler) DeleteRestaurant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid restaurant ID format")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.WriteNotFound(c.Writer, "Restaurant not found")
		return
	}

	utils.WriteNoContent(c.Writer)
}
