package handler

import (
	"errors"
	"strings"

	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/services"
	"github.com/Jiruu246/rms/pkg/pagination"
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
//
//	@Summary		Create a category
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateCategoryRequest	true	"Category details"
//	@Success		201		{object}	utils.APIResponse[dto.Category]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/categories [post]
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
//
//	@Summary		Get a category by ID
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Category ID"	format(uuid)
//	@Success		200	{object}	utils.APIResponse[dto.Category]
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/categories/{id} [get]
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

// UpdateCategory handles PUT /api/categories/{id}
//
//	@Summary		Update a category
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Category ID"	format(uuid)
//	@Param			request	body		dto.UpdateCategoryRequest	true	"Fields to update"
//	@Success		200		{object}	utils.APIResponse[dto.Category]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		404		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/categories/{id} [put]
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
//
//	@Summary		Delete a category
//	@Tags			categories
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Category ID"	format(uuid)
//	@Success		204	"No Content"
//	@Failure		400	{object}	utils.APIResponse[any]
//	@Failure		404	{object}	utils.APIResponse[any]
//	@Failure		500	{object}	utils.APIResponse[any]
//	@Router			/categories/{id} [delete]
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

	utils.WriteNoContent(c.Writer)
}

// GetCategories handles GET /api/categories
//
//	@Summary		List categories
//	@Description	Cursor-paginated list of categories. Sort format: "field:asc,field2:desc" (supported fields: display_order, create_time).
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int		false	"Page size"		default(20)
//	@Param			cursor	query		string	false	"Opaque pagination cursor"
//	@Param			sort	query		string	false	"Sort spec, e.g. display_order:asc"
//	@Success		200		{object}	utils.APIResponse[pagination.PageResponse[dto.Category]]
//	@Failure		400		{object}	utils.APIResponse[any]
//	@Failure		500		{object}	utils.APIResponse[any]
//	@Router			/categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	req, err := pagination.ParsePageRequest(c.Query("limit"), c.Query("cursor"), c.Query("sort"))
	if err != nil {
		utils.WriteBadRequest(c.Writer, err.Error())
		return
	}

	page, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidSortField) ||
			errors.Is(err, pagination.ErrCursorSortMismatch) ||
			errors.Is(err, pagination.ErrInvalidCursor) {
			utils.WriteBadRequest(c.Writer, err.Error())
			return
		}
		utils.WriteInternalError(c.Writer, "Failed to retrieve categories")
		return
	}

	utils.WriteSuccess(c.Writer, page)
}
