package teams

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"time"

	teamFeedDB "pog/database/team_feed"
	teamInvitesDB "pog/database/team_invites"
	teamsDB "pog/database/teams"
	teamsMappingDB "pog/database/teams_mapping"
	userMappingDB "pog/database/user_mapping"
	usersDB "pog/database/users"
	"pog/internal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TeamService struct {
	repo           *teamsDB.TeamsRepository
	mappingRepo    *teamsMappingDB.TeamsMappingRepository
	userMappingRepo *userMappingDB.UserMappingRepository
	inviteRepo     *teamInvitesDB.TeamInvitesRepository
	feedRepo       *teamFeedDB.TeamFeedRepository
	userRepo       *usersDB.UserRepository
}

func NewTeamService(
	repo *teamsDB.TeamsRepository,
	mappingRepo *teamsMappingDB.TeamsMappingRepository,
	userMappingRepo *userMappingDB.UserMappingRepository,
	inviteRepo *teamInvitesDB.TeamInvitesRepository,
	feedRepo *teamFeedDB.TeamFeedRepository,
	userRepo *usersDB.UserRepository,
) *TeamService {
	return &TeamService{
		repo:           repo,
		mappingRepo:    mappingRepo,
		userMappingRepo: userMappingRepo,
		inviteRepo:     inviteRepo,
		feedRepo:       feedRepo,
		userRepo:       userRepo,
	}
}

// ── helpers ─────────────────────────────────────────────────────────

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func baseURL() string {
	if u := os.Getenv("APP_BASE_URL"); u != "" {
		return u
	}
	return "http://localhost:3000"
}

func (s *TeamService) buildTeamResponse(team *teamsDB.Team, role string, memberCount int64) *TeamResponse {
	return &TeamResponse{
		ID:          team.ID.Hex(),
		Name:        team.Name,
		Theme:       team.Theme,
		Description: team.Description,
		Role:        role,
		MemberCount: memberCount,
		InviteLink:  baseURL() + "/join-team/" + team.InviteToken,
		CreatedAt:   team.CreatedAt.Format(time.RFC3339),
	}
}

// requireRole returns the caller's mapping or an error if not
// at least the required role.  "admin" accepts admin+owner.
func (s *TeamService) requireRole(ctx context.Context, teamID, userID string, minRole string) (*teamsMappingDB.TeamMapping, error) {
	member, err := s.mappingRepo.FindMember(ctx, teamID, userID)
	if err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	// Owner > Admin > Member
	if minRole == "admin" && member.Role == "member" {
		return nil, internal.NewForbidden("admin privileges required")
	}

	// Fetch the team to check ownership
	if minRole == "owner" {
		team, err := s.repo.FindID(ctx, teamID) // method renamed to FindID in teams repo
		if err != nil {
			return nil, internal.NewNotFound("team not found")
		}
		if team.OwnerID.Hex() != userID {
			return nil, internal.NewForbidden("only the team owner can perform this action")
		}
	}

	return member, nil
}

// ── 1. Teams CRUD ───────────────────────────────────────────────────

func (s *TeamService) CreateTeam(ctx context.Context, dto *CreateTeamDTO, userID string) (*TeamResponse, error) {
	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user id")
	}

	token, err := generateInviteToken()
	if err != nil {
		return nil, internal.NewInternalError("failed to generate invite token")
	}

	orgID, _ := ctx.Value(internal.OrgIDKey).(string)
	var objOrgID *primitive.ObjectID
	if orgID != "" {
		oid, err := primitive.ObjectIDFromHex(orgID)
		if err == nil {
			objOrgID = &oid
		}
	}

	team := &teamsDB.Team{
		Name:           dto.Name,
		Theme:          dto.Theme,
		Description:    dto.Description,
		OwnerID:        objUserID,
		OrganizationID: objOrgID,
		InviteToken:    token,
	}

	created, err := s.repo.Create(ctx, team)
	if err != nil {
		return nil, internal.NewInternalError("failed to create team")
	}

	// Auto-add creator as owner (old system) — was incorrectly "admin"
	mapping := &teamsMappingDB.TeamMapping{
		TeamID: created.ID,
		UserID: objUserID,
		Role:   "owner",
	}
	if err := s.mappingRepo.AddMember(ctx, mapping); err != nil {
		log.Printf("[SERVICE] Failed to add creator as member (old mapping): %v", err)
	}

	// NEW: Unified UserMapping for Org+Team context
	if objOrgID != nil {
		userMapping := &userMappingDB.UserMapping{
			UserID:         objUserID,
			OrganizationID: *objOrgID,
			TeamID:         &created.ID,
			Type:           "team",
			Role:           "owner",
			Status:         "active",
		}
		if err := s.userMappingRepo.Create(ctx, userMapping); err != nil {
			log.Printf("[SERVICE] Failed to create unified user mapping: %v", err)
		}
	}

	return s.buildTeamResponse(created, "owner", 1), nil
}

