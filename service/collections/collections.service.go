package collections

import (
	"context"
	"log"
	"pog/database/collections"
	"pog/database/constants"
	pogRequestsDB "pog/database/requests"
	"pog/database/teams_mapping"
	"pog/internal"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectionService struct {
	repo         *collections.CollectionRepository
	constRepo    *constants.ConstantRepository
	requestsRepo *pogRequestsDB.RequestRepository
	mappingRepo  *teams_mapping.TeamsMappingRepository
	userLocks    sync.Map
}

func NewCollectionService(
	repo *collections.CollectionRepository,
	constRepo *constants.ConstantRepository,
	requestsRepo *pogRequestsDB.RequestRepository,
	mappingRepo *teams_mapping.TeamsMappingRepository,
) *CollectionService {
	return &CollectionService{
		repo:         repo,
		constRepo:    constRepo,
		requestsRepo: requestsRepo,
		mappingRepo:  mappingRepo,
	}
}

func (s *CollectionService) canModify(ctx context.Context, col *collections.Collection, userID string) bool {
	// Always allowed if you own it personally
	if col.UserID.Hex() == userID {
		return true
	}
	// In team scope: allow if caller is an admin/owner of the team
	if col.TeamID != nil {
		role, err := s.mappingRepo.GetMemberRole(ctx, col.TeamID.Hex(), userID)
		if err == nil && (role == "admin" || role == "owner") {
			return true
		}
	}
	return false
}

func (s *CollectionService) Create(ctx context.Context, payload *CreateCollectionDTO, userID string) (*CollectionResponse, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user id")
	}
	collectionModel := &collections.Collection{
		UserID:          userId,
		Name:            payload.Name,
		Tags:            &payload.Tags,
		Default_Method:  payload.Default_Method,
		Accent_Color:    payload.Accent_Color,
		Pattern:         payload.Pattern,
		TotalRequests:   0,
		Favorite:        payload.Favorite,
		WritePermission: true,
	}

	scope := internal.GetScope(ctx)
	if scope.TeamID != "" {
		if objTeamID, err := primitive.ObjectIDFromHex(scope.TeamID); err == nil {
			collectionModel.TeamID = &objTeamID
		}
	}

	if scope.OrgID != "" {
		if objOrgID, err := primitive.ObjectIDFromHex(scope.OrgID); err == nil {
			collectionModel.OrgID = &objOrgID
		}
	}

	collection, err := s.repo.Create(ctx, collectionModel)
	if err != nil {
		return nil, internal.NewInternalError("Failed to create collection")
	}

	tags := []string{}
	if collection.Tags != nil {
		tags = *collection.Tags
	}

	return &CollectionResponse{
		ID:             collection.ID.Hex(),
		MasterID:       collection.MasterID.Hex(),
		Name:           collection.Name,
		Tags:           tags,
		Default_Method: collection.Default_Method,
		Accent_Color:   collection.Accent_Color,
		Pattern:        collection.Pattern,
		TotalRequests:  collection.TotalRequests,
		Favorite:       collection.Favorite,
		CreatedAt:      collection.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      collection.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CollectionService) ListByUser(ctx context.Context, userID string) ([]CollectionResponse, error) {
	collectionsList, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to list collections")
	}

	responses := make([]CollectionResponse, 0)
	for _, col := range collectionsList {
		tags := []string{}
		if col.Tags != nil {
			tags = *col.Tags
		}
		responses = append(responses, CollectionResponse{
			ID:             col.ID.Hex(),
			MasterID:       col.MasterID.Hex(),
			Name:           col.Name,
			Tags:           tags,
			Default_Method: col.Default_Method,
			Accent_Color:   col.Accent_Color,
			Pattern:        col.Pattern,
			TotalRequests:  col.TotalRequests,
			Favorite:       col.Favorite,
			CreatedAt:      col.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      col.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

func (s *CollectionService) ListAllCollection(ctx context.Context, userID string, page, limit int, filter string) (*PaginatedCollectionResponse, error) {

	var collectionsList []collections.Collection
	var total int64
	var err error

	switch filter {
	case "share":
		collectionsList, total, err = s.repo.ListPaginatedSharedByUserID(ctx, userID, page, limit)
	case "fav":
		collectionsList, total, err = s.repo.ListPaginatedFavByUserID(ctx, userID, page, limit)
	case "mine":
		collectionsList, total, err = s.repo.ListPaginatedMineByUserID(ctx, userID, page, limit)
	default:
		collectionsList, total, err = s.repo.ListPaginatedByUserID(ctx, userID, page, limit)
	}
	if err != nil {
		log.Printf("[SERVICE] ListPaginated error: %v (userID: %s, filter: %s)", err, userID, filter)
		return nil, internal.NewInternalError("Failed to list collections")
	}

	if (filter == "" || filter == "all") && len(collectionsList) == 0 {
		s.CreateDefaultCollections(ctx, userID)
		// Re-fetch after creation to get the fresh list with IDs
		switch filter {
		case "share":
			collectionsList, total, err = s.repo.ListPaginatedSharedByUserID(ctx, userID, page, limit)
		case "mine":
			collectionsList, total, err = s.repo.ListPaginatedMineByUserID(ctx, userID, page, limit)
		default:
			collectionsList, total, err = s.repo.ListPaginatedByUserID(ctx, userID, page, limit)
		}
	}

	responses := make([]CollectionResponse, 0)
	for _, col := range collectionsList {
		tags := []string{}
		if col.Tags != nil {
			tags = *col.Tags
		}
		responses = append(responses, CollectionResponse{
			ID:             col.ID.Hex(),
			MasterID:       col.MasterID.Hex(),
			Name:           col.Name,
			Tags:           tags,
			Default_Method: col.Default_Method,
			Accent_Color:   col.Accent_Color,
			Pattern:        col.Pattern,
			TotalRequests:  col.TotalRequests,
			Favorite:       col.Favorite,
			WritePermission: col.WritePermission,
			CreatedAt:      col.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      col.UpdatedAt.Format(time.RFC3339),
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	return &PaginatedCollectionResponse{
		Collections: responses,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
	}, nil
}

func (s *CollectionService) UpdateFields(ctx context.Context, collectionID string, fields map[string]interface{}, userID string) (*CollectionResponse, error) {
	col, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, internal.NewNotFound("Collection not found")
	}

	if !s.canModify(ctx, col, userID) {
		return nil, internal.NewUnauthorized("Unauthorized to modify this collection")
	}

	allowedFields := map[string]bool{
		"name":           true,
		"tags":           true,
		"default_method": true,
		"accent_color":   true,
		"pattern":        true,
		"favorite":       true,
	}

	updateData := make(map[string]interface{})
	for k, v := range fields {
		if allowedFields[k] {
			updateData[k] = v
		}
	}

	if len(updateData) == 0 {
		return nil, internal.NewBadRequest("No valid fields provided for update")
	}

	err = s.repo.UpdateFields(ctx, collectionID, updateData)
	if err != nil {
		return nil, internal.NewInternalError("Failed to update collection fields")
	}

	updatedCol, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to retrieve updated collection")
	}

	tags := []string{}
	if updatedCol.Tags != nil {
		tags = *updatedCol.Tags
	}

	return &CollectionResponse{
		ID:             updatedCol.ID.Hex(),
		MasterID:       updatedCol.MasterID.Hex(),
		Name:           updatedCol.Name,
		Tags:           tags,
		Default_Method: updatedCol.Default_Method,
		Accent_Color:   updatedCol.Accent_Color,
		Pattern:        updatedCol.Pattern,
		TotalRequests:  updatedCol.TotalRequests,
		Favorite:       updatedCol.Favorite,
		CreatedAt:      updatedCol.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      updatedCol.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *CollectionService) CreateDefaultCollections(ctx context.Context, userID string) error {
	// Acquire per-user lock to prevent race conditions when creating default collections
	lockObj, _ := s.userLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := lockObj.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	// Optionally, we could delete the lock from sync.Map, but it's fine to leave it for the lifetime of the process.

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Check if already has collections to avoid duplicates (double check for race)
	existing, err := s.repo.ListByUserID(ctx, userID)
	if err == nil && len(existing) > 0 {
		return nil
	}

	constItems, err := s.constRepo.ListLatestByTypeRaw(ctx, "collection", 5)
	if err != nil {
		return err
	}

	for _, item := range constItems {
		newCol := collections.Collection{
			UserID:        objUserID,
			TotalRequests: 0,
		}

		if teamID, ok := ctx.Value(internal.TeamIDKey).(string); ok && teamID != "" {
			if objTeamID, err := primitive.ObjectIDFromHex(teamID); err == nil {
				newCol.TeamID = &objTeamID
			}
		}

		if orgID, ok := ctx.Value(internal.OrgIDKey).(string); ok && orgID != "" {
			if objOrgID, err := primitive.ObjectIDFromHex(orgID); err == nil {
				newCol.OrgID = &objOrgID
			}
		}

		if name, ok := item["name"].(string); ok {
			newCol.Name = name
		}
		if method, ok := item["default_method"].(string); ok {
			newCol.Default_Method = method
		}
		if color, ok := item["accent_color"].(string); ok {
			newCol.Accent_Color = color
		}
		if pattern, ok := item["pattern"].(string); ok {
			newCol.Pattern = pattern
		}
		if tagsRaw, ok := item["tags"].(primitive.A); ok {
			var tags []string
			for _, t := range tagsRaw {
				if ts, ok := t.(string); ok {
					tags = append(tags, ts)
				}
			}
			newCol.Tags = &tags
		}

		s.repo.Create(ctx, &newCol)
	}
	return nil
}

func (s *CollectionService) Delete(ctx context.Context, collectionID string, userID string) error {
	col, err := s.repo.GetByID(ctx, collectionID)
	if err != nil {
		return internal.NewNotFound("Collection not found")
	}

	if !s.canModify(ctx, col, userID) {
		return internal.NewUnauthorized("Unauthorized to modify this collection")
	}

	// Delete all requests in this collection
	err = s.requestsRepo.DeleteByCollectionID(ctx, collectionID)
	if err != nil {
		log.Printf("[SERVICE] Failed to delete requests for collection %s: %v", collectionID, err)
	}

	err = s.repo.Delete(ctx, collectionID)
	if err != nil {
		return internal.NewInternalError("Failed to delete collection")
	}

	return nil
}