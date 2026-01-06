package handler

import (
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

func (h *UserHandler) GetProfile(c *gin.Context) {
	claims := c.MustGet("claims").(utils.JWTClaims)

	user, err := h.service.GetProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		utils.WriteNotFound(c.Writer, "user not found")
		return
	}

	utils.WriteSuccess(c.Writer, user)
}

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
		utils.WriteInternalError(c.Writer, "Failed to update profile")
		return
	}

	utils.WriteSuccess(c.Writer, user)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		utils.WriteUnauthorized(c.Writer, "unauthorized")
		return
	}

	userID := claims.(utils.JWTClaims).UserID

	if err := h.service.DeleteAccount(c.Request.Context(), userID); err != nil {
		utils.WriteInternalError(c.Writer, err.Error())
		return
	}

	utils.WriteNoContent(c.Writer)
}
