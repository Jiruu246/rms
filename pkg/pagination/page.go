package pagination

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// PageRequest holds the pagination parameters parsed from an HTTP request.
type PageRequest struct {
	Limit  int
	Cursor string
	Sort   []SortSpec
}

// PageResponse is the paginated result returned to callers.
type PageResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// ParsePageRequest parses raw query-param strings (limit, cursor, sort) into a PageRequest.
// sort format: "field:asc,field2:desc"
func ParsePageRequest(limitStr, cursorStr, sortStr string) (PageRequest, error) {
	limit := DefaultLimit
	if limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil || n <= 0 {
			return PageRequest{}, fmt.Errorf("invalid limit: must be a positive integer")
		}
		if n > MaxLimit {
			n = MaxLimit
		}
		limit = n
	}

	sort, err := ParseSortParam(sortStr)
	if err != nil {
		return PageRequest{}, err
	}

	return PageRequest{
		Limit:  limit,
		Cursor: cursorStr,
		Sort:   sort,
	}, nil
}

// ParseSortParam parses a comma-separated "field:direction" list into SortSpecs.
// Direction defaults to asc when omitted.
func ParseSortParam(s string) ([]SortSpec, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	specs := make([]SortSpec, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		field := strings.TrimSpace(kv[0])
		if field == "" {
			return nil, fmt.Errorf("invalid sort: empty field name in %q", s)
		}
		spec := SortSpec{Field: field}
		if len(kv) == 2 {
			switch strings.ToLower(strings.TrimSpace(kv[1])) {
			case "desc":
				spec.Desc = true
			case "asc", "":
				spec.Desc = false
			default:
				return nil, fmt.Errorf("invalid sort direction %q for field %q: must be asc or desc", kv[1], field)
			}
		}
		specs = append(specs, spec)
	}
	return specs, nil
}
