package requests

import (
	"net/http"

	pogCollectionsDB "pog/database/collections"
	pogRequestsDB "pog/database/requests"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule acts as a bootstrap for the 'requests' domain.
func InitModule(db *mongo.Database, mux *http.ServeMux, authTeam func(http.Handler) http.Handler) {
	// Initialize Dependencies
	repo := pogRequestsDB.NewRequestRepository(db)
	collectionRepo := pogCollectionsDB.NewCollectionRepository(db)
	service := NewRequestService(repo, collectionRepo)
	handler := NewRequestHandler(service)

	// Register Endpoints
	handler.RegisterRoutes(mux, authTeam)
}
