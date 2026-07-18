package apperr

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalid      = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// NotFound builds an ErrNotFound-wrapping error with a caller-supplied,
// client-safe message (e.g. "restaurant %s").
func NotFound(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrNotFound)
}

// Invalid builds an ErrInvalid-wrapping error for request/business input
// that failed validation (e.g. an empty name, a nil UUID).
func Invalid(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrInvalid)
}

// Conflict builds an ErrConflict-wrapping error for state conflicts (e.g. a
// unique constraint violation, a business rule like insufficient stock).
func Conflict(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrConflict)
}

// Unauthorized builds an ErrUnauthorized-wrapping error for a failed
// authentication attempt (e.g. bad credentials, an invalid/expired token).
// The message is client-safe by construction — callers should still avoid
// leaking which part of a credential pair was wrong (e.g. prefer "invalid
// email or password" over "no user with that email").
func Unauthorized(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrUnauthorized)
}
