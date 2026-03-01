package collections

import (
	"encoding/json"
	"log"
	"net/http"
	"pog/internal"
	"pog/middleware"
	"strconv"
)

type CollectionHandler struct {
	service *CollectionService
}

func NewCollectionHandler(service *CollectionService) *CollectionHandler {
	return &CollectionHandler{
		service: service,
	}
}

func (h *CollectionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST "+internal.APIPrefix+"/collections", middleware.Auth(http.HandlerFunc(h.Create)))
	mux.Handle("GET "+internal.APIPrefix+"/collections", middleware.Auth(http.HandlerFunc(h.List)))
	mux.Handle("PATCH "+internal.APIPrefix+"/collections/{collectionID}/mod/", middleware.Auth(http.HandlerFunc(h.UpdateField)))
}

func (h *CollectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	var payload CreateCollectionDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	if !payload.IsValid() {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid default_method"))
		return
	}

	collection, err := h.service.Create(r.Context(), &payload, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("create collection failed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusCreated, collection)
}

func (h *CollectionHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	collectionsList, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("list collections failed"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, collectionsList)
}

func (h *CollectionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 5
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	result, err := h.service.ListAllCollection(r.Context(), userID, page, limit)
	if err != nil {
		internal.ErrorResponse(w, internal.NewInternalError("list paginated collections failed"))
		return
	}

	internal.SuccessResponse(w, http.StatusOK, result)
}

func (h *CollectionHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	collectionID := r.PathValue("collectionID")
	if collectionID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("collectionID is required"))
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[HANDLER] Collection UpdateField - JSON decode error: %v", err)
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	userID, ok := r.Context().Value(internal.UserIDKey).(string)
	if !ok {
		internal.ErrorResponse(w, internal.NewUnauthorized("unauthorized"))
		return
	}

	col, err := h.service.UpdateFields(r.Context(), collectionID, payload, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("update field failed"))
		}
		return
	}
	
	internal.SuccessResponse(w, http.StatusOK, col)
}
