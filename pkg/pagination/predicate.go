package pagination

import "entgo.io/ent/dialect/sql"

// buildCursorPredicate builds the recursive OR/AND keyset WHERE predicate from
// the resolved sort fields. Fields with a nil value (first page) are skipped.
// Returns nil when there is no cursor (first page — no WHERE restriction needed).
//
// For N sort fields [s1, s2, ..., sN] with cursor values [v1, v2, ..., vN] the
// generated predicate is (using cmp = GT for ASC, LT for DESC):
//
//	  cmp(s1, v1)
//	| eq(s1, v1) & cmp(s2, v2)
//	| eq(s1, v1) & eq(s2, v2) & cmp(s3, v3)
//	| ...
//	| eq(s1, v1) & ... & eq(sN-1, vN-1) & cmp(sN, vN)
//
// This is the standard multi-column keyset (seek) predicate.
func buildCursorPredicate[Row any](fields []resolvedField[Row]) func(*sql.Selector) {
	active := make([]resolvedField[Row], 0, len(fields))
	for _, f := range fields {
		if f.value != nil {
			active = append(active, f)
		}
	}
	if len(active) == 0 {
		return nil
	}
	return buildPredicateRecursive(active)
}

// buildPredicateRecursive implements the recursive OR/AND expansion described above.
func buildPredicateRecursive[Row any](fields []resolvedField[Row]) func(*sql.Selector) {
	f := fields[0]

	var cmp func(*sql.Selector)
	if f.desc {
		cmp = f.lt(f.value)
	} else {
		cmp = f.gt(f.value)
	}

	if len(fields) == 1 {
		return cmp
	}

	eq := f.eq(f.value)
	rest := buildPredicateRecursive(fields[1:])

	return sql.OrPredicates(cmp, sql.AndPredicates(eq, rest))
}
