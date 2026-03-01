package collections

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Collection struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MasterID       primitive.ObjectID `bson:"master_id" json:"master_id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name           string             `bson:"name" json:"name"`
	Tags           *[]string          `bson:"tags,omitempty" json:"tags,omitempty"`
	Default_Method string             `bson:"default_method" json:"default_method"`
	Accent_Color   string             `bson:"accent_color" json:"accent_color"`
	Pattern        string             `bson:"pattern" json:"pattern"`
	TotalRequests  int                `bson:"total_requests" json:"total_requests"`
	Favorite       bool               `bson:"favorite" json:"favorite"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
