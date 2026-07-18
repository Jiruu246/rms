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

// wrappedError pairs a client-safe message with a sentinel for errors.Is,
// without appending the sentinel's own text to Error() (unlike
// fmt.Errorf("%s: %w", ...), which would leak the internal classification
// into the message returned to API callers).
type wrappedError struct {
	msg    string
	target error
}

func (e *wrappedError) Error() string { return e.msg }
func (e *wrappedError) Unwrap() error { return e.target }

// NotFound builds an ErrNotFound-wrapping error with a caller-supplied,
// client-safe message (e.g. "restaurant %s").
func NotFound(format string, args ...any) error {
	return &wrappedError{msg: fmt.Sprintf(format, args...), target: ErrNotFound}
}

// Invalid builds an ErrInvalid-wrapping error for request/business input
// that failed validation (e.g. an empty name, a nil UUID).
func Invalid(format string, args ...any) error {
	return &wrappedError{msg: fmt.Sprintf(format, args...), target: ErrInvalid}
}

// Conflict builds an ErrConflict-wrapping error for state conflicts (e.g. a
// unique constraint violation, a business rule like insufficient stock).
func Conflict(format string, args ...any) error {
	return &wrappedError{msg: fmt.Sprintf(format, args...), target: ErrConflict}
}

// Unauthorized builds an ErrUnauthorized-wrapping error for a failed
// authentication attempt (e.g. bad credentials, an invalid/expired token).
// The message is client-safe by construction — callers should still avoid
// leaking which part of a credential pair was wrong (e.g. prefer "invalid
// email or password" over "no user with that email").
func Unauthorized(format string, args ...any) error {
	return &wrappedError{msg: fmt.Sprintf(format, args...), target: ErrUnauthorized}
}
