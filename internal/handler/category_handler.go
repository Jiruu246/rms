package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Jiruu246/rms/internal/models"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

// CreateCategory handles POST /api/categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req models.CreateCategoryRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	created, err := h.service.Create(&req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create category")
		return
	}

	utils.WriteCreated(c.Writer, created)
}

// GetCategory handles GET /api/categories/{id}
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	category, err := h.service.GetByID(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteNotFound(c.Writer, "Category not found")
			return
		}
		utils.WriteInternalError(c.Writer, "Failed to retrieve category")
		return
	}

	utils.WriteSuccess(c.Writer, category)
}

// UpdateCategory handles PATCH /api/categories/{id}
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	var req models.UpdateCategoryRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	updated, err := h.service.Update(uint(id), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteNotFound(c.Writer, "Category not found")
			return
		}
		utils.WriteInternalError(c.Writer, "Failed to update category")
		return
	}

	utils.WriteSuccess(c.Writer, updated)
}

// DeleteCategory handles DELETE /api/categories/{id}
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	err = h.service.Delete(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteNotFound(c.Writer, "Category not found")
			return
		}
		utils.WriteInternalError(c.Writer, "Failed to delete category")
		return
	}

	// Return success with no data for DELETE operations
	utils.WriteResponse(c.Writer, http.StatusNoContent, nil)
}
