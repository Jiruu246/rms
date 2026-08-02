package handler

import (
	"strconv"

	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

type MenuItemHandler struct {
	service services.MenuItemService
}

func NewMenuItemHandler(service services.MenuItemService) *MenuItemHandler {
	return &MenuItemHandler{service: service}
}

// CreateMenuItem handles POST /api/menu-items
//
//	@Summary		Create a menu item
//	@Tags			menu-items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateMenuItemRequest	true	"Menu item details"
//	@Success		201		{object}	utils.APIResponse[dto.MenuItem]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/menu-items [post]
func (h *MenuItemHandler) CreateMenuItem(c *gin.Context) {
	var req dto.CreateMenuItemRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	created, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to create menu item")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetMenuItems handles GET /api/menu-items
//
//	@Summary		List menu items
//	@Tags			menu-items
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	utils.APIResponse[[]dto.MenuItem]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/menu-items [get]
func (h *MenuItemHandler) GetMenuItems(c *gin.Context) {
	items, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to fetch menu items")
		return
	}
	utils.WriteSuccess(c.Writer, items)
}

// GetMenuItem handles GET /api/menu-items/{id}
//
//	@Summary		Get a menu item by ID
//	@Tags			menu-items
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Menu item ID"
//	@Success		200	{object}	utils.APIResponse[dto.MenuItem]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Router			/menu-items/{id} [get]
func (h *MenuItemHandler) GetMenuItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid menu item ID format")
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to retrieve menu item")
		return
	}
	utils.WriteSuccess(c.Writer, item)
}

// UpdateMenuItem handles PATCH /api/menu-items/{id}
//
//	@Summary		Update a menu item
//	@Tags			menu-items
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Menu item ID"
//	@Param			request	body		dto.UpdateMenuItemRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.MenuItem]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/menu-items/{id} [patch]
func (h *MenuItemHandler) UpdateMenuItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid menu item ID format")
		return
	}
	var req dto.UpdateMenuItemRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to update menu item")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteMenuItem handles DELETE /api/menu-items/{id}
//
//	@Summary		Delete a menu item
//	@Tags			menu-items
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Menu item ID"
//	@Success		204	"No Content"
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/menu-items/{id} [delete]
func (h *MenuItemHandler) DeleteMenuItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid menu item ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		apperr.WriteHTTPError(c.Writer, err, "Failed to delete menu item")
		return
	}
	utils.WriteNoContent(c.Writer)
}
