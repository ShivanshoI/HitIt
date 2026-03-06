package user_activity

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "user_activity"

type ActivityRepository struct {
	collection *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) *ActivityRepository {
	return &ActivityRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a new activity event for a user.
func (r *ActivityRepository) Create(ctx context.Context, event *ActivityEvent) (*ActivityEvent, error) {
	if event.ID.IsZero() {
		event.ID = primitive.NewObjectID()
	}
	event.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// ListByUserID fetches the most recent activity events for a user, limited to `limit` items.
func (r *ActivityRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]ActivityEvent, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 50 {
		limit = 10
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objUserID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []ActivityEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	if events == nil {
		events = []ActivityEvent{}
	}
	return events, nil
}
