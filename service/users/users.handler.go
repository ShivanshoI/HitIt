package users

import (
	"encoding/json"
	"log"
	"net/http"

	"pog/database/users"
	"pog/internal"
	"pog/middleware"
)

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// RegisterRoutes attaches the handler's routes to the provided mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+internal.APIPrefix+"/auth/sign-in", h.SignIn)
	mux.HandleFunc("POST "+internal.APIPrefix+"/auth/sign-up", h.SignUp)
	mux.Handle("GET "+internal.APIPrefix+"/auth/me", middleware.Auth(http.HandlerFunc(h.GetMe)))
}

// SignIn handles POST /auth/sign-in
func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	// Call service layer
	user, err := h.service.SignIn(r.Context(), req.Identifier, req.Password)
	if err != nil {
		log.Printf("[HANDLER] SignIn error: %v", err)
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("sign in failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, user)
}

	// SignUp handles POST /auth/sign-up
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var payload users.User
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[HANDLER] JSON decode error: %v", err)
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}
	log.Printf("[HANDLER] SignUp request received for: %v", payload.FirstName)


	// Call service layer
	user, err := h.service.SignUp(r.Context(), &payload)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("sign up failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusCreated, user)
}

// GetMe handles GET /auth/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from auth context
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	// Call service layer
	user, err := h.service.GetMe(r.Context(), userID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("failed to fetch user"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, user)
}
