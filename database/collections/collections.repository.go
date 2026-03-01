package collections

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "collections"

type CollectionRepository struct {
	collection *mongo.Collection
}

func NewCollectionRepository(db *mongo.Database) *CollectionRepository {
	return &CollectionRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a new collection into the database
func (r *CollectionRepository) Create(ctx context.Context, collection *Collection) (*Collection, error) {
	collection.ID = primitive.NewObjectID()
	collection.CreatedAt = time.Now()
	collection.UpdatedAt = time.Now()

	log.Printf("[REPO] Attempting to insert collection into: %s", r.collection.Name())
	_, err := r.collection.InsertOne(ctx, collection)
	if err != nil {
		log.Printf("[REPO] Insert error: %v", err)
		return nil, err
	}
	log.Printf("[REPO] Successfully inserted collection with ID: %s", collection.ID.Hex())
	return collection, nil
}
func (r *CollectionRepository) FindAllByUserID(ctx context.Context, userID string) ([]Collection, error) {
	objId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []Collection
	if err = cursor.All(ctx, &collections); err != nil {
		return nil, err
	}
	return collections, nil
}

// GetByID retrieves a collection by its ID
func (r *CollectionRepository) GetByID(ctx context.Context, id string) (*Collection, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var collection Collection
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&collection)
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

// ListByUserID retrieves all collections for a specific user
func (r *CollectionRepository) ListByUserID(ctx context.Context, userID string) ([]Collection, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objUserID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []Collection
	if err = cursor.All(ctx, &collections); err != nil {
		return nil, err
	}
	return collections, nil
}

// ListPaginatedByUserID retrieves paginated collections for a specific user
func (r *CollectionRepository) ListPaginatedByUserID(ctx context.Context, userID string, page int, limit int) ([]Collection, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := bson.M{"user_id": objUserID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))
	// findOptions.SetSort(bson.M{"created_at": -1}) // Optional: Sort by newest first

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var paginatedCollections []Collection
	if err = cursor.All(ctx, &paginatedCollections); err != nil {
		return nil, 0, err
	}
	return paginatedCollections, total, nil
}

// Update modifies an existing collection
func (r *CollectionRepository) Update(ctx context.Context, id string, collection *Collection) (*Collection, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	collection.UpdatedAt = time.Now()
	update := bson.M{
		"$set": collection,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

// Delete removes a collection from the database
func (r *CollectionRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// UpdateFields updates specific fields of a collection
func (r *CollectionRepository) UpdateFields(ctx context.Context, id string, updateData map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updateData["updated_at"] = time.Now()
	update := bson.M{
		"$set": updateData,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}
