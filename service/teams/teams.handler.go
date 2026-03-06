package teams

import (
	"encoding/json"
	"net/http"
	"pog/internal"
	"pog/middleware"
	"strconv"
)

type TeamHandler struct {
	service *TeamService
}

func NewTeamHandler(service *TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) RegisterRoutes(mux *http.ServeMux) {
	p := internal.APIPrefix

	// ── Routes with literal path segments MUST be registered before
	//    wildcard {teamID} routes to avoid Go 1.22 mux ambiguity. ──

	// Join via invite link (separate top-level path to avoid mux conflict with {teamID} routes)
	mux.Handle("POST "+p+"/join-team/{token}", middleware.Auth(http.HandlerFunc(h.JoinViaToken)))

	// Teams CRUD
	mux.Handle("GET "+p+"/teams", middleware.Auth(http.HandlerFunc(h.ListMyTeams)))
	mux.Handle("POST "+p+"/teams", middleware.Auth(http.HandlerFunc(h.CreateTeam)))
	mux.Handle("PATCH "+p+"/teams/{teamID}", middleware.Auth(http.HandlerFunc(h.UpdateTeam)))
	mux.Handle("DELETE "+p+"/teams/{teamID}", middleware.Auth(http.HandlerFunc(h.DeleteTeam)))

	// Members
	mux.Handle("GET "+p+"/teams/{teamID}/members", middleware.Auth(http.HandlerFunc(h.ListMembers)))
	mux.Handle("PATCH "+p+"/teams/{teamID}/members/{uid}", middleware.Auth(http.HandlerFunc(h.ChangeRole)))
	mux.Handle("DELETE "+p+"/teams/{teamID}/members/{uid}", middleware.Auth(http.HandlerFunc(h.RemoveMember)))

	// Invites
	mux.Handle("POST "+p+"/teams/{teamID}/invite", middleware.Auth(http.HandlerFunc(h.InviteByEmail)))
	mux.Handle("GET "+p+"/teams/{teamID}/invite-link", middleware.Auth(http.HandlerFunc(h.GetInviteLink)))

	// Feed
	mux.Handle("GET "+p+"/teams/{teamID}/feed", middleware.Auth(http.HandlerFunc(h.GetFeed)))
	mux.Handle("POST "+p+"/teams/{teamID}/feed/send", middleware.Auth(http.HandlerFunc(h.SendFeed)))
	mux.Handle("PATCH "+p+"/teams/{teamID}/feed/{feedID}/resolve", middleware.Auth(http.HandlerFunc(h.ResolveFeed)))
}

// ── Teams CRUD ──────────────────────────────────────────────────────

func (h *TeamHandler) ListMyTeams(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())

	teams, err := h.service.ListMyTeams(r.Context(), userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to list teams"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, teams)
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())

	var dto CreateTeamDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	if msg, ok := dto.IsValid(); !ok {
		internal.ErrorResponse(w, internal.NewBadRequest(msg))
		return
	}

	team, err := h.service.CreateTeam(r.Context(), &dto, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to create team"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusCreated, team)
}

func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	var dto UpdateTeamDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	team, err := h.service.UpdateTeam(r.Context(), teamID, &dto, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to update team"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, team)
}

func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	err := h.service.DeleteTeam(r.Context(), teamID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to delete team"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ── Members ─────────────────────────────────────────────────────────

func (h *TeamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	members, err := h.service.ListMembers(r.Context(), teamID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to list members"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, members)
}

func (h *TeamHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")
	targetUID := r.PathValue("uid")

	var dto ChangeRoleDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	err := h.service.ChangeRole(r.Context(), teamID, targetUID, userID, &dto)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to change role"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]string{
		"user_id": targetUID,
		"role":    dto.Role,
	})
}

func (h *TeamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")
	targetUID := r.PathValue("uid")

	err := h.service.RemoveMember(r.Context(), teamID, targetUID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to remove member"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, map[string]bool{"removed": true})
}

// ── Invites ─────────────────────────────────────────────────────────

func (h *TeamHandler) InviteByEmail(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	var dto InviteDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	if len(dto.Emails) == 0 {
		internal.ErrorResponse(w, internal.NewBadRequest("at least one email is required"))
		return
	}

	result, err := h.service.InviteByEmail(r.Context(), teamID, userID, &dto)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to invite"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, result)
}

func (h *TeamHandler) GetInviteLink(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	link, err := h.service.GetInviteLink(r.Context(), teamID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to get invite link"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, link)
}

func (h *TeamHandler) JoinViaToken(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	token := r.PathValue("token")

	team, err := h.service.JoinViaToken(r.Context(), token, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to join team"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, team)
}

// ── Feed ────────────────────────────────────────────────────────────

func (h *TeamHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	feed, err := h.service.GetFeed(r.Context(), teamID, userID, page, limit)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to get feed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, feed)
}

func (h *TeamHandler) SendFeed(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")

	var dto SendFeedDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		internal.ErrorResponse(w, internal.NewBadRequest("invalid payload"))
		return
	}

	if msg, ok := dto.IsValid(); !ok {
		internal.ErrorResponse(w, internal.NewBadRequest(msg))
		return
	}

	item, err := h.service.SendFeed(r.Context(), teamID, userID, &dto)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to send feed"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusCreated, item)
}

func (h *TeamHandler) ResolveFeed(w http.ResponseWriter, r *http.Request) {
	userID := internal.MustUserID(r.Context())
	teamID := r.PathValue("teamID")
	feedID := r.PathValue("feedID")

	result, err := h.service.ResolveFeed(r.Context(), teamID, feedID, userID)
	if err != nil {
		if appErr, ok := err.(*internal.AppError); ok {
			internal.ErrorResponse(w, appErr)
		} else {
			internal.ErrorResponse(w, internal.NewInternalError("failed to resolve issue"))
		}
		return
	}

	internal.SuccessResponse(w, http.StatusOK, result)
}
