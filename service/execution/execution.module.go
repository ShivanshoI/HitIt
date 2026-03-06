package execution

import (
	"net/http"
	"pog/database/history"
	"pog/database/requests"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitModule(db *mongo.Database, mux *http.ServeMux, authTeam func(http.Handler) http.Handler) {
	requestRepo := requests.NewRequestRepository(db)
	historyRepo := history.NewHistoryRepository(db)
	service := NewExecutionService(requestRepo, historyRepo)
	handler := NewExecutionHandler(service)

	handler.RegisterRoutes(mux, authTeam)
}
