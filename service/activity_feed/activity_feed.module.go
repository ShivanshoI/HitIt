package activity_feed

import (
	"net/http"

	pogActivityFeedDB "pog/database/activity_feed"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule wires all activity-feed dependencies and registers routes on mux.
// It is the single entry-point for the module; callers need nothing else.
func InitModule(db *mongo.Database, mux *http.ServeMux) {
	hub := NewHub()
	go hub.Run()

	repo := pogActivityFeedDB.NewMessageRepository(db)
	service := NewActivityFeedService(repo, hub)
	handler := NewActivityFeedHandler(service)

	handler.RegisterRoutes(mux)
}