package apperr

import (
	"errors"
	"log"
	"net/http"

	"github.com/Jiruu246/rms/pkg/utils"
)

// WriteHTTPError maps err to an HTTP status and writes it through the
// standard utils.APIResponse envelope. It is the single place that decides
// error -> status code, so handlers never need their own switch/if chain or
// string-match a repo's error message.
func WriteHTTPError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, ErrNotFound):
		utils.WriteNotFound(w, err.Error())

	// ErrForbidden is deliberately rendered as a 404, not a 403: this API never
	// tells an unauthorized caller that a resource exists.
	case errors.Is(err, ErrForbidden):
		utils.WriteNotFound(w, "not found")

	case errors.Is(err, ErrInvalid):
		utils.WriteBadRequest(w, err.Error())

	case errors.Is(err, ErrConflict):
		utils.WriteError(w, http.StatusConflict, err.Error(), nil)

	case errors.Is(err, ErrUnauthorized):
		utils.WriteUnauthorized(w, err.Error())

	default:
		log.Printf("unexpected error: %v", err)
		utils.WriteInternalError(w, fallback)
	}
}
