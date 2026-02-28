package collections

import (
	"context"
	"log"
	"pog/database/collections"
	"pog/database/constants"
	"pog/internal"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollectionService struct {
	repo      *collections.CollectionRepository
	constRepo *constants.ConstantRepository
}

func NewCollectionService(repo *collections.CollectionRepository, constRepo *constants.ConstantRepository) *CollectionService {
	return &CollectionService{
		repo:      repo,
		constRepo: constRepo,
	}
}

func (s *CollectionService) Create(ctx context.Context, payload *CreateCollectionDTO) (*CollectionResponse, error) {
	userId, err := primitive.ObjectIDFromHex(payload.UserID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user id")
	}

	collectionModel := &collections.Collection{
		UserID:         userId,
		Name:           payload.Name,
		Tags:           &payload.Tags,
		Default_Method: payload.Default_Method,
		Accent_Color:   payload.Accent_Color,
		Pattern:        payload.Pattern,
		TotalRequests:  0,
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
		Name:           collection.Name,
		UserID:         collection.UserID.Hex(),
		Tags:           tags,
		Default_Method: collection.Default_Method,
		Accent_Color:   collection.Accent_Color,
		Pattern:        collection.Pattern,
		TotalRequests:  collection.TotalRequests,
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
			Name:           col.Name,
			UserID:         col.UserID.Hex(),
			Tags:           tags,
			Default_Method: col.Default_Method,
			Accent_Color:   col.Accent_Color,
			Pattern:        col.Pattern,
			CreatedAt:      col.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      col.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

func (s *CollectionService) ListAllCollection(ctx context.Context, userID string, page, limit int) (*PaginatedCollectionResponse, error) {

	collectionsList, total, err := s.repo.ListPaginatedByUserID(ctx, userID, page, limit)
	if err != nil {
		log.Printf("[SERVICE] ListPaginatedByUserID error: %v (userID: %s)", err, userID)
		return nil, internal.NewInternalError("Failed to list collections")
	}

	if len(collectionsList) == 0 {
		objUserID, err := primitive.ObjectIDFromHex(userID)
		if err == nil {
			constItems, err := s.constRepo.ListLatestByTypeRaw(ctx, "collection", 5)
			if err == nil {
				for _, item := range constItems {
					newCol := collections.Collection{
						UserID:        objUserID,
						TotalRequests: 0,
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
					// Handle tags slice carefully
					if tagsRaw, ok := item["tags"].(primitive.A); ok {
						var tags []string
						for _, t := range tagsRaw {
							if ts, ok := t.(string); ok {
								tags = append(tags, ts)
							}
						}
						newCol.Tags = &tags
					}

					createdCol, err := s.repo.Create(ctx, &newCol)
					if err == nil {
						collectionsList = append(collectionsList, *createdCol)
						total++
					}
				}
			}
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
			Name:           col.Name,
			UserID:         col.UserID.Hex(),
			Tags:           tags,
			Default_Method: col.Default_Method,
			Accent_Color:   col.Accent_Color,
			Pattern:        col.Pattern,
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
