package handler

import (
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ModifierHandler struct {
	service services.ModifierService
}

func NewModifierHandler(service services.ModifierService) *ModifierHandler {
	return &ModifierHandler{
		service: service,
	}
}

// CreateModifier handles POST /api/modifiers
func (h *ModifierHandler) CreateModifier(c *gin.Context) {
	var req dto.CreateModifierRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	data := &dto.CreateModifierData{Request: &req}
	created, err := h.service.Create(c.Request.Context(), data)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create modifier")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetModifier handles GET /api/modifiers/{id}
func (h *ModifierHandler) GetModifier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier ID format")
		return
	}
	modifier, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.WriteNotFound(c.Writer, "Modifier not found")
		return
	}
	utils.WriteSuccess(c.Writer, modifier)
}

// GetAllModifiers handles GET /api/modifiers
func (h *ModifierHandler) GetAllModifiers(c *gin.Context) {
	modifiers, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to get modifiers")
		return
	}
	utils.WriteSuccess(c.Writer, modifiers)
}

// UpdateModifier handles PATCH /api/modifiers/{id}
func (h *ModifierHandler) UpdateModifier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier ID format")
		return
	}
	var req dto.UpdateModifierRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to update modifier")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteModifier handles DELETE /api/modifiers/{id}
func (h *ModifierHandler) DeleteModifier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.WriteInternalError(c.Writer, "Failed to delete modifier")
		return
	}
	utils.WriteNoContent(c.Writer)
}
