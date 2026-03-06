package user_activity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityEvent represents a single user activity event stored in the "user_activity" collection.
// Events are created any time a notable action occurs (collection created, team joined, etc.).
type ActivityEvent struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	// Type is a machine-readable enum for the kind of event.
	// Valid values: collection_created, team_joined, request_sent, profile_updated,
	//               password_changed, collection_shared
	Type string `bson:"type" json:"type"`

	// Title is a human-readable description surfaced directly in the UI activity feed.
	Title string `bson:"title" json:"title"`

	// Icon is a semantic hint for the frontend to pick the right SVG/emoji.
	// Valid values: collection, team, request, settings
	Icon string `bson:"icon" json:"icon"`

	CreatedAt time.Time `bson:"created_at" json:"time"`
}
