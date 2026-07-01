package pagination

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

var (
	// ErrInvalidSortField is returned when the client requests a field not present
	// in the entity's SortFieldSpec map.
	ErrInvalidSortField = errors.New("invalid sort field")

	// ErrCursorSortMismatch is returned when the cursor's embedded sort signature
	// differs from the current request's sort parameters.
	ErrCursorSortMismatch = errors.New("cursor does not match requested sort")

	// ErrInvalidCursor is returned when the cursor token cannot be decoded or is malformed.
	ErrInvalidCursor = errors.New("invalid cursor")
)

// QueryExecutor is the function type that Run calls to fetch paginated rows.
//
// The per-entity adapter wraps an ent query builder into this type, applying
// entity-specific type conversions for order options and the cursor predicate.
// filters should already be applied to the underlying ent query before the
// executor is created; the executor only applies pagination concerns.
type QueryExecutor[Row any] func(
	ctx context.Context,
	orders []func(*sql.Selector),
	cursorPred func(*sql.Selector), // nil on the first page
	limit int,
) ([]Row, error)

// Run executes a cursor-paginated query and returns a PageResponse.
//
// It validates sort fields against the whitelist, decodes and validates the cursor
// (including sort-signature matching), builds the ORDER BY and keyset WHERE clauses,
// fetches limit+1 rows, trims to limit, and encodes next_cursor.
//
// Callers should apply a default sort to req.Sort before calling when no sort is
// provided by the client (Run does not apply defaults — unsorted results have no
// stable pagination guarantee).
func Run[Row any](
	ctx context.Context,
	exec QueryExecutor[Row],
	req PageRequest,
	sortFields map[string]SortFieldSpec[Row],
	extractID func(Row) string,
) (*PageResponse[Row], error) {
	sortSpecs := req.Sort

	// 1. Resolve requested sort fields against the whitelist.
	resolved := make([]resolvedField[Row], 0, len(sortSpecs)+1)
	for _, spec := range sortSpecs {
		sf, ok := sortFields[spec.Field]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidSortField, spec.Field)
		}
		var order func(*sql.Selector)
		if spec.Desc {
			order = sf.Desc
		} else {
			order = sf.Asc
		}
		resolved = append(resolved, resolvedField[Row]{
			name:  spec.Field,
			desc:  spec.Desc,
			order: order,
			eq:    sf.Eq,
			lt:    sf.Lt,
			gt:    sf.Gt,
		})
	}

	// Always append id as the implicit tie-breaker (ascending), using raw SQL helpers
	// so the engine stays generic across any UUID-keyed entity.
	resolved = append(resolved, resolvedField[Row]{
		name:  "id",
		desc:  false,
		order: sql.OrderByField("id", sql.OrderAsc()).ToFunc(),
		eq:    func(v any) func(*sql.Selector) { return sql.FieldEQ("id", v) },
		lt:    func(v any) func(*sql.Selector) { return sql.FieldLT("id", v) },
		gt:    func(v any) func(*sql.Selector) { return sql.FieldGT("id", v) },
	})

	// 2. Decode and validate the cursor when present.
	var cursor *Cursor
	if req.Cursor != "" {
		c, err := DecodeCursor(req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
		}
		if !sortSignaturesMatch(sortSpecs, c.Sort) {
			return nil, ErrCursorSortMismatch
		}
		cursor = &c
	}

	// 3. Fill cursor values into the resolved fields.
	if cursor != nil {
		for i := range resolved {
			if resolved[i].name == "id" {
				parsedID, err := uuid.Parse(cursor.ID)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid id %q in cursor", ErrInvalidCursor, cursor.ID)
				}
				resolved[i].value = parsedID
				continue
			}
			rawVal, ok := cursor.Values[resolved[i].name]
			if !ok {
				return nil, fmt.Errorf("%w: missing value for field %q", ErrInvalidCursor, resolved[i].name)
			}
			sf := sortFields[resolved[i].name]
			val, err := sf.Decode(rawVal)
			if err != nil {
				return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidCursor, resolved[i].name, err)
			}
			resolved[i].value = val
		}
	}

	// 4. Build the ORDER BY list and cursor predicate.
	orders := make([]func(*sql.Selector), len(resolved))
	for i, f := range resolved {
		orders[i] = f.order
	}
	cursorPred := buildCursorPredicate(resolved)

	// 5. Execute: fetch limit+1 to detect whether a next page exists.
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	rows, err := exec(ctx, orders, cursorPred, limit+1)
	if err != nil {
		return nil, fmt.Errorf("paginated query: %w", err)
	}

	// 6. Trim to limit and encode next_cursor.
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	if rows == nil {
		rows = make([]Row, 0)
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		nextCursor, err = buildNextCursor(rows[len(rows)-1], sortSpecs, sortFields, extractID)
		if err != nil {
			return nil, fmt.Errorf("build next cursor: %w", err)
		}
	}

	return &PageResponse[Row]{
		Data:       rows,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// sortSignaturesMatch returns true when a and b represent the same ordered sort spec.
func sortSignaturesMatch(a, b []SortSpec) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Field != b[i].Field || a[i].Desc != b[i].Desc {
			return false
		}
	}
	return true
}

// buildNextCursor encodes a Cursor from the last row of the current page.
func buildNextCursor[Row any](
	row Row,
	sortSpecs []SortSpec,
	sortFields map[string]SortFieldSpec[Row],
	extractID func(Row) string,
) (string, error) {
	values := make(map[string]any, len(sortSpecs))
	for _, spec := range sortSpecs {
		sf := sortFields[spec.Field]
		values[spec.Field] = sf.Extract(row)
	}
	c := Cursor{
		Sort:   sortSpecs,
		Values: values,
		ID:     extractID(row),
	}
	return EncodeCursor(c)
}
