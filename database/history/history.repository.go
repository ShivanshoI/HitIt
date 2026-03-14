package history

import (
	"context"
	"log"
	"time"

	"pog/internal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "history"

type HistoryRepository struct {
	collection *mongo.Collection
}

func NewHistoryRepository(db *mongo.Database) *HistoryRepository {
	return &HistoryRepository{
		collection: db.Collection(CollectionName),
	}
}

// applyScope dynamically builds the filter for Personal, Team, and Organization modes.
func applyScope(ctx context.Context, filter bson.M) bson.M {
	orgID, hasOrg := ctx.Value(internal.OrgIDKey).(string)
	teamID, hasTeam := ctx.Value(internal.TeamIDKey).(string)

	isOrgMode := hasOrg && orgID != ""
	isTeamMode := hasTeam && teamID != ""

	if isOrgMode {
		// --- ORGANIZATION MODE ---
		objOrgID, _ := primitive.ObjectIDFromHex(orgID)
		filter["org_id"] = objOrgID

		if isTeamMode {
			// Scenario B: Org + Team (find with orgId + teamId)
			objTeamID, _ := primitive.ObjectIDFromHex(teamID)
			filter["team_id"] = objTeamID
			
			// We remove user_id because the resource belongs to the team/org, not an individual
			delete(filter, "user_id")
		} else {
			// Scenario A: Org Only (find with orgId + userId)
			filter["team_id"] = nil
			// (user_id remains in the filter)
		}
	} else if isTeamMode {
		// --- STANDALONE TEAM MODE ---
		objTeamID, _ := primitive.ObjectIDFromHex(teamID)
		filter["team_id"] = objTeamID
		filter["org_id"] = nil
		
		// Remove user_id because the resource belongs to the team
		delete(filter, "user_id") 
	} else {
		// --- PERSONAL MODE ---
		filter["team_id"] = nil
		filter["org_id"] = nil
		// (user_id remains in the filter natively)
	}

	return filter
}

func (r *HistoryRepository) Create(ctx context.Context, history *RequestHistory) (*RequestHistory, error) {
	if history.ID.IsZero() {
		history.ID = primitive.NewObjectID()
	}
	history.ExecutedAt = time.Now()

	log.Printf("[REPO] Attempting to insert history into: %s", r.collection.Name())
	_, err := r.collection.InsertOne(ctx, history)
	if err != nil {
		log.Printf("[REPO] History insert error: %v", err)
		return nil, err
	}
	return history, nil
}

func (r *HistoryRepository) ListByUserID(ctx context.Context, userID string, page int, limit int) ([]RequestHistory, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := applyScope(ctx, bson.M{"user_id": objUserID})
	
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.M{"executed_at": -1})
	if limit > 0 {
		opts.SetLimit(int64(limit))
		opts.SetSkip(int64((page - 1) * limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var historyLogs []RequestHistory
	if err = cursor.All(ctx, &historyLogs); err != nil {
		return nil, 0, err
	}
	if historyLogs == nil {
		historyLogs = []RequestHistory{}
	}
	return historyLogs, total, nil
}

func (r *HistoryRepository) DeleteByID(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *HistoryRepository) DeleteAllByUserID(ctx context.Context, userID string) error {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"user_id": objUserID})
	return err
}

// CountAllByUserID returns the total number of requests ever executed by a user (all time, all teams).
func (r *HistoryRepository) CountAllByUserID(ctx context.Context, userID string) (int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}
	return r.collection.CountDocuments(ctx, bson.M{"user_id": objUserID})
}

func (r *HistoryRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "team_id", Value: 1},
				{Key: "executed_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "team_id", Value: 1},
				{Key: "executed_at", Value: -1},
			},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

