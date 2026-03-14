package user_mapping

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserMapping defines the relationship between a user and their organizational context (Org/Team).
type UserMapping struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID  `bson:"user_id" json:"user_id"`
	OrganizationID primitive.ObjectID  `bson:"org_id" json:"org_id"`
	TeamID         *primitive.ObjectID `bson:"team_id,omitempty" json:"team_id,omitempty"` // Map to specific Team or nil for Org-level mapping
	
	Type           string              `bson:"type" json:"type"`           // "org" | "team"
	Role           string              `bson:"role" json:"role"`           // "owner" | "admin" | "member" | "viewer"
	Permissions    []string            `bson:"permissions" json:"permissions"` // Granular permissions (e.g., "collection:write", "team:invite")
	
	Status         string              `bson:"status" json:"status"`       // "active" | "invited" | "deactivated"
	JoinedAt       time.Time           `bson:"joined_at" json:"joined_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
	LastActiveAt   time.Time           `bson:"last_active_at" json:"last_active_at"` // To track most recently used org/team
}
