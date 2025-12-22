package handler

import (
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ModifierOptionHandler struct {
	service services.ModifierOptionService
}

func NewModifierOptionHandler(service services.ModifierOptionService) *ModifierOptionHandler {
	return &ModifierOptionHandler{
		service: service,
	}
}

// CreateModifierOption handles POST /api/modifiers/options
func (h *ModifierOptionHandler) CreateModifierOption(c *gin.Context) {
	var req dto.CreateModifierOptionRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	data := &dto.CreateModifierOptionData{Request: &req}
	created, err := h.service.Create(c.Request.Context(), data)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create modifier option")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetModifierOption handles GET /api/modifiers/options/{id}
func (h *ModifierOptionHandler) GetModifierOption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier option ID format")
		return
	}
	option, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.WriteNotFound(c.Writer, "Modifier option not found")
		return
	}
	utils.WriteSuccess(c.Writer, option)
}

// GetAllModifierOptions handles GET /api/modifiers/options
func (h *ModifierOptionHandler) GetAllModifierOptions(c *gin.Context) {
	options, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to get modifier options")
		return
	}
	utils.WriteSuccess(c.Writer, options)
}

// UpdateModifierOption handles PATCH /api/modifiers/options/{id}
func (h *ModifierOptionHandler) UpdateModifierOption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier option ID format")
		return
	}
	var req dto.UpdateModifierOptionRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to update modifier option")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteModifierOption handles DELETE /api/modifiers/options/{id}
func (h *ModifierOptionHandler) DeleteModifierOption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier option ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.WriteInternalError(c.Writer, "Failed to delete modifier option")
		return
	}
	utils.WriteNoContent(c.Writer)
}
