package handler

import (
	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service services.UserService
}

func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetProfile handles GET /api/users/profile
//
//	@Summary		Get the current user's profile
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.APIResponse[dto.User]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	user, err := h.service.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to retrieve profile")
		return
	}

	utils.WriteSuccess(c.Writer, user)
}

// UpdateProfile handles PUT /api/users/profile
//
//	@Summary		Update the current user's profile
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.UpdateUserRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.User]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		401		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		409		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		utils.WriteUnauthorized(c.Writer, "unauthorized")
		return
	}

	userID := claims.(utils.JWTClaims).UserID

	var updates dto.UpdateUserRequest
	if err := utils.ParseAndValidateRequest(c, &updates); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	user, err := h.service.UpdateProfile(c.Request.Context(), userID, &updates)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to update profile")
		return
	}

	utils.WriteSuccess(c.Writer, user)
}
