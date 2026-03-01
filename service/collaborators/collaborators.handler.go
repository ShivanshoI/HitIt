package collaborators

import (
	"encoding/json"
	"net/http"
	"pog/internal"
	"pog/middleware"
)

type CollaboratorHandler struct {
	service *CollaboratorService
}

func NewCollaboratorHandler(service *CollaboratorService) *CollaboratorHandler {
	return &CollaboratorHandler{
		service: service,
	}
}

func (h *CollaboratorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST "+internal.APIPrefix+"/collaborators/import", middleware.Auth(http.HandlerFunc(h.SharedImportByEmail)))
}

func (h *CollaboratorHandler) SharedImportByEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var payload ImportCollaboratorDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	status, err := h.service.ImportDistributer(r.Context(), userID, payload.IDString)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("import failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]string{
		"status":  status,
		"message": "imported successfully",
	})
}