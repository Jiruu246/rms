package handler

import (
	"github.com/Jiruu246/rms/internal/apperr"
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
//
//	@Summary		Create a modifier option
//	@Tags			modifier-options
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateModifierOptionRequest	true	"Modifier option details"
//	@Success		201		{object}	utils.APIResponse[dto.ModifierOption]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/modifiers/options [post]
func (h *ModifierOptionHandler) CreateModifierOption(c *gin.Context) {
	var req dto.CreateModifierOptionRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	data := &dto.CreateModifierOptionData{Request: &req}
	created, err := h.service.Create(c.Request.Context(), data)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to create modifier option")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetModifierOption handles GET /api/modifiers/options/{id}
//
//	@Summary		Get a modifier option by ID
//	@Tags			modifier-options
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Modifier option ID"	format(uuid)
//	@Success		200	{object}	utils.APIResponse[dto.ModifierOption]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/modifiers/options/{id} [get]
func (h *ModifierOptionHandler) GetModifierOption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier option ID format")
		return
	}
	option, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to retrieve modifier option")
		return
	}
	utils.WriteSuccess(c.Writer, option)
}

// GetAllModifierOptions handles GET /api/modifiers/options
//
//	@Summary		List modifier options
//	@Tags			modifier-options
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.APIResponse[[]dto.ModifierOption]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/modifiers/options [get]
func (h *ModifierOptionHandler) GetAllModifierOptions(c *gin.Context) {
	options, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to get modifier options")
		return
	}
	utils.WriteSuccess(c.Writer, options)
}

// UpdateModifierOption handles PATCH /api/modifiers/options/{id}
//
//	@Summary		Update a modifier option
//	@Tags			modifier-options
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Modifier option ID"	format(uuid)
//	@Param			request	body		dto.UpdateModifierOptionRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.ModifierOption]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/modifiers/options/{id} [patch]
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
		apperr.WriteHTTPError(c.Writer, err, "Failed to update modifier option")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteModifierOption handles DELETE /api/modifiers/options/{id}
//
//	@Summary		Delete a modifier option
//	@Tags			modifier-options
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Modifier option ID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/modifiers/options/{id} [delete]
func (h *ModifierOptionHandler) DeleteModifierOption(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid modifier option ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to delete modifier option")
		return
	}
	utils.WriteNoContent(c.Writer)
}