func (s *TeamService) ListMyTeams(ctx context.Context, userID string) ([]TeamResponse, error) {
	scope := internal.GetScope(ctx)
	orgID := scope.OrgID

	mappings, err := s.mappingRepo.ListTeamsByUserID(ctx, userID)
	if err != nil {
		return nil, internal.NewInternalError("failed to list teams")
	}

	responses := make([]TeamResponse, 0, len(mappings))
	for _, m := range mappings {
		team, err := s.repo.FindID(ctx, m.TeamID.Hex())
		if err != nil {
			log.Printf("[SERVICE] Team %s not found for mapping: %v", m.TeamID.Hex(), err)
			continue
		}

		// Filter by Organization if OrgID is present in context
		if orgID != "" {
			if team.OrganizationID == nil || team.OrganizationID.Hex() != orgID {
				continue
			}
		}

		count, _ := s.mappingRepo.CountMembers(ctx, m.TeamID.Hex())
		
		// In ListMyTeams, after finding the team, check OwnerID to assign correct role
		role := m.Role
		if team.OwnerID.Hex() == userID {
			role = "owner"
		}
		responses = append(responses, *s.buildTeamResponse(team, role, count))
	}
	return responses, nil
}

func (s *TeamService) UpdateTeam(ctx context.Context, teamID string, dto *UpdateTeamDTO, userID string) (*TeamResponse, error) {
	if _, err := s.requireRole(ctx, teamID, userID, "admin"); err != nil {
		return nil, err
	}

	updateData := bson.M{}
	if dto.Name != nil {
		if len(*dto.Name) < 3 || len(*dto.Name) > 50 {
			return nil, internal.NewBadRequest("name must be 3-50 characters")
		}
		updateData["name"] = *dto.Name
	}
	if dto.Theme != nil {
		validThemes := map[string]bool{
			"purple": true, "emerald": true, "blue": true,
			"orange": true, "rose": true, "cyan": true,
		}
		if !validThemes[*dto.Theme] {
			return nil, internal.NewBadRequest("invalid theme")
		}
		updateData["theme"] = *dto.Theme
	}
	if dto.Description != nil {
		if len(*dto.Description) > 300 {
			return nil, internal.NewBadRequest("description cannot exceed 300 characters")
		}
		updateData["description"] = *dto.Description
	}

	if len(updateData) == 0 {
		return nil, internal.NewBadRequest("no fields to update")
	}

	if err := s.repo.Update(ctx, teamID, updateData); err != nil {
		return nil, internal.NewInternalError("failed to update team")
	}

	team, _ := s.repo.FindID(ctx, teamID)
	member, _ := s.mappingRepo.FindMember(ctx, teamID, userID)
	count, _ := s.mappingRepo.CountMembers(ctx, teamID)
	return s.buildTeamResponse(team, member.Role, count), nil
}

func (s *TeamService) DeleteTeam(ctx context.Context, teamID, userID string) error {
	if _, err := s.requireRole(ctx, teamID, userID, "owner"); err != nil {
		return err
	}

	// 1. Delete meta document
	if err := s.repo.Delete(ctx, teamID); err != nil {
		return internal.NewInternalError("failed to delete team")
	}

	// 2. Cascade delete linked data
	_ = s.mappingRepo.DeleteAllByTeamID(ctx, teamID)
	_ = s.inviteRepo.DeleteAllByTeamID(ctx, teamID)
	_ = s.feedRepo.DeleteAllByTeamID(ctx, teamID)

	return nil
}

// ── 2. Members ──────────────────────────────────────────────────────

