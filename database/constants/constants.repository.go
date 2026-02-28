package constants

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "constants"

type ConstantRepository struct {
	collection *mongo.Collection
}

func NewConstantRepository(db *mongo.Database) *ConstantRepository {
	return &ConstantRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a new constant into the database
func (r *ConstantRepository) Create(ctx context.Context, constant *Constant) (*Constant, error) {
	if constant.ID.IsZero() {
		constant.ID = primitive.NewObjectID()
	}
	constant.CreatedAt = time.Now()
	constant.UpdatedAt = time.Now()

	log.Printf("[REPO] Inserting constant Type: %s, Code: %s", constant.Type, constant.UniqueCode)
	_, err := r.collection.InsertOne(ctx, constant)
	if err != nil {
		return nil, err
	}
	return constant, nil
}

// GetByUniqueCode finds a constant by its type and unique_code
func (r *ConstantRepository) GetByUniqueCode(ctx context.Context, constType string, uniqueCode string) (*Constant, error) {
	var constant Constant
	filter := bson.M{
		"type":        constType,
		"unique_code": uniqueCode,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&constant)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &constant, nil
}

// ListByType finds all constants of a specific type
func (r *ConstantRepository) ListByType(ctx context.Context, constType string) ([]Constant, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"type": constType})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []Constant
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ListLatestByTypeRaw finds the latest constants of a specific type and returns raw BSON documents
func (r *ConstantRepository) ListLatestByTypeRaw(ctx context.Context, constType string, limit int64) ([]bson.M, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created_at": -1}) // Sort by newest first
	findOptions.SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.M{"type": constType}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []bson.M
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
