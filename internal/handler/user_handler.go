package handler

import (
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

type RegisterUserSchema struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserHandler struct {
	service   services.UserService
	jwtSecret []byte
}

func NewUserHandler(service services.UserService, jwtSecret []byte) *UserHandler {
	return &UserHandler{
		service:   service,
		jwtSecret: jwtSecret,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterUserSchema

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	user, err := h.service.Register(c.Request.Context(), services.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to register")
		return
	}

	utils.WriteCreated(c.Writer, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginUserRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	authResponse, err := h.service.Login(c.Request.Context(), req, h.jwtSecret)
	if err != nil {
		utils.WriteUnauthorized(c.Writer, err.Error())
		return
	}

	utils.WriteSuccess(c.Writer, authResponse)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	response, err := h.service.RefreshAccessToken(c.Request.Context(), req.RefreshToken, h.jwtSecret)
	if err != nil {
		utils.WriteUnauthorized(c.Writer, err.Error())
		return
	}

	utils.WriteSuccess(c.Writer, response)
}

func (h *UserHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	err := h.service.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	utils.WriteNoContent(c.Writer)
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
