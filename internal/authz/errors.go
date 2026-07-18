package authz

import "errors"

var (
	ErrUnauthenticated = errors.New("authz: unauthenticated")
	ErrForbidden       = errors.New("authz: forbidden")
	ErrNotFound        = errors.New("authz: not found")
)

// AsNotFound maps ErrForbidden to ErrNotFound, for endpoints that should not
// reveal a resource's existence to callers who lack access to it. Other
// errors (including nil) pass through unchanged. See README.md "403 vs 404".
func AsNotFound(err error) error {
	if errors.Is(err, ErrForbidden) {
		return ErrNotFound
	}
	return err
}
