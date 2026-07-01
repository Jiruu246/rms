package pagination_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/Jiruu246/rms/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

type testRow struct {
	id        uuid.UUID
	createdAt time.Time
	name      string
}

func (r *testRow) extractID() string { return r.id.String() }

func newTestRows(n int) []*testRow {
	rows := make([]*testRow, n)
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := range rows {
		rows[i] = &testRow{
			id:        uuid.New(),
			createdAt: base.Add(time.Duration(i) * time.Hour),
			name:      fmt.Sprintf("row-%02d", i),
		}
	}
	return rows
}

// testSortFields returns sort field specs for testRow.
// These use sql package helpers directly — they don't need a real DB.
func testSortFields() map[string]pagination.SortFieldSpec[*testRow] {
	return map[string]pagination.SortFieldSpec[*testRow]{
		"created_at": {
			Asc:     sql.OrderByField("created_at", sql.OrderAsc()).ToFunc(),
			Desc:    sql.OrderByField("created_at", sql.OrderDesc()).ToFunc(),
			Extract: func(r *testRow) any { return r.createdAt },
			Eq:      func(v any) func(*sql.Selector) { return sql.FieldEQ("created_at", v) },
			Lt:      func(v any) func(*sql.Selector) { return sql.FieldLT("created_at", v) },
			Gt:      func(v any) func(*sql.Selector) { return sql.FieldGT("created_at", v) },
			Decode: func(v any) (any, error) {
				s, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("expected string for created_at, got %T", v)
				}
				t, err := time.Parse(time.RFC3339Nano, s)
				if err != nil {
					return nil, fmt.Errorf("parse created_at: %w", err)
				}
				return t, nil
			},
		},
		"name": {
			Asc:     sql.OrderByField("name", sql.OrderAsc()).ToFunc(),
			Desc:    sql.OrderByField("name", sql.OrderDesc()).ToFunc(),
			Extract: func(r *testRow) any { return r.name },
			Eq:      func(v any) func(*sql.Selector) { return sql.FieldEQ("name", v) },
			Lt:      func(v any) func(*sql.Selector) { return sql.FieldLT("name", v) },
			Gt:      func(v any) func(*sql.Selector) { return sql.FieldGT("name", v) },
			Decode: func(v any) (any, error) {
				s, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("expected string for name, got %T", v)
				}
				return s, nil
			},
		},
	}
}

// staticExecutor returns the first `limit` rows from pool, or all if pool is smaller.
func staticExecutor(pool []*testRow) pagination.QueryExecutor[*testRow] {
	return func(_ context.Context, _ []func(*sql.Selector), _ func(*sql.Selector), limit int) ([]*testRow, error) {
		if limit >= len(pool) {
			return pool, nil
		}
		return pool[:limit], nil
	}
}

// capturingExecutor returns all rows up to limit and records the orders/predicate it received.
type captureResult struct {
	orders    []func(*sql.Selector)
	predicate func(*sql.Selector)
}

func capturingExecutor(pool []*testRow, out *captureResult) pagination.QueryExecutor[*testRow] {
	return func(_ context.Context, orders []func(*sql.Selector), pred func(*sql.Selector), limit int) ([]*testRow, error) {
		out.orders = orders
		out.predicate = pred
		n := limit
		if n > len(pool) {
			n = len(pool)
		}
		return pool[:n], nil
	}
}

// ---------------------------------------------------------------------------
// Cursor encode / decode
// ---------------------------------------------------------------------------

func TestEncodeCursor_DecodeCursor_RoundTrip(t *testing.T) {
	original := pagination.Cursor{
		Sort: []pagination.SortSpec{
			{Field: "created_at", Desc: true},
			{Field: "name", Desc: false},
		},
		Values: map[string]any{
			"created_at": "2024-01-01T12:00:00Z",
			"name":       "some-name",
		},
		ID: uuid.New().String(),
	}

	token, err := pagination.EncodeCursor(original)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	decoded, err := pagination.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Sort, decoded.Sort)
	assert.Equal(t, original.Values["name"], decoded.Values["name"])
}

