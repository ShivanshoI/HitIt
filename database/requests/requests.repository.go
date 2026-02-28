package requests

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "requests"

type RequestRepository struct {
	collection *mongo.Collection
}

func NewRequestRepository(db *mongo.Database) *RequestRepository {
	repo := &RequestRepository{
		collection: db.Collection(CollectionName),
	}

	// Ensure performance on ListByCollectionID
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "collection_id", Value: 1}},
	}
	_, err := repo.collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Printf("[REPO] Failed to create index on collection_id: %v", err)
	}

	return repo
}

// Create inserts a new request into the database
func (r *RequestRepository) Create(ctx context.Context, request *APIRequest) (*APIRequest, error) {
	request.CreatedAt = time.Now()
	request.UpdatedAt = time.Now()

	log.Printf("[REPO] Attempting to insert request into: %s", r.collection.Name())
	res, err := r.collection.InsertOne(ctx, request)
	if err != nil {
		log.Printf("[REPO] Insert error: %v", err)
		return nil, err
	}

	// Use MongoDB's auto-generated _id
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		request.ID = oid
	}

	log.Printf("[REPO] Successfully inserted request with ID: %s", request.ID.Hex())
	return request, nil
}

// GetByID retrieves a request by its ID
func (r *RequestRepository) GetByID(ctx context.Context, id string) (*APIRequest, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var request APIRequest
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// ListByCollectionID retrieves all requests for a specific collection
func (r *RequestRepository) ListByCollectionID(ctx context.Context, collectionID string) ([]APIRequest, error) {
	objCollectionID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"collection_id": objCollectionID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []APIRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

// ListSummariesByCollectionID retrieves lightweight request summaries for a specific collection
func (r *RequestRepository) ListSummariesByCollectionID(ctx context.Context, collectionID string) ([]APIRequestSummary, error) {
	objCollectionID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetProjection(bson.M{
		"_id":           1,
		"collection_id": 1,
		"name":          1,
		"method":        1,
	})

	cursor, err := r.collection.Find(ctx, bson.M{"collection_id": objCollectionID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []APIRequestSummary
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, err
	}

	if requests == nil {
		requests = make([]APIRequestSummary, 0)
	}

	return requests, nil
}

// Update modifies an existing request
func (r *RequestRepository) Update(ctx context.Context, id string, request *APIRequest) (*APIRequest, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	request.UpdatedAt = time.Now()
	update := bson.M{
		"$set": request,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	return request, nil
}

// UpdateFields modifies specific fields of a request
func (r *RequestRepository) UpdateFields(ctx context.Context, id string, fields map[string]interface{}) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	fields["updated_at"] = time.Now()
	update := bson.M{
		"$set": fields,
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Delete removes a request from the database
func (r *RequestRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
