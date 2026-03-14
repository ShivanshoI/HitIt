package organizations

import (
	"log"
	"net/http"

	orgDB "pog/database/organizations"
	mappingDB "pog/database/user_mapping"
	userDB "pog/database/users"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule bootstraps the organizations domain.
func InitModule(db *mongo.Database, mux *http.ServeMux) func(http.Handler) http.Handler {
	log.Println("[MODULE] Initializing Organizations Module")

	orgRepo := orgDB.NewOrganizationRepository(db)
	userRepo := userDB.NewUserRepository(db)
	mappingRepo := mappingDB.NewUserMappingRepository(db)

	svc := NewOrganizationService(orgRepo, userRepo, mappingRepo)
	handler := NewOrganizationHandler(svc)

	handler.RegisterRoutes(mux)
	log.Println("[MODULE] Organizations Module Initialized")

	return AuthWithOrg(db)
}