func (s *TeamService) ListMembers(ctx context.Context, teamID, userID string) ([]MemberResponse, error) {
	// Verify caller is a member
	if _, err := s.mappingRepo.FindMember(ctx, teamID, userID); err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	members, err := s.mappingRepo.ListMembersByTeamID(ctx, teamID)
	if err != nil {
		return nil, internal.NewInternalError("failed to list members")
	}

	responses := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		user, err := s.userRepo.GetByID(ctx, m.UserID.Hex())
		if err != nil {
			log.Printf("[SERVICE] User %s not found: %v", m.UserID.Hex(), err)
			continue
		}

		name := user.FirstName
		if user.LastName != nil {
			name += " " + *user.LastName
		}
		email := ""
		if user.EmailAddress != nil {
			email = *user.EmailAddress
		}

		responses = append(responses, MemberResponse{
			ID:       m.UserID.Hex(),
			Name:     name,
			Email:    email,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		})
	}
	return responses, nil
}

func (s *TeamService) ChangeRole(ctx context.Context, teamID, targetUID, userID string, dto *ChangeRoleDTO) error {
	if _, err := s.requireRole(ctx, teamID, userID, "admin"); err != nil {
		return err
	}
	if dto.Role != "admin" && dto.Role != "member" {
		return internal.NewBadRequest("role must be 'admin' or 'member'")
	}

	// Cannot change owner's role
	team, err := s.repo.FindID(ctx, teamID)
	if err != nil {
		return internal.NewNotFound("team not found")
	}
	if team.OwnerID.Hex() == targetUID {
		return internal.NewForbidden("cannot change the owner's role")
	}

	// Target must be a member
	if _, err := s.mappingRepo.FindMember(ctx, teamID, targetUID); err != nil {
		return internal.NewNotFound("user is not a member of this team")
	}

	return s.mappingRepo.UpdateMemberRole(ctx, teamID, targetUID, dto.Role)
}

func (s *TeamService) RemoveMember(ctx context.Context, teamID, targetUID, userID string) error {
	if _, err := s.requireRole(ctx, teamID, userID, "admin"); err != nil {
		return err
	}

	team, err := s.repo.FindID(ctx, teamID)
	if err != nil {
		return internal.NewNotFound("team not found")
	}
	if team.OwnerID.Hex() == targetUID {
		return internal.NewForbidden("cannot remove the team owner")
	}

	return s.mappingRepo.RemoveMember(ctx, teamID, targetUID)
}

func (s *TeamService) TransferOwnership(ctx context.Context, teamID, newOwnerID, currentUserID string) error {
	// 1. Verify caller is owner
	if _, err := s.requireRole(ctx, teamID, currentUserID, "owner"); err != nil {
		return err
	}

	// 2. Verify target is a member
	_, err := s.mappingRepo.FindMember(ctx, teamID, newOwnerID)
	if err != nil {
		return internal.NewBadRequest("target user is not a member of this team")
	}

	// 3. Update Team document
	if err := s.repo.Update(ctx, teamID, bson.M{"owner_id": internal.MustObjectID(newOwnerID)}); err != nil {
		return internal.NewInternalError("failed to update team owner")
	}

	// 4. Update roles in old mapping system
	if err := s.mappingRepo.UpdateMemberRole(ctx, teamID, currentUserID, "admin"); err != nil {
		log.Printf("[SERVICE] Failed to demote old owner: %v", err)
	}
	if err := s.mappingRepo.UpdateMemberRole(ctx, teamID, newOwnerID, "admin"); err != nil {
		log.Printf("[SERVICE] Failed to update new owner role: %v", err)
	}

	// 5. Update roles in NEW mapping system
	// Demote old owner
	oldMapping, err := s.userMappingRepo.FindByUserOrgAndTeam(ctx, currentUserID, "", teamID) // OrgID empty as check is team-scoped
	if err == nil && oldMapping != nil {
		s.userMappingRepo.UpdateRole(ctx, oldMapping.ID.Hex(), "admin", nil)
	}
	// Promote new owner
	newMapping, err := s.userMappingRepo.FindByUserOrgAndTeam(ctx, newOwnerID, "", teamID)
	if err == nil && newMapping != nil {
		s.userMappingRepo.UpdateRole(ctx, newMapping.ID.Hex(), "owner", nil)
	}

	return nil
}

