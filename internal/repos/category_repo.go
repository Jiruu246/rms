package repos

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Jiruu246/rms/internal/apperr"
	"github.com/Jiruu246/rms/internal/dto"
	"github.com/Jiruu246/rms/internal/ent"
	"github.com/Jiruu246/rms/internal/ent/category"
	"github.com/Jiruu246/rms/internal/ent/predicate"
	"github.com/Jiruu246/rms/pkg/pagination"
	"github.com/google/uuid"
)

// categorySortFields declares the sortable fields for Category pagination.
// "id" is NOT listed here — the engine appends it automatically as a tie-breaker.
// Each new sortable field also needs a composite DB index (field, id).
var categorySortFields = map[string]pagination.SortFieldSpec[*ent.Category]{
	category.FieldCreateTime: {
		Asc:     category.ByCreateTime(sql.OrderAsc()),
		Desc:    category.ByCreateTime(sql.OrderDesc()),
		Extract: func(r *ent.Category) any { return r.CreateTime },
		Eq:      func(v any) func(*sql.Selector) { return category.CreateTimeEQ(v.(time.Time)) },
		Lt:      func(v any) func(*sql.Selector) { return category.CreateTimeLT(v.(time.Time)) },
		Gt:      func(v any) func(*sql.Selector) { return category.CreateTimeGT(v.(time.Time)) },
		Decode: func(v any) (any, error) {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("create_time: expected string, got %T", v)
			}
			t, err := time.Parse(time.RFC3339Nano, s)
			if err != nil {
				return nil, fmt.Errorf("create_time: %w", err)
			}
			return t, nil
		},
	},
	category.FieldName: {
		Asc:     category.ByName(sql.OrderAsc()),
		Desc:    category.ByName(sql.OrderDesc()),
		Extract: func(r *ent.Category) any { return r.Name },
		Eq:      func(v any) func(*sql.Selector) { return category.NameEQ(v.(string)) },
		Lt:      func(v any) func(*sql.Selector) { return category.NameLT(v.(string)) },
		Gt:      func(v any) func(*sql.Selector) { return category.NameGT(v.(string)) },
		Decode: func(v any) (any, error) {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("name: expected string, got %T", v)
			}
			return s, nil
		},
	},
	category.FieldDisplayOrder: {
		Asc:     category.ByDisplayOrder(sql.OrderAsc()),
		Desc:    category.ByDisplayOrder(sql.OrderDesc()),
		Extract: func(r *ent.Category) any { return r.DisplayOrder },
		Eq:      func(v any) func(*sql.Selector) { return category.DisplayOrderEQ(v.(int)) },
		Lt:      func(v any) func(*sql.Selector) { return category.DisplayOrderLT(v.(int)) },
		Gt:      func(v any) func(*sql.Selector) { return category.DisplayOrderGT(v.(int)) },
		Decode: func(v any) (any, error) {
			// JSON unmarshalling into map[string]any produces float64 for all numbers.
			f, ok := v.(float64)
			if !ok {
				return nil, fmt.Errorf("display_order: expected number, got %T", v)
			}
			if f < 0 || f > math.MaxInt || f != math.Trunc(f) {
				return nil, fmt.Errorf("display_order: expected non-negative integer, got %v", v)
			}
			return int(f), nil
		},
	},
}

var defaultSortFields = []pagination.SortSpec{
	{Field: category.FieldCreateTime, Desc: true},
}

