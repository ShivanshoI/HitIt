package collaborators

import (
	"context"
	"net/url"
	"pog/database/collaborators"
	"pog/database/collections"
	"pog/database/requests"
	"pog/database/users"
	"pog/internal"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CollaboratorService struct {
	repo           *collaborators.CollaboratorRepository
	collectionRepo *collections.CollectionRepository
	requestRepo    *requests.RequestRepository
	userRepo       *users.UserRepository
}

func NewCollaboratorService(
	repo *collaborators.CollaboratorRepository,
	collectionRepo *collections.CollectionRepository,
	requestRepo *requests.RequestRepository,
	userRepo *users.UserRepository,
) *CollaboratorService {
	return &CollaboratorService{
		repo:           repo,
		collectionRepo: collectionRepo,
		requestRepo:    requestRepo,
		userRepo:       userRepo,
	}
}

func (s *CollaboratorService) ImportDistributer(ctx context.Context, userID string, shareLink string) (string, error) {
	linkPayload := linkParser(shareLink)

	switch linkPayload.EntityType {
	case "c":
		return s.importCollection(ctx, userID, linkPayload)
	case "r":
		return s.importRequest(ctx, userID, linkPayload)
	default:
		return "failed", internal.NewBadRequest("only collections and requests can be imported currently")
	}
}

func (s *CollaboratorService) importCollection(ctx context.Context, userID string, linkPayload LinkPayload) (string, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "failed", internal.NewBadRequest("invalid user id")
	}

	var originalCol *collections.Collection
	if linkPayload.IsNew {
		originalCol, err = s.collectionRepo.GetByID(ctx, linkPayload.IDString)
		if err != nil {
			return "failed", internal.NewNotFound("original collection not found")
		}
	} else {
		originalCol, err = s.collectionRepo.GetByMasterID(ctx, linkPayload.IDString)
		if err != nil {
			return "failed", internal.NewNotFound("original collection not found")
		}
	}

	var tags []string
	if originalCol.Tags != nil {
		tags = append(tags, *originalCol.Tags...)
	}

	isShared := false
	for _, t := range tags {
		if t == "shared" {
			isShared = true
			break
		}
	}
	if !isShared {
		tags = append(tags, "shared")
	}

	newCol := &collections.Collection{
		UserID:          objUserID,
		MasterID:        originalCol.MasterID,
		Name:            originalCol.Name,
		Tags:            &tags,
		Default_Method:  originalCol.Default_Method,
		Accent_Color:    originalCol.Accent_Color,
		Pattern:         originalCol.Pattern,
		WritePermission: linkPayload.Permission,
	}

	_, err = s.collectionRepo.Create(ctx, newCol)
	if err != nil {
		return "failed", err
	}

	// 4. Clone requests in background
	go func(colID string, newColID primitive.ObjectID, masterID primitive.ObjectID, userID primitive.ObjectID, writePermission bool, isNew bool) {
		bgCtx := context.Background()

		originalRequests, err := s.requestRepo.ListByCollectionID(bgCtx, colID)
		if err != nil || len(originalRequests) == 0 {
			// use a logger here (e.g., log.Printf) instead of returning
			return
		}

		newRequests := make([]interface{}, len(originalRequests))
		for i, req := range originalRequests {
			clonedReq := &requests.APIRequest{
				UserID:          userID,
				MasterID:        req.MasterID,
				CollectionID:    newColID,
				Name:            req.Name,
				Tags:            req.Tags,
				Method:          req.Method,
				URL:             req.URL,
				Headers:         req.Headers,
				Params:          req.Params,
				Body:            req.Body,
				Auth:            req.Auth,
				Note:            req.Note,
				WritePermission: writePermission,
			}

			if isNew {
				id := primitive.NewObjectID()
				clonedReq.ID = id
				clonedReq.MasterID = id
			}

			newRequests[i] = clonedReq
		}

		err = s.requestRepo.BulkCreate(bgCtx, newRequests)
		if err != nil {
			// use a logger here
			return
		}
	}(linkPayload.IDString, newCol.ID, newCol.MasterID, objUserID, newCol.WritePermission, linkPayload.IsNew)

	return "success", nil
}

func (s *CollaboratorService) importRequest(ctx context.Context, userID string, linkPayload LinkPayload) (string, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "failed", internal.NewBadRequest("invalid user id")
	}

	originalReq, err := s.requestRepo.GetByID(ctx, linkPayload.IDString)
	if err != nil {
		return "failed", internal.NewNotFound("original request not found")
	}

	clonedReq := &requests.APIRequest{
		UserID:          objUserID,
		MasterID:        originalReq.MasterID,
		Name:            originalReq.Name,
		Method:          originalReq.Method,
		URL:             originalReq.URL,
		Headers:         originalReq.Headers,
		Params:          originalReq.Params,
		Body:            originalReq.Body,
		Auth:            originalReq.Auth,
		Note:            originalReq.Note,
		WritePermission: linkPayload.Permission,
	}

	if linkPayload.IsNew {
		id := primitive.NewObjectID()
		clonedReq.ID = id
		clonedReq.MasterID = id
	}

	_, err = s.requestRepo.Create(ctx, clonedReq)
	if err != nil {
		return "failed", err
	}

	return "success", nil
}

func linkParser(link string) (linkPayload LinkPayload) {
	u, err := url.Parse(link)
	if err != nil {
		return
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) <= 1 {
		return
	}

	code := parts[len(parts)-1]
	codeParts := strings.Split(code, "-")

	if len(codeParts) >= 3 {
		linkPayload.EntityType = codeParts[0]         // 'c' or 'r'
		linkPayload.Permission = codeParts[1] == "rw" // 'ro' or 'rw'
		linkPayload.IDString = codeParts[2]
	}

	linkPayload.IsNew = u.Query().Get("new") == "true"

	return linkPayload
}

func (s *CollaboratorService) GetCollaboratorsForCollection(ctx context.Context, collectionID string) ([]CollaboratorResponse, error) {
	col, err := s.collectionRepo.GetByID(ctx, collectionID)
	if err != nil {
		return nil, internal.NewNotFound("collection not found")
	}

	masterID := col.MasterID.Hex()

	collections, err := s.collectionRepo.FindAllByMasterID(ctx, masterID)
	if err != nil {
		return nil, internal.NewInternalError("failed to find collections by master id")
	}

	if len(collections) == 0 {
		return []CollaboratorResponse{}, nil
	}

	writePermissionMap := make(map[primitive.ObjectID]bool)
	var userIDs []primitive.ObjectID

	for _, col := range collections {
		userIDs = append(userIDs, col.UserID)
		writePermissionMap[col.UserID] = col.WritePermission
	}

	usersList, err := s.userRepo.FindMultipleByIDs(ctx, userIDs)
	if err != nil {
		return nil, internal.NewInternalError("failed to find users")
	}

	var responses []CollaboratorResponse
	for _, u := range usersList {
		email := ""
		if u.EmailAddress != nil {
			email = *u.EmailAddress
		}

		name := u.FirstName
		if u.LastName != nil && *u.LastName != "" {
			name = name + " " + *u.LastName
		}

		responses = append(responses, CollaboratorResponse{
			UserID:          u.ID.Hex(),
			Name:            name,
			EmailAddress:    email,
			WritePermission: writePermissionMap[u.ID],
		})
	}

	return responses, nil
}
