package teams

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Team represents the core team document in the "teams" collection.
type Team struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Theme       string             `bson:"theme" json:"theme"`
	Description string             `bson:"description,omitempty" json:"description"`
	OwnerID        primitive.ObjectID  `bson:"owner_id" json:"owner_id"`
	OrganizationID *primitive.ObjectID `bson:"org_id,omitempty" json:"org_id,omitempty"`
	ParentTeamID   *primitive.ObjectID `bson:"parent_team_id,omitempty" json:"parent_team_id,omitempty"`
	InviteToken    string              `bson:"invite_token" json:"invite_token"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