type CategoryRepository interface {
	Create(ctx context.Context, category *dto.CreateCategoryRequest) (*dto.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error)
	List(ctx context.Context, req pagination.PageRequest) (*pagination.PageResponse[*dto.Category], error)
	Update(ctx context.Context, id uuid.UUID, category *dto.UpdateCategoryRequest) (*dto.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepository struct {
	client *ent.Client
}

// CategoryListFilters holds optional filter options for listing categories.
type CategoryListFilters struct {
	RestaurantID *uuid.UUID
	IsActive     *bool
}

func NewEntCategoryRepository(client *ent.Client) CategoryRepository {
	return &categoryRepository{
		client: client,
	}
}

func (r *categoryRepository) Create(ctx context.Context, cat *dto.CreateCategoryRequest) (*dto.Category, error) {
	created, err := r.client.Category.
		Create().
		SetName(cat.Name).
		SetDescription(cat.Description).
		SetDisplayOrder(cat.DisplayOrder).
		SetIsActive(cat.IsActive).
		SetRestaurantID(cat.RestaurantID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &dto.Category{
		ID:           created.ID,
		Name:         created.Name,
		Description:  created.Description,
		DisplayOrder: created.DisplayOrder,
		IsActive:     created.IsActive,
	}, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.Category, error) {
	cat, err := r.client.Category.
		Query().
		Where(category.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperr.NotFound("category %s", id)
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &dto.Category{
		ID:           cat.ID,
		Name:         cat.Name,
		Description:  cat.Description,
		DisplayOrder: cat.DisplayOrder,
		IsActive:     cat.IsActive,
	}, nil
}

func (r *categoryRepository) List(ctx context.Context, req pagination.PageRequest) (*pagination.PageResponse[*dto.Category], error) {
	page, err := ListCategories(ctx, r.client, req, CategoryListFilters{})
	if err != nil {
		return nil, err
	}

	data := make([]*dto.Category, len(page.Data))
	for i, cat := range page.Data {
		data[i] = &dto.Category{
			ID:           cat.ID,
			Name:         cat.Name,
			Description:  cat.Description,
			DisplayOrder: cat.DisplayOrder,
			IsActive:     cat.IsActive,
		}
	}

	return &pagination.PageResponse[*dto.Category]{
		Data:       data,
		NextCursor: page.NextCursor,
		PrevCursor: page.PrevCursor,
		HasMore:    page.HasMore,
	}, nil
}

func (r *categoryRepository) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateCategoryRequest) (*dto.Category, error) {
	updateBuilder := r.client.Category.UpdateOneID(id)

	hasUpdates := false

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, apperr.Invalid("name cannot be empty")
		}
		updateBuilder.SetName(*req.Name)
		hasUpdates = true
	}

	if req.Description != nil {
		updateBuilder.SetDescription(*req.Description)
		hasUpdates = true
	}

	if req.DisplayOrder != nil {
		updateBuilder.SetDisplayOrder(*req.DisplayOrder)
		hasUpdates = true
	}

	if req.IsActive != nil {
		updateBuilder.SetIsActive(*req.IsActive)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, apperr.Invalid("no valid fields provided for update")
	}

	updatedCat, err := updateBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &dto.Category{
		ID:           updatedCat.ID,
		Name:         updatedCat.Name,
		Description:  updatedCat.Description,
		DisplayOrder: updatedCat.DisplayOrder,
		IsActive:     updatedCat.IsActive,
	}, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.Category.
		DeleteOneID(id).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperr.NotFound("category %s", id)
		}
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ListCategories executes a cursor-paginated category query with optional filters.
//
// When req.Sort is empty, the default sort is create_time DESC (most recent first).
// Filters are orthogonal to pagination — they are applied before the cursor predicate
// and must remain stable across pages (changing filters between pages is undefined).
func ListCategories(ctx context.Context, client *ent.Client, req pagination.PageRequest, filters CategoryListFilters) (*pagination.PageResponse[*ent.Category], error) {
	if len(req.Sort) == 0 {
		req.Sort = defaultSortFields
	}

	q := client.Category.Query()
	if filters.RestaurantID != nil {
		q = q.Where(category.RestaurantIDEQ(*filters.RestaurantID))
	}
	if filters.IsActive != nil {
		q = q.Where(category.IsActiveEQ(*filters.IsActive))
	}

	return pagination.Run(ctx, newCategoryQueryExecutor(q), req, categorySortFields, func(r *ent.Category) string {
		return r.ID.String()
	})
}

// newCategoryQueryExecutor wraps an *ent.CategoryQuery for use with pagination.Run.
// Apply all business-logic filters to q before calling this; the executor only
// applies ORDER BY and the cursor's keyset WHERE predicate.
func newCategoryQueryExecutor(q *ent.CategoryQuery) pagination.QueryExecutor[*ent.Category] {
	return func(ctx context.Context, orders []func(*sql.Selector), cursorPred func(*sql.Selector), limit int) ([]*ent.Category, error) {
		catOrders := make([]category.OrderOption, len(orders))
		for i, o := range orders {
			// category.OrderOption and func(*sql.Selector) share the same underlying type.
			catOrders[i] = category.OrderOption(o)
		}
		if len(catOrders) > 0 {
			q = q.Order(catOrders...)
		}
		if cursorPred != nil {
			q = q.Where(predicate.Category(cursorPred))
		}
		return q.Limit(limit).All(ctx)
	}
}
