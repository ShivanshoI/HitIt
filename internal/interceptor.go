package internal

import (
	"encoding/json"
	"log"
	"net/http"
)

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
	log.Printf("[RESPONSE] Status: %d, Data: %+v", status, data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}
