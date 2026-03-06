package teams

import (
	"net/http"

	teamFeedDB "pog/database/team_feed"
	teamInvitesDB "pog/database/team_invites"
	teamsDB "pog/database/teams"
	teamsMappingDB "pog/database/teams_mapping"
	usersDB "pog/database/users"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule bootstraps the teams domain and returns the
// AuthWithTeam middleware for use by other modules.
func InitModule(db *mongo.Database, mux *http.ServeMux) func(http.Handler) http.Handler {
	// Repositories
	repo := teamsDB.NewTeamsRepository(db)
	mappingRepo := teamsMappingDB.NewTeamsMappingRepository(db)
	inviteRepo := teamInvitesDB.NewTeamInvitesRepository(db)
	feedRepo := teamFeedDB.NewTeamFeedRepository(db)
	userRepo := usersDB.NewUserRepository(db)

	// Service + Handler
	service := NewTeamService(repo, mappingRepo, inviteRepo, feedRepo, userRepo)
	handler := NewTeamHandler(service)

	// Register team-specific endpoints
	handler.RegisterRoutes(mux)

	// Return the team-scoping middleware so other modules can use it
	return AuthWithTeam(db)
}
