package organizations

import (
	"context"
	"strings"

	orgDB "pog/database/organizations"
	mappingDB "pog/database/user_mapping"
	userDB "pog/database/users"
	"pog/internal"
)

type OrganizationService struct {
	orgRepo     *orgDB.OrganizationRepository
	userRepo    *userDB.UserRepository
	mappingRepo *mappingDB.UserMappingRepository
}

func NewOrganizationService(orgRepo *orgDB.OrganizationRepository, userRepo *userDB.UserRepository, mappingRepo *mappingDB.UserMappingRepository) *OrganizationService {
	return &OrganizationService{
		orgRepo:     orgRepo,
		userRepo:    userRepo,
		mappingRepo: mappingRepo,
	}
}

// GetDetails checks if the user is affiliated with the requested organization and returns its details
func (s *OrganizationService) GetDetails(ctx context.Context, userID, orgID string) (*OrganizationResponse, error) {
	// 1. Fetch User
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to fetch user")
	}
	if user == nil {
		return nil, internal.NewUnauthorized("User not found")
	}

	// 2. Fetch Organization
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to fetch organization")
	}
	if org == nil {
		return nil, internal.NewNotFound("Organization not found")
	}

	// 3. Determine Affiliation
	isAffiliated := false
	isVerified := false

	if user.OrganizationID != nil && user.OrganizationID.Hex() == orgID {
		isAffiliated = true
		isVerified = true
	} else if user.EmailAddress != nil && org.Email != "" {
		userParts := strings.Split(*user.EmailAddress, "@")
		orgParts := strings.Split(org.Email, "@")
		if len(userParts) == 2 && len(orgParts) == 2 && strings.EqualFold(userParts[1], orgParts[1]) {
			isAffiliated = true
		}
	}

	if !isAffiliated {
		return nil, internal.NewForbidden("You are not affiliated with this organization")
	}

	memberCount, err := s.userRepo.CountByOrganizationID(ctx, orgID)
	if err != nil {
		memberCount = 1 // Default to 1 if we fail to count
	}

	return &OrganizationResponse{
		Success: true,
		Org: OrganizationDetails{
			ID:          org.ID.Hex(),
			Name:        org.Name,
			MemberCount: memberCount, 
			Role:        "Member", 
			IsVerified:  isVerified,
		},
	}, nil
}

// Verify connects the user to the organization
func (s *OrganizationService) Verify(ctx context.Context, userID, orgID string) (*VerifyResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, internal.NewUnauthorized("User not found")
	}

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil || org == nil {
		return nil, internal.NewBadRequest("Organization not found")
	}

	canVerify := false
	if user.OrganizationID != nil && user.OrganizationID.Hex() == orgID {
		canVerify = true
	} else if user.EmailAddress != nil && org.Email != "" {
		userParts := strings.Split(*user.EmailAddress, "@")
		orgParts := strings.Split(org.Email, "@")
		if len(userParts) == 2 && len(orgParts) == 2 && strings.EqualFold(userParts[1], orgParts[1]) {
			canVerify = true
		}
	}

	if !canVerify {
		return nil, internal.NewForbidden("You are not eligible to verify for this organization")
	}

	err = s.userRepo.ConnectToOrganization(ctx, userID, orgID)
	if err != nil {
		return nil, internal.NewInternalError("Failed to connect to organization")
	}

	// 4. Create UserMapping for org context
	mapping := &mappingDB.UserMapping{
		UserID:         internal.MustObjectID(userID),
		OrganizationID: internal.MustObjectID(orgID),
		Type:           "org",
		Role:           "member",
		Status:         "active",
	}
	if err := s.mappingRepo.Create(ctx, mapping); err != nil {
		// Log but don't fail the whole request (soft-fail for now)
		internal.LogInfo("Failed to create user mapping during org verification: %v", err)
	}

	return &VerifyResponse{
		Success: true,
		Message: "Successfully affiliated with the organization.", 
		Org: VerifiedOrgDetails{
			ID:         org.ID.Hex(),
			Name:       org.Name,
			IsVerified: true,
		},
	}, nil
}
