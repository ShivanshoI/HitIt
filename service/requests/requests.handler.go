package requests

import (
	"encoding/json"
	"log"
	"net/http"
	"pog/internal"
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
	mux.HandleFunc("POST "+internal.APIPrefix+"/requests", h.Create)
	mux.HandleFunc("GET "+internal.APIPrefix+"/requests/collection/{collectionID}", h.ListByCollection)
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var payload CreateRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[HANDLER] Request Create - JSON decode error: %v", err)
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	req, err := h.service.Create(r.Context(), &payload)
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
