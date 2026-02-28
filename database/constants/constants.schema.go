package constants

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Constant struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       string             `bson:"type" json:"type"`
	UniqueCode string             `bson:"unique_code" json:"unique_code"`
	Meta       any                `bson:"meta" json:"meta"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
