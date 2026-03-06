package team_feed

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamFeed stores messages and issues in the "team_feed" collection.
type TeamFeed struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	TeamID     primitive.ObjectID   `bson:"team_id" json:"team_id"`
	UserID     primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Type       string               `bson:"type" json:"type"` // "message" | "issue"
	Message    string               `bson:"message" json:"message"`
	Title      string               `bson:"title,omitempty" json:"title,omitempty"`
	Resolved   bool                 `bson:"resolved" json:"resolved"`
	ResolvedBy *primitive.ObjectID  `bson:"resolved_by,omitempty" json:"resolved_by,omitempty"`
	Mentions   []primitive.ObjectID `bson:"mentions,omitempty" json:"mentions,omitempty"`
	CreatedAt  time.Time            `bson:"created_at" json:"created_at"`
}
