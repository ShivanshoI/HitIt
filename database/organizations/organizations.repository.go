package organizations

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionName = "organizations"

type OrganizationRepository struct {
	collection *mongo.Collection
}

func NewOrganizationRepository(db *mongo.Database) *OrganizationRepository {
	return &OrganizationRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create inserts a new organization into the database
func (r *OrganizationRepository) Create(ctx context.Context, org *Organization) (*Organization, error) {
	org.ID = primitive.NewObjectID()
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, org)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// GetByID retrieves an organization by its ID string
func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*Organization, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var org Organization
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// GetByParams retrieves an organization by custom parameters (like Name or Email)
func (r *OrganizationRepository) GetByParams(ctx context.Context, params bson.M) (*Organization, error) {
	var org Organization
	err := r.collection.FindOne(ctx, params).Decode(&org)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

// List retrieves all organizations
func (r *OrganizationRepository) List(ctx context.Context) ([]Organization, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orgs []Organization
	if err = cursor.All(ctx, &orgs); err != nil {
		return nil, err
	}
	return orgs, nil
}
