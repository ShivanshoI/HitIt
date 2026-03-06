package history

import (
	"context"
	"log"
	"time"

	"pog/internal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "history"

type HistoryRepository struct {
	collection *mongo.Collection
}

func NewHistoryRepository(db *mongo.Database) *HistoryRepository {
	return &HistoryRepository{
		collection: db.Collection(CollectionName),
	}
}

// applyTeamScope injects the X-Team-Id into the MongoDB filter.
func applyTeamScope(ctx context.Context, filter bson.M) bson.M {
	teamID, ok := ctx.Value(internal.TeamIDKey).(string)
	if ok && teamID != "" {
		objID, err := primitive.ObjectIDFromHex(teamID)
		if err == nil {
			filter["team_id"] = objID
			// In team scope, we show all history for the team
			delete(filter, "user_id") 
		}
	} else {
		// Personal scope: ensure team_id does not exist
		filter["team_id"] = nil
	}
	return filter
}

func (r *HistoryRepository) Create(ctx context.Context, history *RequestHistory) (*RequestHistory, error) {
	if history.ID.IsZero() {
		history.ID = primitive.NewObjectID()
	}
	history.ExecutedAt = time.Now()

	log.Printf("[REPO] Attempting to insert history into: %s", r.collection.Name())
	_, err := r.collection.InsertOne(ctx, history)
	if err != nil {
		log.Printf("[REPO] History insert error: %v", err)
		return nil, err
	}
	return history, nil
}

func (r *HistoryRepository) ListByUserID(ctx context.Context, userID string, page int, limit int) ([]RequestHistory, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := applyTeamScope(ctx, bson.M{"user_id": objUserID})
	
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.M{"executed_at": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64((page - 1) * limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var historyLogs []RequestHistory
	if err = cursor.All(ctx, &historyLogs); err != nil {
		return nil, 0, err
	}
	if historyLogs == nil {
		historyLogs = []RequestHistory{}
	}
	return historyLogs, total, nil
}

func (r *HistoryRepository) DeleteByID(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *HistoryRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"user_id": objUserID})
	return err
}