func (s *TeamService) BulkRemoveMembers(ctx context.Context, teamID string, userIDs []string, currentUserID string) error {
	if _, err := s.requireRole(ctx, teamID, currentUserID, "admin"); err != nil {
		return err
	}

	team, err := s.repo.FindID(ctx, teamID)
	if err != nil {
		return internal.NewNotFound("team not found")
	}

	for _, uid := range userIDs {
		if uid == team.OwnerID.Hex() {
			continue // Skip owner
		}
		// Remove from old system
		s.mappingRepo.RemoveMember(ctx, teamID, uid)
		
		// Remove from new system
		m, err := s.userMappingRepo.FindByUserOrgAndTeam(ctx, uid, "", teamID)
		if err == nil && m != nil {
			s.userMappingRepo.Delete(ctx, m.ID.Hex())
		}
	}

	return nil
}

// ── 3. Invites ──────────────────────────────────────────────────────

func (s *TeamService) InviteByEmail(ctx context.Context, teamID, userID string, dto *InviteDTO) (*InviteResponse, error) {
	if _, err := s.requireRole(ctx, teamID, userID, "admin"); err != nil {
		return nil, err
	}

	objTeamID, _ := primitive.ObjectIDFromHex(teamID)

	invited := 0
	alreadyMember := 0

	for _, email := range dto.Emails {
		// Check if user exists by email
		user, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil || user == nil {
			return nil, internal.NewBadRequest("User with email " + email + " not found. They must sign up first.")
		}

		// User exists — check if already a member
		isMember, _ := s.mappingRepo.IsMember(ctx, teamID, user.ID.Hex())
		if isMember {
			alreadyMember++
			continue
		}

		// Auto-add existing user
		mapping := &teamsMappingDB.TeamMapping{
			TeamID: objTeamID,
			UserID: user.ID,
			Role:   "member",
		}
		if err := s.mappingRepo.AddMember(ctx, mapping); err == nil {
			invited++
		}
	}

	return &InviteResponse{Invited: invited, AlreadyMember: alreadyMember}, nil
}

func (s *TeamService) GetInviteLink(ctx context.Context, teamID, userID string) (*InviteLinkResponse, error) {
	if _, err := s.mappingRepo.FindMember(ctx, teamID, userID); err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	team, err := s.repo.FindID(ctx, teamID)
	if err != nil {
		return nil, internal.NewNotFound("team not found")
	}

	return &InviteLinkResponse{
		Link:  baseURL() + "/join-team/" + team.InviteToken,
		Token: team.InviteToken,
	}, nil
}

func (s *TeamService) JoinViaToken(ctx context.Context, token, userID string) (*TeamResponse, error) {
	team, err := s.repo.FindByInviteToken(ctx, token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, internal.NewNotFound("invalid or expired invite link")
		}
		return nil, internal.NewInternalError("failed to look up invite link")
	}

	isMember, _ := s.mappingRepo.IsMember(ctx, team.ID.Hex(), userID)
	if isMember {
		return nil, internal.NewConflict("you are already a member of this team")
	}

	objUserID, _ := primitive.ObjectIDFromHex(userID)
	mapping := &teamsMappingDB.TeamMapping{
		TeamID: team.ID,
		UserID: objUserID,
		Role:   "member",
	}
	if err := s.mappingRepo.AddMember(ctx, mapping); err != nil {
		return nil, internal.NewInternalError("failed to join team (old mapping)")
	}

	// NEW: Unified UserMapping
	if team.OrganizationID != nil {
		userMapping := &userMappingDB.UserMapping{
			UserID:         objUserID,
			OrganizationID: *team.OrganizationID,
			TeamID:         &team.ID,
			Type:           "team",
			Role:           "member",
			Status:         "active",
		}
		if err := s.userMappingRepo.Create(ctx, userMapping); err != nil {
			log.Printf("[SERVICE] Failed to create unified user mapping on join: %v", err)
		}
	}

	count, _ := s.mappingRepo.CountMembers(ctx, team.ID.Hex())
	return s.buildTeamResponse(team, "member", count), nil
}

// ── 4. Feed ─────────────────────────────────────────────────────────

