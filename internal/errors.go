package internal

import (
	"encoding/json"
	"net/http"
)

// AppError represents a structured API error.
type AppError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// ── constructors ────────────────────────────────────────────────────

func NewBadRequest(msg string) *AppError {
	return &AppError{StatusCode: http.StatusBadRequest, Message: msg}
}

func NewNotFound(msg string) *AppError {
	return &AppError{StatusCode: http.StatusNotFound, Message: msg}
}

func NewInternal(msg string) *AppError {
	return &AppError{StatusCode: http.StatusInternalServerError, Message: msg}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{StatusCode: http.StatusUnauthorized, Message: msg}
}

// ── response helper ─────────────────────────────────────────────────

// ErrorResponse writes a JSON error to the response writer.
// Use this in handlers to return consistent error payloads.
func ErrorResponse(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error":   err,
	})
}

// SuccessResponse writes a JSON success payload.
func SuccessResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}
