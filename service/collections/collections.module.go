package collections

import (
	"context"
	"log"
	"net/http"

	pogCollectionsDB "pog/database/collections"
	pogConstantsDB "pog/database/constants"
	pogRequestsDB "pog/database/requests"
	"pog/database/teams_mapping"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule acts as a bootstrap for the 'collections' domain.
func InitModule(db *mongo.Database, mux *http.ServeMux, authTeam func(http.Handler) http.Handler) {
	// Initialize Dependencies
	repo := pogCollectionsDB.NewCollectionRepository(db)
	constRepo := pogConstantsDB.NewConstantRepository(db)
	requestsRepo := pogRequestsDB.NewRequestRepository(db)
	mappingRepo := teams_mapping.NewTeamsMappingRepository(db)
	service := NewCollectionService(repo, constRepo, requestsRepo, mappingRepo)
	handler := NewCollectionHandler(service)

	// Ensure database indexes are created
	if err := repo.EnsureIndexes(context.Background()); err != nil {
		log.Printf("[WARN] Failed to create collection indexes: %v", err)
	}

	// Register Endpoints
	handler.RegisterRoutes(mux, authTeam)
}
