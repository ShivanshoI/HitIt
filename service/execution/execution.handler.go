package execution

import (
	"net/http"
	"pog/internal"
	"pog/middleware"
)

type ExecutionHandler struct {
	service *ExecutionService
}

func NewExecutionHandler(service *ExecutionService) *ExecutionHandler {
	return &ExecutionHandler{
		service: service,
	}
}

func (h *ExecutionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST "+internal.APIPrefix+"/requests/{requestID}/hit", middleware.Auth(http.HandlerFunc(h.Hit)))
	mux.Handle("GET "+internal.APIPrefix+"/history", middleware.Auth(http.HandlerFunc(h.GetHistory)))
	mux.Handle("DELETE "+internal.APIPrefix+"/history", middleware.Auth(http.HandlerFunc(h.ClearHistory)))
}

func (h *ExecutionHandler) Hit(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")
	if requestID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("requestID is required"))
		return
	}

	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	execResult, err := h.service.ExecuteRequest(r.Context(), requestID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("Execution failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, execResult)
}

func (h *ExecutionHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	history, err := h.service.GetHistory(r.Context(), userID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("Failed to get history"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, history)
}

func (h *ExecutionHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	err := h.service.ClearHistory(r.Context(), userID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("Failed to clear history"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]string{"message": "History cleared"})
}
