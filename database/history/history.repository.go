package history

import (
	"context"
	"log"
	"time"

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

	filter := bson.M{"user_id": objUserID}
	
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
