package collections

import (
	"net/http"

	pogCollectionsDB "pog/database/collections"
	pogConstantsDB "pog/database/constants"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule acts as a bootstrap for the 'collections' domain.
func InitModule(db *mongo.Database, mux *http.ServeMux) {
	// Initialize Dependencies
	repo := pogCollectionsDB.NewCollectionRepository(db)
	constRepo := pogConstantsDB.NewConstantRepository(db)
	service := NewCollectionService(repo, constRepo)
	handler := NewCollectionHandler(service)

	// Register Endpoints
	handler.RegisterRoutes(mux)
}
