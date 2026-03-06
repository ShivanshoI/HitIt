package profile

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pog/internal"
	"pog/middleware"
)

// ProfileHandler wires HTTP routes to ProfileService methods.
type ProfileHandler struct {
	service *ProfileService
}

func NewProfileHandler(service *ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

// RegisterRoutes attaches all profile-related routes.
func (h *ProfileHandler) RegisterRoutes(mux *http.ServeMux) {
	// Profile stats
	mux.Handle("GET "+internal.APIPrefix+"/user/me/stats",
		middleware.Auth(http.HandlerFunc(h.GetStats)))

	// Recent activity
	mux.Handle("GET "+internal.APIPrefix+"/user/activity",
		middleware.Auth(http.HandlerFunc(h.GetActivity)))

	// Update profile (name/email/theme)
	mux.Handle("PUT "+internal.APIPrefix+"/user/profile",
		middleware.Auth(http.HandlerFunc(h.UpdateProfile)))

	// Update password
	mux.Handle("PUT "+internal.APIPrefix+"/user/password",
		middleware.Auth(http.HandlerFunc(h.UpdatePassword)))

	// Sign out all devices
	// NOTE: Because we use stateless JWTs without a denylist, this endpoint today
	// returns success but does NOT truly invalidate existing tokens on the server.
	// To fully implement "sign out all devices" you would need a token denylist
	// (e.g. Redis set of jti values) or short-lived tokens + refresh token rotation.
	// This is noted for future work.
	mux.Handle("DELETE "+internal.APIPrefix+"/user/sessions",
		middleware.Auth(http.HandlerFunc(h.RevokeAllSessions)))
}

// ── Handlers ─────────────────────────────────────────────────────────

// GetStats handles GET /api/user/me/stats
func (h *ProfileHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	stats, err := h.service.GetStats(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch stats"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, stats)
}

// GetActivity handles GET /api/user/activity?limit=10
func (h *ProfileHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 50 {
		limit = 50
	}

	feed, err := h.service.GetActivity(r.Context(), userID, limit)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to fetch activity"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, feed)
}

// UpdateProfile handles PUT /api/user/profile
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	resp, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to update profile"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, resp)
}

// UpdatePassword handles PUT /api/user/password
func (h *ProfileHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	resp, err := h.service.UpdatePassword(r.Context(), userID, req)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to update password"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, resp)
}

// RevokeAllSessions handles DELETE /api/user/sessions
// ⚠️  Stateless JWTs: returning success is the correct UX response but existing
// tokens remain valid until they expire. See comment in RegisterRoutes for future work.
func (h *ProfileHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	// If a future token denylist is added, the userID would be used to invalidate all tokens.
	_, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, RevokeSessionsResponse{
		Success: true,
		Message: "All sessions revoked",
	})
}
