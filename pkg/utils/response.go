package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      T         `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteResponse[T any](w http.ResponseWriter, status int, data T) {
	response := APIResponse[T]{
		Success:   status >= 200 && status < 300,
		Data:      data,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func WriteError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	response := APIResponse[any]{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func WriteSuccess(w http.ResponseWriter, data any) {
	WriteResponse(w, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, data any) {
	WriteResponse(w, http.StatusCreated, data)
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message, nil)
}

func WriteForbidden(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}
