package collaborators

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Collaborator struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	Permission   string             `bson:"permission" json:"permission"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
