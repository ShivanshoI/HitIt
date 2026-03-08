package execution

import (
	"log"
	"net/http"
	"os"
	"time"

	"pog/database/history"
	"pog/database/requests"
	"pog/internal/localbridge"

	"go.mongodb.org/mongo-driver/mongo"
)

func initBridgeClient() *localbridge.Client {
	bridgeURL := os.Getenv("BRIDGE_URL")
	if bridgeURL == "" {
		log.Println("[bridge] BRIDGE_URL not set — localhost requests will be rejected")
		return nil
	}

	client := localbridge.New(localbridge.Config{
		ServerURL: bridgeURL,
		AuthToken: os.Getenv("BRIDGE_TOKEN"),
		Timeout:   30 * time.Second,
	})

	log.Println("[bridge] ✓ bridge client configured, extension status will be checked at request time")
	return client
}

func InitModule(db *mongo.Database, mux *http.ServeMux, authTeam func(http.Handler) http.Handler) {
	requestRepo := requests.NewRequestRepository(db)
	historyRepo := history.NewHistoryRepository(db)
	bridge := initBridgeClient()
	service := NewExecutionService(requestRepo, historyRepo, bridge)
	handler := NewExecutionHandler(service)

	handler.RegisterRoutes(mux, authTeam)
}
