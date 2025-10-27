package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ParseAndValidateRequest parses JSON request body and validates it
func ParseAndValidateRequest(c *gin.Context, req interface{}) error {
	// Bind JSON to struct
	if err := c.ShouldBindJSON(req); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	// Validate struct
	if err := validate.Struct(req); err != nil {
		// Format validation errors nicely
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, formatValidationError(err))
		}
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, ", "))
	}

	return nil
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
