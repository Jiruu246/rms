package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// SortSpec describes one sort field and its direction, embedded in every cursor.
type SortSpec struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// Cursor is the internal representation of a page position. It is serialised to
// and from an opaque base64-JSON token; clients never construct or inspect it.
type Cursor struct {
	// Sort is the sort spec active when this cursor was issued.
	// Validated on reuse to detect sort changes between requests.
	Sort []SortSpec `json:"sort"`
	// Values holds the last row's value for each non-id sort field.
	Values map[string]any `json:"values"`
	// ID is the last row's UUID (always used as the tie-breaker).
	ID string `json:"id"`
}

// EncodeCursor serialises a Cursor to a URL-safe base64-encoded JSON token.
func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses a token produced by EncodeCursor.
func DecodeCursor(s string) (Cursor, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("decode cursor: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("parse cursor: %w", err)
	}
	return c, nil
}
