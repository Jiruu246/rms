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

type AuthHandler struct {
    service   services.AuthService
    jwtSecret []byte
}

func NewAuthHandler(service services.AuthService, jwtSecret []byte) *AuthHandler {
    return &AuthHandler{
        service:   service,
        jwtSecret: jwtSecret,
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

    authResponse, err := h.service.Login(c.Request.Context(), req, h.jwtSecret)
    if err != nil {
        utils.WriteUnauthorized(c.Writer, err.Error())
        return
    }

    utils.WriteSuccess(c.Writer, authResponse)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
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

func (h *AuthHandler) Logout(c *gin.Context) {
    var req dto.LogoutRequest

    if err := utils.ParseAndValidateRequest(c, &req); err != nil {
        utils.WriteBadRequest(c.Writer, err.Error())
        return
    }

    if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
        utils.WriteBadRequest(c.Writer, err.Error())
        return
    }

    utils.WriteNoContent(c.Writer)
}
