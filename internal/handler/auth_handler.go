package handler

import (
	"time"

	"github.com/Jiruu246/rms/internal/apperr"
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

// Register handles POST /api/auth/register
//
//	@Summary		Register a new user
//	@Description	Creates a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterUserSchema	true	"Registration details"
//	@Success		201		{object}	utils.APIResponse[dto.User]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		409		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/auth/register [post]
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
		apperr.WriteHTTPError(c.Writer, err, "Failed to register")
		return
	}

	utils.WriteCreated(c.Writer, user)
}

// Login handles POST /api/auth/login
//
//	@Summary		Log in
//	@Description	Authenticates a user and returns an access token; sets a refresh token cookie
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginUserRequest	true	"Login credentials"
//	@Success		200		{object}	utils.APIResponse[dto.AccessToken]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		401		{object}	utils.APIResponse[any]
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginUserRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	accessToken, refreshToken, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to log in")
		return
	}

	cookie := h.cookieFactory.NewRefreshToken(refreshToken.Token, time.Until(refreshToken.ExpiresAt))
	c.SetCookieData(cookie)

	utils.WriteSuccess(c.Writer, accessToken)
}

// Refresh handles POST /api/auth/refresh
//
//	@Summary		Refresh access token
//	@Description	Exchanges the refresh token cookie for a new access token
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	utils.APIResponse[dto.AccessToken]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		401	{object}	utils.APIResponse[any]
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		utils.WriteBadRequest(c.Writer, "refresh token cookie is required")
		return
	}

	response, err := h.service.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to refresh access token")
		return
	}

	utils.WriteSuccess(c.Writer, response)
}

// Logout handles POST /api/auth/logout
//
//	@Summary		Log out
//	@Description	Revokes the refresh token and clears the refresh token cookie
//	@Tags			auth
//	@Success		204	{object}	nil
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		401	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")

	if err != nil {
		utils.WriteBadRequest(c.Writer, "refresh token cookie is required")
		return
	}

	if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to logout")
		return
	}

	cookie := h.cookieFactory.ExpireRefreshToken()
	c.SetCookieData(cookie)

	utils.WriteNoContent(c.Writer)
}
