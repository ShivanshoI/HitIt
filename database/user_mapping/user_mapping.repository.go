package user_mapping

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "user_mappings"

type UserMappingRepository struct {
	collection *mongo.Collection
}

func NewUserMappingRepository(db *mongo.Database) *UserMappingRepository {
	return &UserMappingRepository{
		collection: db.Collection(CollectionName),
	}
}

// Create creates a new user mapping record
func (r *UserMappingRepository) Create(ctx context.Context, mapping *UserMapping) error {
	if mapping.ID.IsZero() {
		mapping.ID = primitive.NewObjectID()
	}
	now := time.Now()
	mapping.JoinedAt = now
	mapping.UpdatedAt = now
	
	_, err := r.collection.InsertOne(ctx, mapping)
	return err
}

// FindByID retrieves a mapping by its document ID
func (r *UserMappingRepository) FindByID(ctx context.Context, id string) (*UserMapping, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var mapping UserMapping
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&mapping)
	return &mapping, err
}

// FindByUserAndOrg finds all mappings for a specific user in an organization
func (r *UserMappingRepository) FindByUserAndOrg(ctx context.Context, userID, orgID string) ([]UserMapping, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	oID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": uID, "org_id": oID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []UserMapping
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// FindByUserOrgAndTeam finds a specific team mapping
func (r *UserMappingRepository) FindByUserOrgAndTeam(ctx context.Context, userID, orgID, teamID string) (*UserMapping, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	oID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, err
	}
	tID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	var mapping UserMapping
	err = r.collection.FindOne(ctx, bson.M{
		"user_id": uID,
		"org_id":  oID,
		"team_id": tID,
	}).Decode(&mapping)
	return &mapping, err
}

// ListByOrg retrieves all mappings for an organization
func (r *UserMappingRepository) ListByOrg(ctx context.Context, orgID string) ([]UserMapping, error) {
	oID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"org_id": oID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []UserMapping
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// ListByTeam retrieves all mappings for a specific team
func (r *UserMappingRepository) ListByTeam(ctx context.Context, teamID string) ([]UserMapping, error) {
	tID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"team_id": tID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []UserMapping
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

// UpdateRole updates the role and permissions of a mapping
func (r *UserMappingRepository) UpdateRole(ctx context.Context, id string, role string, permissions []string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx, 
		bson.M{"_id": objID},
		bson.M{
			"$set": bson.M{
				"role":        role,
				"permissions": permissions,
				"updated_at":  time.Now(),
			},
		},
	)
	return err
}

// MarkLastActive updates the last active timestamp
func (r *UserMappingRepository) MarkLastActive(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"last_active_at": time.Now()}},
	)
	return err
}

// Delete removes a mapping
func (r *UserMappingRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// EnsureIndexes creates necessary indexes for performance and uniqueness
func (r *UserMappingRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "org_id", Value: 1},
				{Key: "team_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "org_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "team_id", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
