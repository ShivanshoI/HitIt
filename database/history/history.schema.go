package history

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestHistory struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"user_id"`
	RequestID         primitive.ObjectID `bson:"request_id" json:"request_id"`
	CollectionID      primitive.ObjectID `bson:"collection_id" json:"collection_id"`
	Name              string             `bson:"name" json:"name"`
	Method            string             `bson:"method" json:"method"`
	URL               string             `bson:"url" json:"url"`
	StatusCode        int                `bson:"status_code" json:"status_code"`
	ResponseTimeMs    int64              `bson:"response_time_ms" json:"response_time_ms"`
	ResponseSizeBytes int                `bson:"response_size_bytes" json:"response_size_bytes"`
	ExecutedAt        time.Time          `bson:"executed_at" json:"executed_at"`
}
