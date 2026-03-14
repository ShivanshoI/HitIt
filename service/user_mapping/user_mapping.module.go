package user_mapping

import (
	"context"
	"log"
	"net/http"

	mappingDB "pog/database/user_mapping"
	"go.mongodb.org/mongo-driver/mongo"
)

func InitModule(db *mongo.Database, mux *http.ServeMux) {
	repo := mappingDB.NewUserMappingRepository(db)
	
	// Ensure indexes are created
	ctx := context.Background()
	if err := repo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create user mapping indexes: %v", err)
	}

	service := NewUserMappingService(repo)
	_ = service // Will be used by handlers or exported via middleware
	
	// TODO: Register routes if needed (e.g., /api/user/mappings)
}
