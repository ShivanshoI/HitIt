package collections

import (
	"encoding/json"
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
	mux.HandleFunc("GET "+internal.APIPrefix+"/collections/user/{userID}", h.ListByUser)
	mux.Handle("GET "+internal.APIPrefix+"/collections", middleware.Auth(http.HandlerFunc(h.List)))
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

	payload.UserID = userID // Force the userID from the token

	if !payload.IsValid() {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid default_method"))
		return
	}

	collection, err := h.service.Create(r.Context(), &payload)
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
	userID := r.PathValue("userID")
	if userID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("userID is required"))
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
