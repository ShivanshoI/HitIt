package collections

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

const CollectionName = "collections"

type CollectionRepository struct {
	collection *mongo.Collection
}

func NewCollectionRepository(db *mongo.Database) *CollectionRepository {
	return &CollectionRepository{
		collection: db.Collection(CollectionName),
	}
}

// applyScope dynamically builds the filter for Personal, Team, and Organization modes.
func applyScope(ctx context.Context, filter bson.M, skipUserIDDelete bool) bson.M {
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
			
			// If skipUserIDDelete is false, we remove user_id because the resource belongs to the team/org
			if !skipUserIDDelete {
				delete(filter, "user_id")
			}
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
		
		// If skipUserIDDelete is false, remove user_id because the resource belongs to the team
		if !skipUserIDDelete {
			delete(filter, "user_id")
		}
	} else {
		// --- PERSONAL MODE ---
		filter["team_id"] = nil
		filter["org_id"] = nil
		// (user_id remains in the filter natively)
	}

	return filter
}

// Create inserts a new collection into the database
func (r *CollectionRepository) Create(ctx context.Context, collection *Collection) (*Collection, error) {
	if collection.ID.IsZero() {
		collection.ID = primitive.NewObjectID()
	}
	if collection.MasterID.IsZero() {
		collection.MasterID = collection.ID
	}
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
	filter := applyScope(ctx, bson.M{"user_id": objId}, false)
	
	cursor, err := r.collection.Find(ctx, filter)
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

// FindAllByMasterID retrieves all collections with exactly this master_id
func (r *CollectionRepository) FindAllByMasterID(ctx context.Context, masterID string) ([]Collection, error) {
	objID, err := primitive.ObjectIDFromHex(masterID)
	if err != nil {
		return nil, err
	}

	filter := applyScope(ctx, bson.M{"master_id": objID}, false)

	cursor, err := r.collection.Find(ctx, filter)
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

	filter := applyScope(ctx, bson.M{"_id": objID}, false)

	var collection Collection
	err = r.collection.FindOne(ctx, filter).Decode(&collection)
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (r *CollectionRepository) GetByMasterID(ctx context.Context, id string) (*Collection, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := applyScope(ctx, bson.M{"master_id": objID}, false)

	var collection Collection
	err = r.collection.FindOne(ctx, filter).Decode(&collection)
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

	filter := applyScope(ctx, bson.M{"user_id": objUserID}, false)

	opts := options.Find().SetSort(bson.M{"updated_at": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
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

// listPaginated is the shared implementation — one DB round-trip for both data and count.
func (r *CollectionRepository) listPaginated(
	ctx context.Context,
	filter bson.M,
	page, limit int,
) ([]Collection, int64, error) {
	skip := int64((page - 1) * limit)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$facet", Value: bson.M{
			"data": bson.A{
				bson.M{"$sort": bson.M{"updated_at": -1}},
				bson.M{"$skip": skip},
				bson.M{"$limit": int64(limit)},
			},
			"totalCount": bson.A{
				bson.M{"$count": "count"},
			},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Data       []Collection `bson:"data"`
		TotalCount []struct {
			Count int64 `bson:"count"`
		} `bson:"totalCount"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	if len(results) == 0 {
		return []Collection{}, 0, nil
	}

	var total int64
	if len(results[0].TotalCount) > 0 {
		total = results[0].TotalCount[0].Count
	}
	return results[0].Data, total, nil
}

func (r *CollectionRepository) ListPaginatedByUserID(ctx context.Context, userID string, page, limit int) ([]Collection, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
	filter := applyScope(ctx, bson.M{"user_id": objUserID}, false)
	return r.listPaginated(ctx, filter, page, limit)
}

func (r *CollectionRepository) ListPaginatedSharedByUserID(ctx context.Context, userID string, page, limit int) ([]Collection, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
	filter := applyScope(ctx, bson.M{
		"user_id": objUserID,
		"$expr":   bson.M{"$ne": bson.A{"$_id", "$master_id"}},
	}, false)
	return r.listPaginated(ctx, filter, page, limit)
}

func (r *CollectionRepository) ListPaginatedMineByUserID(ctx context.Context, userID string, page, limit int) ([]Collection, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
	filter := applyScope(ctx, bson.M{"user_id": objUserID}, true)
	return r.listPaginated(ctx, filter, page, limit)
}

func (r *CollectionRepository) ListPaginatedFavByUserID(ctx context.Context, userID string, page, limit int) ([]Collection, int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, 0, err
	}
	filter := applyScope(ctx, bson.M{
		"user_id":  objUserID,
		"favorite": true,
	}, false)
	return r.listPaginated(ctx, filter, page, limit)
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

// CountByUserID returns the total number of collections owned by this user
// (personal scope — no team filter applied, intentionally scoped to user_id only).
func (r *CollectionRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, err
	}
	// We count all collections where user_id matches, regardless of team_id,
	// so stats reflect the true total across both personal and team collections.
	return r.collection.CountDocuments(ctx, bson.M{"user_id": objUserID})
}

func (r *CollectionRepository) CreateMany(ctx context.Context, cols []*Collection) ([]Collection, error) {
	now := time.Now()
	docs := make([]interface{}, len(cols))
	for i, c := range cols {
		if c.ID.IsZero() {
			c.ID = primitive.NewObjectID()
		}
		if c.MasterID.IsZero() {
			c.MasterID = c.ID
		}
		c.CreatedAt = now
		c.UpdatedAt = now
		docs[i] = c
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return nil, err
	}

	result := make([]Collection, len(cols))
	for i, c := range cols {
		result[i] = *c
	}
	return result, nil
}

func (r *CollectionRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "team_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "team_id", Value: 1},
				{Key: "favorite", Value: 1},
				{Key: "updated_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "team_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "team_id", Value: 1},
				{Key: "favorite", Value: 1},
				{Key: "updated_at", Value: -1},
			},
		},
		{
			Keys: bson.D{{Key: "master_id", Value: 1}},
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}