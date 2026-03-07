package handler

import (
	"time"

	"github.com/Jiruu246/rms/internal/cookies"
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

type AuthHandler struct {
	cookieFactory *cookies.Factory
	service       services.AuthService
}

func NewAuthHandler(cookieFactory *cookies.Factory, service services.AuthService) *AuthHandler {
	return &AuthHandler{
		cookieFactory: cookieFactory,
		service:       service,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginUserRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	accessToken, refreshToken, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		utils.WriteUnauthorized(c.Writer, err.Error())
		return
	}

	cookie := h.cookieFactory.NewRefreshToken(refreshToken.Token, time.Until(refreshToken.ExpiresAt))
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	utils.WriteSuccess(c.Writer, accessToken)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.WriteBadRequest(c.Writer, "refresh token cookie is required")
		return
	}

	response, err := h.service.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		utils.WriteUnauthorized(c.Writer, err.Error())
		return
	}

	utils.WriteSuccess(c.Writer, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")

	if err != nil {
		utils.WriteBadRequest(c.Writer, "refresh token cookie is required")
		return
	}

	if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
		utils.WriteInternalError(c.Writer, "Failed to logout")
		return
	}

	cookie := h.cookieFactory.ExpireRefreshToken()
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	utils.WriteNoContent(c.Writer)
}
