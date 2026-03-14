package user_mapping

import (
	"context"
	mappingDB "pog/database/user_mapping"
	"pog/internal"
)

type UserMappingService struct {
	repo *mappingDB.UserMappingRepository
}

func NewUserMappingService(repo *mappingDB.UserMappingRepository) *UserMappingService {
	return &UserMappingService{
		repo: repo,
	}
}

// MapUserToOrg links a user to an organization with a specific role
func (s *UserMappingService) MapUserToOrg(ctx context.Context, userID, orgID, role string) error {
	mapping := &mappingDB.UserMapping{
		UserID:         internal.MustObjectID(userID),
		OrganizationID: internal.MustObjectID(orgID),
		Type:           "org",
		Role:           role,
		Status:         "active",
	}
	return s.repo.Create(ctx, mapping)
}

// MapUserToTeam links a user to a specific team within an organization
func (s *UserMappingService) MapUserToTeam(ctx context.Context, userID, orgID, teamID, role string) error {
	mapping := &mappingDB.UserMapping{
		UserID:         internal.MustObjectID(userID),
		OrganizationID: internal.MustObjectID(orgID),
		TeamID:         internal.PtrObjectID(teamID),
		Type:           "team",
		Role:           role,
		Status:         "active",
	}
	return s.repo.Create(ctx, mapping)
}

// GetUserContext retrieves the last active or default context for a user
func (s *UserMappingService) GetUserContext(ctx context.Context, userID string) ([]mappingDB.UserMapping, error) {
	// This would eventually be used to determine what the user sees on login
	return nil, nil // TODO: Implement logic to list all orgs/teams for a user
}

// ValidatePermission checks if a user has a specific permission in a context
func (s *UserMappingService) ValidatePermission(ctx context.Context, userID, orgID, teamID, permission string) (bool, error) {
	// Check org-level first (admins might have full access)
	// Then check team-level
	return false, nil // TODO: Implement hierarchy check logic
}
