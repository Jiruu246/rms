package handler

import (
	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/authz"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const errInvalidRestaurantIDFormat = "Invalid restaurant ID format"

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
//	@Success		201		{object}	utils.APIResponse[dto.Restaurant]
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
		apperr.WriteHTTPError(c.Writer, err, "Failed to create restaurant")
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
//	@Success		200	{object}	utils.APIResponse[dto.Restaurant]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id} [get]
func (h *RestaurantHandler) GetRestaurant(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	restaurant, err := h.service.GetByID(c.Request.Context(), authz.NewActorFromClaims(claims), id)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to retrieve restaurant")
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
//	@Success		200	{object}	utils.APIResponse[[]dto.Restaurant]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/restaurants [get]
func (h *RestaurantHandler) GetRestaurants(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	restaurants, err := h.service.GetAll(c.Request.Context(), authz.NewActorFromClaims(claims))
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to fetch restaurants")
		return
	}

	utils.WriteSuccess(c.Writer, restaurants)
}

// UpdateRestaurant handles PATCH /api/restaurants/{id}
//
//	@Summary		Update a restaurant's attributes
//	@Description	Updates plain restaurant attributes only. Images are a
//	@Description	separate subresource.
//	@Tags			restaurants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Restaurant ID"	format(uuid)
//	@Param			request	body		dto.UpdateRestaurantRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.Restaurant]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id} [patch]
func (h *RestaurantHandler) UpdateRestaurant(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	var req dto.UpdateRestaurantRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	updated, err := h.service.Update(c.Request.Context(), authz.NewActorFromClaims(claims), id, &req)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to update restaurant")
		return
	}

	utils.WriteSuccess(c.Writer, updated)
}

// CreateRestaurantImageUpload handles
// POST /api/restaurants/{id}/images/{slot}/uploads
//
//	@Summary		Request an upload for a restaurant image slot
//	@Description	Request a presigned upload URL for the named image slot. After the upload is complete, call PUT /restaurants/{id}/images/{slot} with the returned upload_id to attach it.
//	@Tags			restaurants
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Restaurant ID"	format(uuid)
//	@Param			slot	path		string	true	"Image slot"	Enums(logo, cover)
//	@Success		201		{object}	utils.APIResponse[dto.CreateUploadResult]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id}/images/{slot}/uploads [post]
func (h *RestaurantHandler) CreateRestaurantImageUpload(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	slot, ok := dto.ParseRestaurantImageSlot(c.Param("slot"))
	if !ok {
		utils.WriteBadRequest(c.Writer, "Invalid image slot")
		return
	}

	result, err := h.service.CreateImageUpload(c.Request.Context(), authz.NewActorFromClaims(claims), id, slot)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to create restaurant image upload")
		return
	}

	utils.WriteCreated(c.Writer, result)
}

// UpdateRestaurantImage handles PUT /api/restaurants/{id}/images/{slot}
//
//	@Summary		Assign a restaurant image slot
//	@Description	Consumes the given upload (see
//	@Description	POST /restaurants/{id}/images/{slot}/uploads) and assigns
//	@Description	the resulting media asset to the named image slot.
//	@Description	Whatever was previously assigned to this slot, if
//	@Description	anything, is left in storage untouched by this call.
//	@Tags			restaurants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Restaurant ID"	format(uuid)
//	@Param			slot	path		string								true	"Image slot"	Enums(logo, cover)
//	@Param			request	body		dto.UpdateRestaurantImageRequest	true	"Upload to assign"
//	@Success		200		{object}	utils.APIResponse[dto.Restaurant]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		409		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id}/images/{slot} [put]
func (h *RestaurantHandler) UpdateRestaurantImage(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	slot, ok := dto.ParseRestaurantImageSlot(c.Param("slot"))
	if !ok {
		utils.WriteBadRequest(c.Writer, "Invalid image slot")
		return
	}

	var req dto.UpdateRestaurantImageRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	updated, err := h.service.UpdateImage(c.Request.Context(), authz.NewActorFromClaims(claims), id, slot, req.UploadID)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to update restaurant image")
		return
	}

	utils.WriteSuccess(c.Writer, updated)
}

// DeleteRestaurantImage handles DELETE /api/restaurants/{id}/images/{slot}
//
//	@Summary		Clear a restaurant image slot
//	@Description	Detaches whatever media asset is assigned to the named
//	@Description	slot. Idempotent: clearing an already-empty slot succeeds.
//	@Description	The detached asset, if any, is left in storage untouched
//	@Description	by this call.
//	@Tags			restaurants
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"Restaurant ID"	format(uuid)
//	@Param			slot	path		string	true	"Image slot"	Enums(logo, cover)
//	@Success		200		{object}	utils.APIResponse[dto.Restaurant]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/restaurants/{id}/images/{slot} [delete]
func (h *RestaurantHandler) DeleteRestaurantImage(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	slot, ok := dto.ParseRestaurantImageSlot(c.Param("slot"))
	if !ok {
		utils.WriteBadRequest(c.Writer, "Invalid image slot")
		return
	}

	updated, err := h.service.ClearImage(c.Request.Context(), authz.NewActorFromClaims(claims), id, slot)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to clear restaurant image")
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
	claims := c.MustGet("claims").(utils.JWTClaims)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, errInvalidRestaurantIDFormat)
		return
	}

	if err := h.service.Delete(c.Request.Context(), authz.NewActorFromClaims(claims), id); err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to delete restaurant")
		return
	}

	utils.WriteNoContent(c.Writer)
}