func (s *TeamService) GetFeed(ctx context.Context, teamID, userID string, page, limit int) (*PaginatedFeedResponse, error) {
	if _, err := s.mappingRepo.FindMember(ctx, teamID, userID); err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	items, total, err := s.feedRepo.ListByTeamID(ctx, teamID, page, limit)
	if err != nil {
		return nil, internal.NewInternalError("failed to get feed")
	}

	responses := make([]FeedItemResponse, 0, len(items))
	for _, item := range items {
		userName := "Unknown"
		user, err := s.userRepo.GetByID(ctx, item.UserID.Hex())
		if err == nil {
			userName = user.FirstName
			if user.LastName != nil {
				userName += " " + *user.LastName
			}
		}

		fr := FeedItemResponse{
			ID:        item.ID.Hex(),
			Type:      item.Type,
			UserID:    item.UserID.Hex(),
			UserName:  userName,
			Message:   item.Message,
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
		}

		// Mentions
		if len(item.Mentions) > 0 {
			mentions := make([]string, len(item.Mentions))
			for i, m := range item.Mentions {
				mentions[i] = m.Hex()
			}
			fr.Mentions = mentions
		}

		// Issue-specific fields
		if item.Type == "issue" {
			fr.Title = item.Title
			fr.Resolved = &item.Resolved
			if item.ResolvedBy != nil {
				rb := item.ResolvedBy.Hex()
				fr.ResolvedBy = &rb
			}
		}

		responses = append(responses, fr)
	}

	return &PaginatedFeedResponse{
		Data: responses,
		Pagination: PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}

func (s *TeamService) SendFeed(ctx context.Context, teamID, userID string, dto *SendFeedDTO) (*FeedItemResponse, error) {
	if _, err := s.mappingRepo.FindMember(ctx, teamID, userID); err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	objTeamID, _ := primitive.ObjectIDFromHex(teamID)
	objUserID, _ := primitive.ObjectIDFromHex(userID)

	// Parse mentions
	var mentions []primitive.ObjectID
	for _, mid := range dto.Mentions {
		if oid, err := primitive.ObjectIDFromHex(mid); err == nil {
			mentions = append(mentions, oid)
		}
	}

	feedItem := &teamFeedDB.TeamFeed{
		TeamID:   objTeamID,
		UserID:   objUserID,
		Type:     dto.Type,
		Message:  dto.Message,
		Mentions: mentions,
	}
	if dto.Title != nil {
		feedItem.Title = *dto.Title
	}

	created, err := s.feedRepo.Create(ctx, feedItem)
	if err != nil {
		return nil, internal.NewInternalError("failed to send feed item")
	}

	// Get sender name
	userName := "You"
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		userName = user.FirstName
		if user.LastName != nil {
			userName += " " + *user.LastName
		}
	}

	resp := &FeedItemResponse{
		ID:        created.ID.Hex(),
		Type:      created.Type,
		UserID:    userID,
		UserName:  userName,
		Message:   created.Message,
		CreatedAt: created.CreatedAt.Format(time.RFC3339),
	}
	if created.Type == "issue" {
		resp.Title = created.Title
		resolved := false
		resp.Resolved = &resolved
	}
	if len(dto.Mentions) > 0 {
		resp.Mentions = dto.Mentions
	}

	return resp, nil
}

func (s *TeamService) ResolveFeed(ctx context.Context, teamID, feedID, userID string) (*FeedItemResponse, error) {
	if _, err := s.mappingRepo.FindMember(ctx, teamID, userID); err != nil {
		return nil, internal.NewForbidden("you are not a member of this team")
	}

	item, err := s.feedRepo.FindByID(ctx, feedID)
	if err != nil {
		return nil, internal.NewNotFound("feed item not found")
	}
	if item.Type != "issue" {
		return nil, internal.NewBadRequest("only issues can be resolved")
	}
	if item.Resolved {
		return nil, internal.NewBadRequest("issue is already resolved")
	}

	if err := s.feedRepo.Resolve(ctx, feedID, userID); err != nil {
		return nil, internal.NewInternalError("failed to resolve issue")
	}

	resolved := true
	return &FeedItemResponse{
		ID:         item.ID.Hex(),
		Resolved:   &resolved,
		ResolvedBy: &userID,
	}, nil
}
