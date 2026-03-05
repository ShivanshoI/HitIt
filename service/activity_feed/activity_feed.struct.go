package activity_feed

import (
	pogActivityFeedDB "pog/database/activity_feed"
)

// SendMessageDTO carries the payload for a user-initiated message.
type SendMessageDTO struct {
	Type    pogActivityFeedDB.MessageType  `json:"type"`
	Content string                         `json:"content"`
	Scope   pogActivityFeedDB.MessageScope `json:"scope"`
}

// AIQueryDTO carries the payload for an AI assistant query.
type AIQueryDTO struct {
	Prompt   string                         `json:"prompt"`
	Scope    pogActivityFeedDB.MessageScope `json:"scope"`
	MasterID string                         `json:"master_id"`
	Context  map[string]interface{}         `json:"context"`
}

// AIResponseDTO is returned after a successful AI query, containing both
// the echo of the user's message and the generated AI reply.
type AIResponseDTO struct {
	UserMessage *pogActivityFeedDB.FeedItem `json:"user_message"`
	AIMessage   *pogActivityFeedDB.FeedItem `json:"ai_message"`
}