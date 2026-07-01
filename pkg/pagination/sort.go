package pagination

import "entgo.io/ent/dialect/sql"

// SortFieldSpec declares how to paginate by a specific field on entity Row.
// Declare one entry per client-visible sortable field in the entity's sort map.
// Do NOT include "id" — the engine appends it automatically as a tie-breaker.
type SortFieldSpec[Row any] struct {
	// Asc and Desc are the ent ORDER-BY functions for each direction.
	// Build them with the generated helpers, e.g. category.ByCreateTime(sql.OrderAsc()).
	Asc  func(*sql.Selector)
	Desc func(*sql.Selector)

	// Extract returns the field's value from a result row, used to encode the next cursor.
	Extract func(row Row) any

	// Eq, Lt, Gt build raw WHERE predicates for the keyset comparison.
	// v is guaranteed to be the Go type returned by Decode.
	Eq func(v any) func(*sql.Selector)
	Lt func(v any) func(*sql.Selector)
	Gt func(v any) func(*sql.Selector)

	// Decode converts the JSON-decoded value (from cursor.Values) back to the correct Go type.
	// JSON unmarshalling of map[string]any always produces string, float64, bool, etc.,
	// so this function must re-parse (e.g. parse time.Time from string, int from float64).
	Decode func(v any) (any, error)
}

// resolvedField is an internal runtime struct that pairs sort configuration with
// the cursor value for a single sort column.
type resolvedField[Row any] struct {
	name  string
	desc  bool
	order func(*sql.Selector) // ORDER BY function
	value any                 // cursor value; nil on the first page (no cursor)
	eq    func(v any) func(*sql.Selector)
	lt    func(v any) func(*sql.Selector)
	gt    func(v any) func(*sql.Selector)
}
