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
//
//	@Summary		Create a restaurant
//	@Tags			restaurants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateRestaurantRequest	true	"Restaurant details"
//	@Success		201		{object}	utils.APIResponse[dto.RestaurantResponse]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants [post]
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
//
//	@Summary		Get a restaurant by ID
//	@Tags			restaurants
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Restaurant ID"	format(uuid)
//	@Success		200	{object}	utils.APIResponse[dto.RestaurantResponse]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id} [get]
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
//
//	@Summary		List restaurants
//	@Tags			restaurants
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.APIResponse[[]dto.RestaurantResponse]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/restaurants [get]
func (h *RestaurantHandler) GetRestaurants(c *gin.Context) {
	restaurants, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to fetch restaurants")
		return
	}

	utils.WriteSuccess(c.Writer, restaurants)
}

// UpdateRestaurant handles PUT /api/restaurants/{id}
//
//	@Summary		Update a restaurant
//	@Tags			restaurants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Restaurant ID"	format(uuid)
//	@Param			request	body		dto.UpdateRestaurantRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.RestaurantResponse]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id} [put]
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
//
//	@Summary		Delete a restaurant
//	@Tags			restaurants
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Restaurant ID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id} [delete]
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
