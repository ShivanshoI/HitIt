package activity_feed

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "messages"

type MessageRepository struct {
	collection *mongo.Collection
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a single feed item into the messages collection
func (r *MessageRepository) Create(ctx context.Context, item *FeedItem) (*FeedItem, error) {
	if item.ID.IsZero() {
		item.ID = primitive.NewObjectID()
	}
	item.CreatedAt = time.Now()

	// Ensure IsResolved is explicitly false for issues upon creation
	if item.Type == IssueType {
		// keep current value if it was set, else false
	} else {
		item.IsResolved = false
	}

	_, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// FetchHistory retrieves feed items with pagination based on masterID, user (for personal), and scope
func (r *MessageRepository) FetchHistory(ctx context.Context, masterID primitive.ObjectID, userID primitive.ObjectID, scope MessageScope, page int, limit int) ([]FeedItem, int64, error) {
	filter := bson.M{
		"master_id": masterID,
		"scope":     scope,
	}

	// Strict Personal Scoping: If scope is personal, only return messages for that specific UserID
	if scope == PersonalScope {
		filter["user_id"] = userID
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1}) // Latest first
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64((page - 1) * limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []FeedItem
	if err = cursor.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []FeedItem{}
	}
	return items, total, nil
}

// ResolveIssue marks an issue as resolved
func (r *MessageRepository) ResolveIssue(ctx context.Context, issueID primitive.ObjectID) error {
	filter := bson.M{"_id": issueID}
	update := bson.M{
		"$set": bson.M{
			"is_resolved": true,
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
