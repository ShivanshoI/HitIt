package users

import (
	"net/http"

	pogUsersDB "pog/database/users"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule acts as a bootstrap for the 'users' domain.
// It sets up the repository, service, and handler layers, binding routes to the mux.
func InitModule(db *mongo.Database, mux *http.ServeMux) {
	// Initialize Dependencies
	repo := pogUsersDB.NewUserRepository(db)
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	// Register Endpoints
	handler.RegisterRoutes(mux)
}
