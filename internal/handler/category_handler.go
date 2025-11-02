package handler

import (
	"net/http"
	"strings"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var req dto.CreateCategoryRequest

	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	created, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to create category")
		return
	}

	utils.WriteCreated(c.Writer, created)
}

// GetCategory handles GET /api/categories/{id}
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	category, err := h.service.GetByID(c.Request.Context(), id)
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	var req dto.UpdateCategoryRequest
	if err := utils.ParseAndValidateRequest(c, &req); err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	updated, err := h.service.Update(c.Request.Context(), id, &req)
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.WriteBadRequest(c.Writer, "Invalid category ID format")
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteNotFound(c.Writer, "Category not found")
			return
		}
		utils.WriteInternalError(c.Writer, "Failed to delete category")
		return
	}

	// Return success with no data for DELETE operations
	utils.WriteResponse[any](c.Writer, http.StatusNoContent, nil)
}

// GetCategories handles GET /api/categories
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		utils.WriteInternalError(c.Writer, "Failed to retrieve categories")
		return
	}

	utils.WriteSuccess(c.Writer, categories)
}
