package repos

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
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
	"create_time": {
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
	"name": {
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
	"display_order": {
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
			return int(f), nil
		},
	},
}

// NewCategoryQueryExecutor wraps an *ent.CategoryQuery for use with pagination.Run.
// Apply all business-logic filters to q before calling this; the executor only
// applies ORDER BY and the cursor's keyset WHERE predicate.
func NewCategoryQueryExecutor(q *ent.CategoryQuery) pagination.QueryExecutor[*ent.Category] {
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

// CategoryListFilters holds optional filter options for listing categories.
type CategoryListFilters struct {
	RestaurantID *uuid.UUID
	IsActive     *bool
}

// ListCategories executes a cursor-paginated category query with optional filters.
//
// When req.Sort is empty, the default sort is create_time DESC (most recent first).
// Filters are orthogonal to pagination — they are applied before the cursor predicate
// and must remain stable across pages (changing filters between pages is undefined).
//
// Example HTTP wiring:
//
//	req, err := pagination.ParsePageRequest(c.Query("limit"), c.Query("cursor"), c.Query("sort"))
//	filters := repos.CategoryListFilters{RestaurantID: &restaurantID}
//	resp, err := repos.ListCategories(ctx, client, req, filters)
func ListCategories(ctx context.Context, client *ent.Client, req pagination.PageRequest, filters CategoryListFilters) (*pagination.PageResponse[*ent.Category], error) {
	if len(req.Sort) == 0 {
		req.Sort = []pagination.SortSpec{{Field: "create_time", Desc: true}}
	}

	q := client.Category.Query()
	if filters.RestaurantID != nil {
		q = q.Where(category.RestaurantIDEQ(*filters.RestaurantID))
	}
	if filters.IsActive != nil {
		q = q.Where(category.IsActiveEQ(*filters.IsActive))
	}

	return pagination.Run(ctx, NewCategoryQueryExecutor(q), req, categorySortFields, func(r *ent.Category) string {
		return r.ID.String()
	})
}
