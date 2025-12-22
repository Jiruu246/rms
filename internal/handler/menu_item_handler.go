package handler

import (
	"strconv"

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
func (h *MenuItemHandler) CreateMenuItem(c *gin.Context) {
	var req dto.CreateMenuItemRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}
	created, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create menu item")
		return
	}
	utils.WriteCreated(c.Writer, created)
}

// GetMenuItems handles GET /api/menu-items
func (h *MenuItemHandler) GetMenuItems(c *gin.Context) {
	items, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to fetch menu items")
		return
	}
	utils.WriteSuccess(c.Writer, items)
}

// GetMenuItem handles GET /api/menu-items/{id}
func (h *MenuItemHandler) GetMenuItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid menu item ID format")
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.WriteNotFound(c.Writer, "Menu item not found")
		return
	}
	utils.WriteSuccess(c.Writer, item)
}

// UpdateMenuItem handles PATCH /api/menu-items/{id}
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
		utils.WriteInternalError(c.Writer, "Failed to update menu item")
		return
	}
	utils.WriteSuccess(c.Writer, updated)
}

// DeleteMenuItem handles DELETE /api/menu-items/{id}
func (h *MenuItemHandler) DeleteMenuItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid menu item ID format")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		utils.WriteInternalError(c.Writer, "Failed to delete menu item")
		return
	}
	utils.WriteNoContent(c.Writer)
}
