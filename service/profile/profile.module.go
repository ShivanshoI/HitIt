package profile

import (
	"net/http"

	pogCollectionsDB "pog/database/collections"
	pogHistoryDB "pog/database/history"
	pogTeamsMappingDB "pog/database/teams_mapping"
	pogUserActivityDB "pog/database/user_activity"
	pogUsersDB "pog/database/users"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule bootstraps the profile domain — repositories, service, and handler.
func InitModule(db *mongo.Database, mux *http.ServeMux) {
	usersRepo := pogUsersDB.NewUserRepository(db)
	collectionsRepo := pogCollectionsDB.NewCollectionRepository(db)
	historyRepo := pogHistoryDB.NewHistoryRepository(db)
	teamsMappingRepo := pogTeamsMappingDB.NewTeamsMappingRepository(db)
	activityRepo := pogUserActivityDB.NewActivityRepository(db)

	service := NewProfileService(usersRepo, collectionsRepo, historyRepo, teamsMappingRepo, activityRepo)
	handler := NewProfileHandler(service)

	handler.RegisterRoutes(mux)
}
