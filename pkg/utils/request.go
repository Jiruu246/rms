package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	// Make gin's built-in validator read `validate:"..."` tags instead of its
	// default `binding:"..."`, so request structs declare their rules once and
	// ShouldBind* is the only place validation runs.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.SetTagName("validate")
	}
}

// ParseAndValidateRequest parses the JSON request body into req and validates
// it against its `validate` struct tags.
func ParseAndValidateRequest(c *gin.Context, req any) error {
	if err := c.ShouldBindJSON(req); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			return fmt.Errorf("validation failed: %s", formatValidationErrors(validationErrs))
		}
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	return nil
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, formatValidationError(err))
	}
	return strings.Join(messages, ", ")
}

// formatValidationError formats validation error into a user-friendly message
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", strings.ToLower(err.Field()))
	case "min":
		if err.Type().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters long", strings.ToLower(err.Field()), err.Param())
		}
		return fmt.Sprintf("%s must be at least %s", strings.ToLower(err.Field()), err.Param())
	case "max":
		if err.Type().String() == "string" {
			return fmt.Sprintf("%s must not exceed %s characters", strings.ToLower(err.Field()), err.Param())
		}
		return fmt.Sprintf("%s must not exceed %s", strings.ToLower(err.Field()), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", strings.ToLower(err.Field()))
	}
}
