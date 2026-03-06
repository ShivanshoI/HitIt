package profile

import (
	"context"
	"strings"

	pogCollectionsDB "pog/database/collections"
	pogHistoryDB "pog/database/history"
	pogTeamsMappingDB "pog/database/teams_mapping"
	pogUserActivityDB "pog/database/user_activity"
	pogUsersDB "pog/database/users"
	"pog/internal"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProfileService orchestrates business logic for the profile APIs.
// It depends on repos from several other packages.
type ProfileService struct {
	usersRepo        *pogUsersDB.UserRepository
	collectionsRepo  *pogCollectionsDB.CollectionRepository
	historyRepo      *pogHistoryDB.HistoryRepository
	teamsMappingRepo *pogTeamsMappingDB.TeamsMappingRepository
	activityRepo     *pogUserActivityDB.ActivityRepository
}

func NewProfileService(
	usersRepo *pogUsersDB.UserRepository,
	collectionsRepo *pogCollectionsDB.CollectionRepository,
	historyRepo *pogHistoryDB.HistoryRepository,
	teamsMappingRepo *pogTeamsMappingDB.TeamsMappingRepository,
	activityRepo *pogUserActivityDB.ActivityRepository,
) *ProfileService {
	return &ProfileService{
		usersRepo:        usersRepo,
		collectionsRepo:  collectionsRepo,
		historyRepo:      historyRepo,
		teamsMappingRepo: teamsMappingRepo,
		activityRepo:     activityRepo,
	}
}

// GetStats returns totals for collections, requests sent, and teams joined.
func (s *ProfileService) GetStats(ctx context.Context, userID string) (*ProfileStatsResponse, error) {
	collectionsCount, err := s.collectionsRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to count collections")
	}

	requestsCount, err := s.historyRepo.CountAllByUserID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to count requests")
	}

	teamMappings, err := s.teamsMappingRepo.ListTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to count teams")
	}

	return &ProfileStatsResponse{
		TotalCollections: int(collectionsCount),
		RequestsSent:     int(requestsCount),
		TeamsJoined:      len(teamMappings),
	}, nil
}

// GetActivity returns the most recent activity events for a user.
func (s *ProfileService) GetActivity(ctx context.Context, userID string, limit int) (*ActivityFeedResponse, error) {
	events, err := s.activityRepo.ListByUserID(ctx, userID, limit)
	if err != nil {
		return nil, internal.NewInternalError("failed to fetch activity")
	}

	items := make([]ActivityEventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, ActivityEventResponse{
			ID:    e.ID.Hex(),
			Type:  e.Type,
			Title: e.Title,
			Time:  e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Icon:  e.Icon,
		})
	}
	return &ActivityFeedResponse{Activity: items}, nil
}

// UpdateProfile updates the user's display name, email, and theme.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
	// Basic validation
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" || req.Email == "" {
		return nil, internal.NewBadRequest("name and email are required")
	}

	// Check email uniqueness
	clash, err := s.usersRepo.ExistsByEmailExcludingID(ctx, req.Email, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to validate email")
	}
	if clash {
		return nil, internal.NewConflict("email already in use")
	}

	// Split name into first + last (best-effort)
	parts := strings.SplitN(req.Name, " ", 2)
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
		lastName = parts[1]
	}

	// Validate theme
	validThemes := map[string]bool{"system": true, "dark": true, "light": true, "": true}
	if !validThemes[req.Theme] {
		return nil, internal.NewBadRequest("theme must be one of: system, dark, light")
	}

	updated, err := s.usersRepo.UpdateProfile(ctx, userID, firstName, lastName, req.Email, req.Theme)
	if err != nil {
		return nil, internal.NewInternalError("failed to update profile")
	}

	// Construct full display name
	displayName := updated.FirstName
	if updated.LastName != nil && *updated.LastName != "" {
		displayName += " " + *updated.LastName
	}

	email := ""
	if updated.EmailAddress != nil {
		email = *updated.EmailAddress
	}

	return &UpdateProfileResponse{
		Success: true,
		Message: "Profile updated successfully",
		User: UserProfileResponse{
			ID:    updated.ID.Hex(),
			Name:  displayName,
			Email: email,
			Theme: updated.Theme,
		},
	}, nil
}

// UpdatePassword validates the current password and sets a new one.
// NOTE: Passwords are stored as plain text in the current implementation
// (matching the existing sign-in flow). TODO: migrate to bcrypt hashing.
func (s *ProfileService) UpdatePassword(ctx context.Context, userID string, req UpdatePasswordRequest) (*UpdatePasswordResponse, error) {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return nil, internal.NewBadRequest("currentPassword and newPassword are required")
	}
	if len(req.NewPassword) < 8 {
		return nil, internal.NewBadRequest("newPassword must be at least 8 characters")
	}

	// Fetch user and verify current password
	user, err := s.usersRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, internal.NewInternalError("failed to fetch user")
	}

	// Plain-text comparison (mirrors existing sign-in logic)
	if user.Password != req.CurrentPassword {
		return nil, internal.NewUnauthorized("The current password you entered is incorrect.")
	}

	if err := s.usersRepo.UpdatePassword(ctx, userID, req.NewPassword); err != nil {
		return nil, internal.NewInternalError("failed to update password")
	}

	// Record activity event
	_ = s.LogActivity(ctx, userID, "password_changed", "Password was changed", "settings")

	return &UpdatePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// LogActivity creates a new activity event for a user.
// Errors are intentionally swallowed — activity logging is best-effort and
// must never block the main call path.
func (s *ProfileService) LogActivity(ctx context.Context, userID, eventType, title, icon string) error {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	_, err = s.activityRepo.Create(ctx, &pogUserActivityDB.ActivityEvent{
		UserID: objUserID,
		Type:   eventType,
		Title:  title,
		Icon:   icon,
	})
	return err
}
