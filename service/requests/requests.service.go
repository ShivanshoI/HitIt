package requests

import (
	"context"
	"pog/database/collections"
	"pog/database/requests"
	"pog/internal"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestService struct {
	repo           *requests.RequestRepository
	collectionRepo *collections.CollectionRepository
}

func NewRequestService(repo *requests.RequestRepository, collectionRepo *collections.CollectionRepository) *RequestService {
	return &RequestService{
		repo:           repo,
		collectionRepo: collectionRepo,
	}
}

func (s *RequestService) Create(ctx context.Context, payload *CreateRequestDTO, userID string) (*RequestResponse, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user id")
	}

	collectionId, err := primitive.ObjectIDFromHex(payload.CollectionID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid collection id")
	}
	//assinging default method in case of no method is passed
	if payload.Method == "" {
		collectionDetails, err := s.collectionRepo.GetByID(ctx, payload.CollectionID)
		if err != nil {
			return nil, internal.NewBadRequest("invalid collection id")
		}
		payload.Method = collectionDetails.Default_Method
	}

	dbHeaders := make([]requests.KeyValuePair, len(payload.Headers))
	for i, h := range payload.Headers {
		dbHeaders[i] = requests.KeyValuePair{Key: h.Key, Value: h.Value}
	}

	dbParams := make([]requests.KeyValuePair, len(payload.Params))
	for i, p := range payload.Params {
		dbParams[i] = requests.KeyValuePair{Key: p.Key, Value: p.Value}
	}

	reqModel := &requests.APIRequest{
		UserID:       userId,
		CollectionID: collectionId,
		Name:         payload.Name,
		Method:       payload.Method,
		URL:          payload.URL,
		Headers:      dbHeaders,
		Params:       dbParams,
		Body:         payload.Body,
		Auth:         payload.Auth,
	}

	req, err := s.repo.Create(ctx, reqModel)
	if err != nil {
		return nil, internal.NewInternalError("Failed to create request")
	}

	respHeaders := make([]KeyValuePair, len(req.Headers))
	for i, h := range req.Headers {
		respHeaders[i] = KeyValuePair{Key: h.Key, Value: h.Value}
	}

	respParams := make([]KeyValuePair, len(req.Params))
	for i, p := range req.Params {
		respParams[i] = KeyValuePair{Key: p.Key, Value: p.Value}
	}

	// Incrementing the total requests in the background
	go func(colID string) {
		bgCtx := context.Background()

		collectionDetails, err := s.collectionRepo.GetByID(bgCtx, colID)
		if err != nil {
			// use a logger here (e.g., log.Printf) instead of returning
			return
		}

		collectionDetails.TotalRequests++
		_, err = s.collectionRepo.Update(bgCtx, colID, collectionDetails)
		if err != nil {
			// use a logger here
			return
		}
	}(payload.CollectionID)

	return &RequestResponse{
		ID:           req.ID.Hex(),
		CollectionID: req.CollectionID.Hex(),
		Name:         req.Name,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      respHeaders,
		Params:       respParams,
		Body:         req.Body,
		Auth:         req.Auth,
		CreatedAt:    req.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    req.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *RequestService) ListByCollection(ctx context.Context, collectionID string) ([]RequestSummaryResponse, error) {
	requestsList, err := s.repo.ListSummariesByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to list requests")
	}

	responses := make([]RequestSummaryResponse, 0, len(requestsList))
	for _, req := range requestsList {
		responses = append(responses, RequestSummaryResponse{
			ID:           req.ID.Hex(),
			CollectionID: req.CollectionID.Hex(),
			Name:         req.Name,
			Method:       req.Method,
		})
	}

	return responses, nil
}

func (s *RequestService) GetByID(ctx context.Context, requestID string) (*RequestResponse, error) {
	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, internal.NewNotFound("Request not found")
	}

	respHeaders := make([]KeyValuePair, len(req.Headers))
	for i, h := range req.Headers {
		respHeaders[i] = KeyValuePair{Key: h.Key, Value: h.Value}
	}

	respParams := make([]KeyValuePair, len(req.Params))
	for i, p := range req.Params {
		respParams[i] = KeyValuePair{Key: p.Key, Value: p.Value}
	}

	return &RequestResponse{
		ID:           req.ID.Hex(),
		CollectionID: req.CollectionID.Hex(),
		Name:         req.Name,
		Method:       req.Method,
		URL:          req.URL,
		Headers:      respHeaders,
		Params:       respParams,
		Body:         req.Body,
		Auth:         req.Auth,
		CreatedAt:    req.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    req.UpdatedAt.Format(time.RFC3339),
	}, nil
}
