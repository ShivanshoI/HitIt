package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"pog/internal"
)

// ── types ───────────────────────────────────────────────────────────

// statusWriter wraps ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// Auth verifies the JWT token in the Authorization header.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			internal.ErrorResponse(w, internal.NewUnauthorized("Authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			internal.ErrorResponse(w, internal.NewUnauthorized("Invalid Authorization header format"))
			return
		}

		tokenString := parts[1]
		userID, err := internal.ValidateToken(tokenString)
		if err != nil {
			internal.ErrorResponse(w, internal.NewUnauthorized("Invalid or expired token"))
			return
		}

		// Add userID to the request context
		ctx := context.WithValue(r.Context(), internal.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── middleware functions ────────────────────────────────────────────

// Logger logs every incoming request with method, path, status and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}

		next.ServeHTTP(sw, r)

		log.Printf("[%s] %s %d %v", r.Method, r.URL.Path, sw.code, time.Since(start))
	})
}

// CORS sets standard CORS headers and handles OPTIONS preflight.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Recovery catches panics, logs the error, and returns a 500 JSON response.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"success": false,
					"error":   map[string]string{"message": "internal server error"},
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// JSONContentType sets the Content-Type header to application/json.
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// ── chain helper ────────────────────────────────────────────────────

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain composes multiple middlewares into a single middleware.
// Middlewares are applied in the order they are passed.
// e.g. Chain(Recovery, Logger, CORS) → Recovery wraps Logger wraps CORS wraps handler.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
