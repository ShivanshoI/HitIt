package activity_feed

import (
	"encoding/json"
	"net/http"

	pogActivityFeedDB "pog/database/activity_feed"
	"pog/internal"
	"pog/middleware"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityFeedHandler wires HTTP routes to the ActivityFeedService.
type ActivityFeedHandler struct {
	service *ActivityFeedService
}

// NewActivityFeedHandler constructs a handler backed by the given service.
func NewActivityFeedHandler(service *ActivityFeedService) *ActivityFeedHandler {
	return &ActivityFeedHandler{service: service}
}

// RegisterRoutes attaches all activity-feed endpoints to mux.
func (h *ActivityFeedHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.Auth // alias for readability

	mux.Handle("GET "+internal.APIPrefix+"/feed/{masterId}", auth(http.HandlerFunc(h.FetchHistory)))
	mux.Handle("POST "+internal.APIPrefix+"/feed/{masterId}/send", auth(http.HandlerFunc(h.SendMessage)))
	mux.Handle("PATCH "+internal.APIPrefix+"/feed/issue/{id}/resolve", auth(http.HandlerFunc(h.ResolveIssue)))
	mux.Handle("POST "+internal.APIPrefix+"/feed/ai/query", auth(http.HandlerFunc(h.AIQuery)))
}

// FetchHistory godoc
// GET /feed/{masterId}?scope=<scope>
func (h *ActivityFeedHandler) FetchHistory(w http.ResponseWriter, r *http.Request) {
	masterID := r.PathValue("masterId")
	userID := internal.MustUserID(r.Context())
	scope := pogActivityFeedDB.MessageScope(r.URL.Query().Get("scope"))

	if scope == "" {
		scope = pogActivityFeedDB.GroupScope
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	history, total, err := h.service.FetchHistory(r.Context(), masterID, userID, scope, page, limit)
	if err != nil {
		respondErr(w, err)
		return
	}

	respondOK(w, map[string]interface{}{
		"history": history,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// SendMessage godoc
// POST /feed/{masterId}/send
func (h *ActivityFeedHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	masterID := r.PathValue("masterId")
	userID := internal.MustUserID(r.Context())

	var dto SendMessageDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid request body"))
		return
	}

	mID, err := primitive.ObjectIDFromHex(masterID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid master ID"))
		return
	}
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid user ID"))
		return
	}

	if dto.Scope == "" {
		dto.Scope = pogActivityFeedDB.GroupScope
	}

	item := &pogActivityFeedDB.FeedItem{
		MasterID: mID,
		UserID:   uID,
		Type:     dto.Type,
		Content:  dto.Content,
		Scope:    dto.Scope,
	}

	created, err := h.service.SendMessage(r.Context(), item)
	if err != nil {
		respondErr(w, err)
		return
	}

	respondOK(w, created)
}

// ResolveIssue godoc
// PATCH /feed/issue/{id}/resolve?master_id=<masterID>
func (h *ActivityFeedHandler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("id")
	masterID := r.URL.Query().Get("master_id")
	if masterID == "" {
		internal.ErrorResponse(w, internal.NewBadRequest("master_id query param is required"))
		return
	}

	if err := h.service.ResolveIssue(r.Context(), issueID, masterID); err != nil {
		respondErr(w, err)
		return
	}

	respondOK(w, map[string]string{"message": "issue resolved"})
}

// AIQuery godoc
// POST /feed/ai/query
func (h *ActivityFeedHandler) AIQuery(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())

	var dto AIQueryDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid request body"))
		return
	}

	response, err := h.service.AIQuery(r.Context(), &dto, userID)
	if err != nil {
		respondErr(w, err)
		return
	}

	respondOK(w, response)
}

// ---------- helpers ----------

// respondOK writes a JSON success envelope.
func respondOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// respondErr coerces err to an *internal.AppError and writes the appropriate
// HTTP error response.
func respondErr(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*internal.AppError); ok {
		internal.ErrorResponse(w, appErr)
		return
	}
	internal.ErrorResponse(w, internal.NewInternalError(err.Error()))
}