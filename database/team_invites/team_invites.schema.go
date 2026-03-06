package team_invites

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamInvite stores pending email invitations in "team_invites".
type TeamInvite struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID    primitive.ObjectID `bson:"team_id" json:"team_id"`
	Email     string             `bson:"email" json:"email"`
	InvitedBy primitive.ObjectID `bson:"invited_by" json:"invited_by"`
	Status    string             `bson:"status" json:"status"` // "pending" | "accepted"
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at"`
}
