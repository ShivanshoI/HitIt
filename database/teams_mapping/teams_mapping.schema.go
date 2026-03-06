package teams_mapping

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TeamMapping is the junction document in "teams_mapping" for team↔user membership.
type TeamMapping struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TeamID   primitive.ObjectID `bson:"team_id" json:"team_id"`
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	Role     string             `bson:"role" json:"role"` // "admin" | "member"
	JoinedAt time.Time          `bson:"joined_at" json:"joined_at"`
}
