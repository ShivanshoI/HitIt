package activity_feed

import (
	"context"
	"log"
	"net/http"

	pogActivityFeedDB "pog/database/activity_feed"
	"pog/service/websockets"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule wires all activity-feed dependencies and registers routes on mux.
// It is the single entry-point for the module; callers need nothing else.
func InitModule(db *mongo.Database, mux *http.ServeMux, wsHub *websockets.Hub) {
	repo := pogActivityFeedDB.NewMessageRepository(db)

	ctx := context.Background()
	if err := repo.EnsureIndexes(ctx); err != nil {
		log.Printf("[WARN] Failed to create activity feed indexes: %v", err)
	}

	service := NewActivityFeedService(repo, wsHub)
	handler := NewActivityFeedHandler(service)

	handler.RegisterRoutes(mux)
}