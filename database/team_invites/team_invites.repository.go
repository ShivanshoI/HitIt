package team_invites

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const CollectionName = "team_invites"

type TeamInvitesRepository struct {
	collection *mongo.Collection
}

func NewTeamInvitesRepository(db *mongo.Database) *TeamInvitesRepository {
	return &TeamInvitesRepository{
		collection: db.Collection(CollectionName),
	}
}

func (r *TeamInvitesRepository) Create(ctx context.Context, invite *TeamInvite) error {
	if invite.ID.IsZero() {
		invite.ID = primitive.NewObjectID()
	}
	invite.CreatedAt = time.Now()
	invite.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	_, err := r.collection.InsertOne(ctx, invite)
	return err
}

func (r *TeamInvitesRepository) DeleteAllByTeamID(ctx context.Context, teamID string) error {
	objTeamID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteMany(ctx, bson.M{"team_id": objTeamID})
	return err
}
