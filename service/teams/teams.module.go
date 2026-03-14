package teams

import (
	"context"
	"log"
	"net/http"

	teamFeedDB "pog/database/team_feed"
	teamInvitesDB "pog/database/team_invites"
	teamsDB "pog/database/teams"
	teamsMappingDB "pog/database/teams_mapping"
	userMappingDB "pog/database/user_mapping"
	usersDB "pog/database/users"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule bootstraps the teams domain and returns the
// AuthWithTeam middleware for use by other modules.
func InitModule(db *mongo.Database, mux *http.ServeMux, authOrg func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	// Repositories
	repo := teamsDB.NewTeamsRepository(db)
	mappingRepo := teamsMappingDB.NewTeamsMappingRepository(db)
	userMappingRepo := userMappingDB.NewUserMappingRepository(db)
	inviteRepo := teamInvitesDB.NewTeamInvitesRepository(db)
	feedRepo := teamFeedDB.NewTeamFeedRepository(db)
	userRepo := usersDB.NewUserRepository(db)

	// Ensure database indexes are created
	ctx := context.Background()
	if err := repo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create team indexes: %v", err)
	}
	if err := mappingRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create team mapping indexes: %v", err)
	}
	if err := inviteRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create team invite indexes: %v", err)
	}
	if err := feedRepo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create team feed indexes: %v", err)
	}

	// Service + Handler
	service := NewTeamService(repo, mappingRepo, userMappingRepo, inviteRepo, feedRepo, userRepo)
	handler := NewTeamHandler(service)

		// Register team-specific endpoints
	handler.RegisterRoutes(mux, authOrg)

	// Return the team-scoping middleware so other modules can use it
	return AuthWithTeam(db)
}
