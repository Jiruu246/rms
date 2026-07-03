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
//
//	@Summary		Create a modifier
//	@Tags			modifiers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateModifierRequest	true	"Modifier details"
//	@Success		201		{object}	utils.APIResponse[dto.Modifier]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/modifiers [post]
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
//
//	@Summary		Get a modifier by ID
//	@Tags			modifiers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Modifier ID"	format(uuid)
//	@Success		200	{object}	utils.APIResponse[dto.Modifier]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/modifiers/{id} [get]
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
//
//	@Summary		List modifiers
//	@Tags			modifiers
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.APIResponse[[]dto.Modifier]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/modifiers [get]
func (h *ModifierHandler) GetAllModifiers(c *gin.Context) {
	modifiers, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to get modifiers")
		return
	}
	utils.WriteSuccess(c.Writer, modifiers)
}

// UpdateModifier handles PATCH /api/modifiers/{id}
//
//	@Summary		Update a modifier
//	@Tags			modifiers
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Modifier ID"	format(uuid)
//	@Param			request	body		dto.UpdateModifierRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.Modifier]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/modifiers/{id} [patch]
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
//
//	@Summary		Delete a modifier
//	@Tags			modifiers
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Modifier ID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/modifiers/{id} [delete]
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
