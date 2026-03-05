package activity_feed

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageScope defines the visibility of the message
type MessageScope string

const (
	GroupScope    MessageScope = "group"
	PersonalScope MessageScope = "personal"
)

// MessageType defines the kind of message
type MessageType string

const (
	UserChatType    MessageType = "user_chat"
	IssueType       MessageType = "issue"
	AIAssistantType MessageType = "ai_assistant"
)

// FeedItem explicitly represents a unified message in the activity feed
type FeedItem struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MasterID   primitive.ObjectID `bson:"master_id" json:"master_id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Scope      MessageScope       `bson:"scope" json:"scope"`
	Type       MessageType        `bson:"type" json:"type"`
	Content    string             `bson:"content" json:"content"`
	IsResolved bool               `bson:"is_resolved" json:"is_resolved"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
