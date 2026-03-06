package internal

import (
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

func NewInternalError(msg string) *AppError {
	return &AppError{StatusCode: http.StatusInternalServerError, Message: msg}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{StatusCode: http.StatusUnauthorized, Message: msg}
}

func NewForbidden(msg string) *AppError {
	return &AppError{StatusCode: http.StatusForbidden, Message: msg}
}

func NewConflict(msg string) *AppError {
	return &AppError{StatusCode: http.StatusConflict, Message: msg}
}

