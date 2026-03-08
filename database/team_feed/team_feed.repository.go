package team_feed

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "team_feed"

type TeamFeedRepository struct {
	collection *mongo.Collection
}

func NewTeamFeedRepository(db *mongo.Database) *TeamFeedRepository {
	return &TeamFeedRepository{
		collection: db.Collection(CollectionName),
	}
}

func (r *TeamFeedRepository) Create(ctx context.Context, item *TeamFeed) (*TeamFeed, error) {
	if item.ID.IsZero() {
		item.ID = primitive.NewObjectID()
	}
	item.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *TeamFeedRepository) ListByTeamID(ctx context.Context, teamID string, page, limit int) ([]TeamFeed, int64, error) {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"team_id": objTeamID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": 1}). // ASC — oldest first (chat-like)
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []TeamFeed
	if err = cursor.All(ctx, &items); err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []TeamFeed{}
	}
	return items, total, nil
}

func (r *TeamFeedRepository) FindByID(ctx context.Context, feedID string) (*TeamFeed, error) {
	objID, err := primitive.ObjectIDFromHex(feedID)
	if err != nil {
		return nil, err
	}
	var item TeamFeed
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeamFeedRepository) Resolve(ctx context.Context, feedID, resolvedByUserID string) error {
	objID, err := primitive.ObjectIDFromHex(feedID)
	if err != nil {
		return err
	}
	objResolvedBy, err := primitive.ObjectIDFromHex(resolvedByUserID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{
			"resolved":    true,
			"resolved_by": objResolvedBy,
		}},
	)
	return err
}

func (r *TeamFeedRepository) DeleteAllByTeamID(ctx context.Context, teamID string) error {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteMany(ctx, bson.M{"team_id": objTeamID})
	return err
}

func (r *TeamFeedRepository) EnsureIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "team_id", Value: 1},
			{Key: "created_at", Value: 1},
		},
	}
	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

