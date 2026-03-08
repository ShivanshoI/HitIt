package teams

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionName = "teams"

type TeamsRepository struct {
	collection *mongo.Collection
}

func NewTeamsRepository(db *mongo.Database) *TeamsRepository {
	return &TeamsRepository{
		collection: db.Collection(CollectionName),
	}
}

func (r *TeamsRepository) Create(ctx context.Context, team *Team) (*Team, error) {
	if team.ID.IsZero() {
		team.ID = primitive.NewObjectID()
	}
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, team)
	if err != nil {
		log.Printf("[REPO] Teams insert error: %v", err)
		return nil, err
	}
	return team, nil
}

func (r *TeamsRepository) FindID(ctx context.Context, id string) (*Team, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var team Team
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamsRepository) FindByInviteToken(ctx context.Context, token string) (*Team, error) {
	var team Team
	err := r.collection.FindOne(ctx, bson.M{"invite_token": token}).Decode(&team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamsRepository) Update(ctx context.Context, id string, updateData bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	updateData["updated_at"] = time.Now()
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	return err
}

func (r *TeamsRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *TeamsRepository) EnsureIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "invite_token", Value: 1}},
	}
	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}
