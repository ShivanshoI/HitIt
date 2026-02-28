package requests

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KeyValuePair struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

type APIRequest struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	Name         string             `bson:"name" json:"name"`
	Method       string             `bson:"method" json:"method"`
	URL          string             `bson:"url" json:"url"`
	Headers      []KeyValuePair     `bson:"headers" json:"headers"`
	Params       []KeyValuePair     `bson:"params" json:"params"`
	Body         string             `bson:"body" json:"body"`
	Auth         string             `bson:"auth" json:"auth"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type APIRequestSummary struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CollectionID primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	Name         string             `bson:"name" json:"name"`
	Method       string             `bson:"method" json:"method"`
}