func TestDecodeCursor_InvalidToken(t *testing.T) {
	_, err := pagination.DecodeCursor("not-valid-base64!!")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// ParseSortParam
// ---------------------------------------------------------------------------

func TestParseSortParam(t *testing.T) {
	tests := []struct {
		input   string
		want    []pagination.SortSpec
		wantErr bool
	}{
		{"", nil, false},
		{"name:asc", []pagination.SortSpec{{Field: "name", Desc: false}}, false},
		{"name:desc", []pagination.SortSpec{{Field: "name", Desc: true}}, false},
		{"name", []pagination.SortSpec{{Field: "name", Desc: false}}, false},
		{"created_at:desc,name:asc", []pagination.SortSpec{
			{Field: "created_at", Desc: true},
			{Field: "name", Desc: false},
		}, false},
		{"name:invalid", nil, true},
		{":asc", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := pagination.ParseSortParam(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParsePageRequest
// ---------------------------------------------------------------------------

func TestParsePageRequest_Defaults(t *testing.T) {
	req, err := pagination.ParsePageRequest("", "", "")
	require.NoError(t, err)
	assert.Equal(t, pagination.DefaultLimit, req.Limit)
	assert.Empty(t, req.Cursor)
	assert.Nil(t, req.Sort)
}

func TestParsePageRequest_CapsAtMaxLimit(t *testing.T) {
	req, err := pagination.ParsePageRequest("9999", "", "")
	require.NoError(t, err)
	assert.Equal(t, pagination.MaxLimit, req.Limit)
}

func TestParsePageRequest_InvalidLimit(t *testing.T) {
	_, err := pagination.ParsePageRequest("-1", "", "")
	assert.Error(t, err)

	_, err = pagination.ParsePageRequest("abc", "", "")
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Run — basic scenarios
// ---------------------------------------------------------------------------

func TestRun_EmptyResult(t *testing.T) {
	req := pagination.PageRequest{
		Limit: 10,
		Sort:  []pagination.SortSpec{{Field: "created_at", Desc: false}},
	}
	exec := staticExecutor(nil)

	resp, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	assert.Empty(t, resp.Data)
	assert.False(t, resp.HasMore)
	assert.Empty(t, resp.NextCursor)
}

func TestRun_ExactPage_NoHasMore(t *testing.T) {
	pool := newTestRows(5)
	req := pagination.PageRequest{
		Limit: 5,
		Sort:  []pagination.SortSpec{{Field: "created_at", Desc: false}},
	}
	// executor returns exactly 5 (limit), so limit+1=6 is requested but only 5 exist
	exec := staticExecutor(pool)

	resp, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	assert.Len(t, resp.Data, 5)
	assert.False(t, resp.HasMore)
	assert.Empty(t, resp.NextCursor)
}

func TestRun_HasMore_NextCursorSet(t *testing.T) {
	pool := newTestRows(10)
	req := pagination.PageRequest{
		Limit: 3,
		Sort:  []pagination.SortSpec{{Field: "created_at", Desc: false}},
	}
	// executor receives limit+1 = 4 and pool has 10, so it returns 4 rows
	exec := staticExecutor(pool)

	resp, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	assert.Len(t, resp.Data, 3)
	assert.True(t, resp.HasMore)
	assert.NotEmpty(t, resp.NextCursor)
}

// ---------------------------------------------------------------------------
// Run — next_cursor encodes correct values
// ---------------------------------------------------------------------------

func TestRun_NextCursor_EncodesLastRowValues(t *testing.T) {
	pool := newTestRows(5)
	req := pagination.PageRequest{
		Limit: 3,
		Sort:  []pagination.SortSpec{{Field: "created_at", Desc: false}},
	}
	exec := staticExecutor(pool)

	resp, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	require.NotEmpty(t, resp.NextCursor)

	// Decode the cursor and verify it reflects the third (last) row.
	c, err := pagination.DecodeCursor(resp.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, pool[2].id.String(), c.ID)
	assert.Equal(t, []pagination.SortSpec{{Field: "created_at", Desc: false}}, c.Sort)

	// Values map must contain the created_at of the last row.
	rawCA, ok := c.Values["created_at"]
	require.True(t, ok, "cursor must carry created_at value")

	// JSON round-trips time as string.
	caStr, ok := rawCA.(string)
	require.True(t, ok)
	decoded, err := time.Parse(time.RFC3339Nano, caStr)
	require.NoError(t, err)
	assert.True(t, pool[2].createdAt.Equal(decoded))
}

// ---------------------------------------------------------------------------
// Run — cursor reuse for page 2
// ---------------------------------------------------------------------------

func TestRun_CursorCanBeUsedForPage2(t *testing.T) {
	pool := newTestRows(6)

	sort := []pagination.SortSpec{{Field: "created_at", Desc: false}}
	fields := testSortFields()
	extractID := func(r *testRow) string { return r.extractID() }

	// Page 1
	req1 := pagination.PageRequest{Limit: 3, Sort: sort}
	exec1 := staticExecutor(pool)
	resp1, err := pagination.Run(context.Background(), exec1, req1, fields, extractID)
	require.NoError(t, err)
	require.True(t, resp1.HasMore)
	require.NotEmpty(t, resp1.NextCursor)

	// Page 2 — the executor returns the remaining rows (simulates DB filtering by cursor)
	req2 := pagination.PageRequest{Limit: 3, Sort: sort, Cursor: resp1.NextCursor}
	exec2 := staticExecutor(pool[3:]) // rows 3..5 after cursor
	resp2, err := pagination.Run(context.Background(), exec2, req2, fields, extractID)
	require.NoError(t, err)
	assert.Len(t, resp2.Data, 3)
	assert.False(t, resp2.HasMore)
	assert.Empty(t, resp2.NextCursor)
}

// ---------------------------------------------------------------------------
// Run — tie-breaking: id always appended to orders
// ---------------------------------------------------------------------------

func TestRun_IDTieBreaker_AppendedToOrders(t *testing.T) {
	pool := newTestRows(3)
	var cap captureResult
	req := pagination.PageRequest{
		Limit: 10,
		Sort:  []pagination.SortSpec{{Field: "created_at", Desc: true}},
	}
	exec := capturingExecutor(pool, &cap)

	_, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)

	// One user field + implicit id = 2 order options.
	assert.Len(t, cap.orders, 2, "id tie-breaker must be appended to orders")
	// No cursor → no WHERE predicate.
	assert.Nil(t, cap.predicate)
}

func TestRun_NoSort_OnlyIDInOrders(t *testing.T) {
	pool := newTestRows(2)
	var cap captureResult
	req := pagination.PageRequest{Limit: 10, Sort: nil}
	exec := capturingExecutor(pool, &cap)

	_, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	assert.Len(t, cap.orders, 1, "only id tie-breaker when no sort is specified")
}

// ---------------------------------------------------------------------------
// Run — multi-field mixed-direction sort
// ---------------------------------------------------------------------------

func TestRun_MultiField_MixedDirection(t *testing.T) {
	pool := newTestRows(5)
	var cap captureResult
	req := pagination.PageRequest{
		Limit: 10,
		Sort: []pagination.SortSpec{
			{Field: "created_at", Desc: true},
			{Field: "name", Desc: false},
		},
	}
	exec := capturingExecutor(pool, &cap)

	_, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.NoError(t, err)
	// created_at DESC + name ASC + id ASC = 3 order options
	assert.Len(t, cap.orders, 3)
}

func TestRun_MultiField_CursorPredicateBuilt(t *testing.T) {
	pool := newTestRows(5)
	sort := []pagination.SortSpec{
		{Field: "created_at", Desc: false},
		{Field: "name", Desc: false},
	}
	fields := testSortFields()
	extractID := func(r *testRow) string { return r.extractID() }

	// Get page 1 cursor
	req1 := pagination.PageRequest{Limit: 2, Sort: sort}
	resp1, err := pagination.Run(context.Background(), staticExecutor(pool), req1, fields, extractID)
	require.NoError(t, err)
	require.NotEmpty(t, resp1.NextCursor)

	// Page 2: verify cursor predicate is non-nil (cursor is set)
	var cap captureResult
	req2 := pagination.PageRequest{Limit: 10, Sort: sort, Cursor: resp1.NextCursor}
	_, err = pagination.Run(context.Background(), capturingExecutor(pool[2:], &cap), req2, fields, extractID)
	require.NoError(t, err)
	assert.NotNil(t, cap.predicate, "cursor predicate must be built when cursor is present")
}

// ---------------------------------------------------------------------------
// Run — tie-breaking when primary sort field has duplicate values
// ---------------------------------------------------------------------------

func TestRun_DuplicatePrimarySort_TieBreaksByID(t *testing.T) {
	// All rows share the same created_at — id must distinguish them.
	sharedTime := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	pool := []*testRow{
		{id: uuid.MustParse("00000000-0000-0000-0000-000000000001"), createdAt: sharedTime, name: "a"},
		{id: uuid.MustParse("00000000-0000-0000-0000-000000000002"), createdAt: sharedTime, name: "b"},
		{id: uuid.MustParse("00000000-0000-0000-0000-000000000003"), createdAt: sharedTime, name: "c"},
	}

	sort := []pagination.SortSpec{{Field: "created_at", Desc: false}}
	fields := testSortFields()
	extractID := func(r *testRow) string { return r.extractID() }

	req := pagination.PageRequest{Limit: 2, Sort: sort}
	resp, err := pagination.Run(context.Background(), staticExecutor(pool), req, fields, extractID)
	require.NoError(t, err)
	require.NotEmpty(t, resp.NextCursor)

	// Cursor must carry the last row's id for tie-breaking.
	c, err := pagination.DecodeCursor(resp.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, pool[1].id.String(), c.ID, "cursor id must be the last displayed row's uuid")
}

// ---------------------------------------------------------------------------
// Run — error cases
// ---------------------------------------------------------------------------

func TestRun_InvalidSortField_ReturnsError(t *testing.T) {
	req := pagination.PageRequest{
		Limit: 10,
		Sort:  []pagination.SortSpec{{Field: "unknown_field", Desc: false}},
	}
	_, err := pagination.Run(context.Background(), staticExecutor(nil), req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.Error(t, err)
	assert.True(t, errors.Is(err, pagination.ErrInvalidSortField))
}

func TestRun_CursorSortMismatch_ReturnsError(t *testing.T) {
	pool := newTestRows(5)
	fields := testSortFields()
	extractID := func(r *testRow) string { return r.extractID() }

	// Issue cursor with sort = created_at:asc
	req1 := pagination.PageRequest{Limit: 2, Sort: []pagination.SortSpec{{Field: "created_at", Desc: false}}}
	resp1, err := pagination.Run(context.Background(), staticExecutor(pool), req1, fields, extractID)
	require.NoError(t, err)

	// Reuse the cursor but change sort to name:asc — must fail.
	req2 := pagination.PageRequest{
		Limit:  10,
		Sort:   []pagination.SortSpec{{Field: "name", Desc: false}},
		Cursor: resp1.NextCursor,
	}
	_, err = pagination.Run(context.Background(), staticExecutor(pool), req2, fields, extractID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, pagination.ErrCursorSortMismatch))
}

func TestRun_MalformedCursor_ReturnsError(t *testing.T) {
	req := pagination.PageRequest{
		Limit:  10,
		Sort:   []pagination.SortSpec{{Field: "created_at", Desc: false}},
		Cursor: "this-is-not-a-valid-cursor",
	}
	_, err := pagination.Run(context.Background(), staticExecutor(nil), req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.Error(t, err)
	assert.True(t, errors.Is(err, pagination.ErrInvalidCursor))
}

func TestRun_ExecutorError_Propagated(t *testing.T) {
	execErr := errors.New("db exploded")
	exec := pagination.QueryExecutor[*testRow](func(_ context.Context, _ []func(*sql.Selector), _ func(*sql.Selector), _ int) ([]*testRow, error) {
		return nil, execErr
	})
	req := pagination.PageRequest{Limit: 5}
	_, err := pagination.Run(context.Background(), exec, req, testSortFields(), func(r *testRow) string { return r.extractID() })
	require.Error(t, err)
	assert.ErrorContains(t, err, "db exploded")
}
