package teams_mapping

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "teams_mapping"

type TeamsMappingRepository struct {
	collection *mongo.Collection
}

func NewTeamsMappingRepository(db *mongo.Database) *TeamsMappingRepository {
	return &TeamsMappingRepository{
		collection: db.Collection(CollectionName),
	}
}

func (r *TeamsMappingRepository) AddMember(ctx context.Context, mapping *TeamMapping) error {
	if mapping.ID.IsZero() {
		mapping.ID = primitive.NewObjectID()
	}
	mapping.JoinedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, mapping)
	return err
}

func (r *TeamsMappingRepository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return false, err
	}
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"team_id": objTeamID,
		"user_id": objUserID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TeamsMappingRepository) FindMember(ctx context.Context, teamID, userID string) (*TeamMapping, error) {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	var mapping TeamMapping
	err = r.collection.FindOne(ctx, bson.M{
		"team_id": objTeamID,
		"user_id": objUserID,
	}).Decode(&mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *TeamsMappingRepository) ListMembersByTeamID(ctx context.Context, teamID string) ([]TeamMapping, error) {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.M{"team_id": objTeamID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []TeamMapping
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	if members == nil {
		members = []TeamMapping{}
	}
	return members, nil
}

func (r *TeamsMappingRepository) ListTeamsByUserID(ctx context.Context, userID string) ([]TeamMapping, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": objUserID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []TeamMapping
	if err = cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	if mappings == nil {
		mappings = []TeamMapping{}
	}
	return mappings, nil
}

func (r *TeamsMappingRepository) CountMembers(ctx context.Context, teamID string) (int64, error) {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return 0, err
	}
	return r.collection.CountDocuments(ctx, bson.M{"team_id": objTeamID})
}

func (r *TeamsMappingRepository) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx,
		bson.M{"team_id": objTeamID, "user_id": objUserID},
		bson.M{"$set": bson.M{"role": role}},
	)
	return err
}

func (r *TeamsMappingRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"team_id": objTeamID, "user_id": objUserID})
	return err
}

func (r *TeamsMappingRepository) DeleteAllByTeamID(ctx context.Context, teamID string) error {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteMany(ctx, bson.M{"team_id": objTeamID})
	return err
}

func (r *TeamsMappingRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "team_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

