package requests

import (
	"encoding/json"
	"log"
	"net/http"
	"pog/internal"
	"pog/middleware"
)

type RequestHandler struct {
	service *RequestService
}

func NewRequestHandler(service *RequestService) *RequestHandler {
	return &RequestHandler{
		service: service,
	}
}

func (h *RequestHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST "+internal.APIPrefix+"/requests", middleware.Auth(http.HandlerFunc(h.Create)))
	mux.Handle("GET "+internal.APIPrefix+"/requests/collections/{collectionID}", middleware.Auth(http.HandlerFunc(h.ListByCollection)))
	mux.Handle("GET "+internal.APIPrefix+"/requests/{requestID}", middleware.Auth(http.HandlerFunc(h.GetByID)))
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload CreateRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[HANDLER] Request Create - JSON decode error: %v", err)
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	req, err := h.service.Create(r.Context(), &payload, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("create request failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusCreated, req)
}

func (h *RequestHandler) ListByCollection(w http.ResponseWriter, r *http.Request) {
	collectionID := r.PathValue("collectionID")
	if collectionID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("collectionID is required"))
		return
	}

	requestsList, err := h.service.ListByCollection(r.Context(), collectionID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("list requests failed"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, requestsList)
}

func (h *RequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")
	if requestID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("requestID is required"))
		return
	}

	req, err := h.service.GetByID(r.Context(), requestID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("get request failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, req)
}
