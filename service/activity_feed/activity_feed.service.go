package activity_feed

import (
	"context"
	"encoding/json"
	"fmt"

	pogActivityFeedDB "pog/database/activity_feed"
	"pog/internal"
	"pog/service/websockets"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityFeedService owns all business logic for the activity feed.
type ActivityFeedService struct {
	repo *pogActivityFeedDB.MessageRepository
	hub  *websockets.Hub
}

// NewActivityFeedService constructs a service with the given repository and hub.
func NewActivityFeedService(repo *pogActivityFeedDB.MessageRepository, hub *websockets.Hub) *ActivityFeedService {
	return &ActivityFeedService{repo: repo, hub: hub}
}

// SendMessage persists item and broadcasts it to all relevant hub clients.
func (s *ActivityFeedService) SendMessage(ctx context.Context, item *pogActivityFeedDB.FeedItem) (*pogActivityFeedDB.FeedItem, error) {
	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, internal.NewInternalError("failed to save message")
	}

	s.broadcast(websockets.BroadcastMessage{
		MasterID: created.MasterID.Hex(),
		UserID:   created.UserID.Hex(),
		Scope:    string(created.Scope),
		Data:     mustMarshal(created),
	})

	return created, nil
}

// FetchHistory returns feed items for the given master and user, filtered by scope and paginated.
func (s *ActivityFeedService) FetchHistory(
	ctx context.Context,
	masterID, userID string,
	scope pogActivityFeedDB.MessageScope,
	page, limit int,
) ([]pogActivityFeedDB.FeedItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	mID, err := primitive.ObjectIDFromHex(masterID)
	if err != nil {
		return nil, 0, internal.NewBadRequest("invalid master ID")
	}
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, internal.NewBadRequest("invalid user ID")
	}

	return s.repo.FetchHistory(ctx, mID, uID, scope, page, limit)
}

// ResolveIssue marks issueID as resolved and notifies group-scope subscribers.
func (s *ActivityFeedService) ResolveIssue(ctx context.Context, issueID, masterID string) error {
	id, err := primitive.ObjectIDFromHex(issueID)
	if err != nil {
		return internal.NewBadRequest("invalid issue ID")
	}
	if err := s.repo.ResolveIssue(ctx, id); err != nil {
		return internal.NewInternalError("failed to resolve issue")
	}

	s.broadcast(websockets.BroadcastMessage{
		MasterID: masterID,
		Scope:    string(pogActivityFeedDB.GroupScope),
		Data: mustMarshal(map[string]interface{}{
			"type":        "update",
			"action":      "resolve",
			"issue_id":    issueID,
			"is_resolved": true,
		}),
	})

	return nil
}

// AIQuery creates the user's prompt message, generates an AI reply, persists
// both, and (for group scope) broadcasts them to connected clients.
func (s *ActivityFeedService) AIQuery(ctx context.Context, query *AIQueryDTO, userID string) (*AIResponseDTO, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user ID")
	}
	mID, err := primitive.ObjectIDFromHex(query.MasterID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid master ID")
	}

	// 1. Persist the user's message.
	userMsg, err := s.repo.Create(ctx, &pogActivityFeedDB.FeedItem{
		MasterID: mID,
		UserID:   uID,
		Scope:    query.Scope,
		Type:     pogActivityFeedDB.UserChatType,
		Content:  query.Prompt,
	})
	if err != nil {
		return nil, internal.NewInternalError("failed to save user message")
	}

	// 2. Generate AI response (TODO: replace stub with real LLM call).
	aiContent := fmt.Sprintf("Mocked AI Response for: %s", query.Prompt)

	// 3. Persist the AI message.
	// UserID is NilObjectID to signal a system/AI origin.
	aiMsg, err := s.repo.Create(ctx, &pogActivityFeedDB.FeedItem{
		MasterID: mID,
		UserID:   primitive.NilObjectID,
		Scope:    query.Scope,
		Type:     pogActivityFeedDB.AIAssistantType,
		Content:  aiContent,
	})
	if err != nil {
		return nil, internal.NewInternalError("failed to save AI message")
	}

	// 4. Broadcast both messages for group-scoped feeds so other connected
	//    clients see the exchange in real time.
	if query.Scope == pogActivityFeedDB.GroupScope {
		s.broadcast(websockets.BroadcastMessage{MasterID: query.MasterID, Scope: string(pogActivityFeedDB.GroupScope), Data: mustMarshal(userMsg)})
		s.broadcast(websockets.BroadcastMessage{MasterID: query.MasterID, Scope: string(pogActivityFeedDB.GroupScope), Data: mustMarshal(aiMsg)})
	}

	return &AIResponseDTO{UserMessage: userMsg, AIMessage: aiMsg}, nil
}

// ---------- private helpers ----------

// broadcast is a thin wrapper that makes call-sites more readable.
func (s *ActivityFeedService) broadcast(msg websockets.BroadcastMessage) {
	s.hub.Broadcast <- msg
}

// mustMarshal serialises v to JSON, panicking only on programmer error
// (i.e. a type that cannot be marshalled). All domain structs are safe.
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic("activity_feed: mustMarshal: " + err.Error())
	}
	return data
}