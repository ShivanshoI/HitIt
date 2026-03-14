package organizations

import (
	"net/http"

	"pog/internal"
	"pog/middleware"
)

type OrganizationHandler struct {
	service *OrganizationService
}

func NewOrganizationHandler(service *OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{
		service: service,
	}
}

// RegisterRoutes attaches the handler's routes to the provided mux.
func (h *OrganizationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET "+internal.APIPrefix+"/orgs/{id}", middleware.Auth(http.HandlerFunc(h.GetDetails)))
	mux.Handle("POST "+internal.APIPrefix+"/orgs/{id}/verify", middleware.Auth(http.HandlerFunc(h.Verify)))
}

// GetDetails handles GET /api/orgs/{id}
func (h *OrganizationHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	orgID := r.PathValue("id")
	if orgID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("orgId is required"))
		return
	}

	res, err := h.service.GetDetails(r.Context(), userID, orgID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to get organization details"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, res)
}

// Verify handles POST /api/orgs/{id}/verify
func (h *OrganizationHandler) Verify(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	orgID := r.PathValue("id")
	if orgID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("orgId is required"))
		return
	}

	res, err := h.service.Verify(r.Context(), userID, orgID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to verify organization"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, res)
}
