package collaborators

import (
	"net/http"
	pogCollaboratorsDB "pog/database/collaborators"
	pogCollectionsDB "pog/database/collections"
	pogRequestsDB "pog/database/requests"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitModule(db *mongo.Database, mux *http.ServeMux) {
	repo := pogCollaboratorsDB.NewCollaboratorRepository(db)
	collectionRepo := pogCollectionsDB.NewCollectionRepository(db)
	requestRepo := pogRequestsDB.NewRequestRepository(db)
	service := NewCollaboratorService(repo, collectionRepo, requestRepo)
	handler := NewCollaboratorHandler(service)

	handler.RegisterRoutes(mux)
}
