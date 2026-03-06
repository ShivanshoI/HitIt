package plans

import (
	"net/http"

	pogBillingDB "pog/database/billing"

	"go.mongodb.org/mongo-driver/mongo"
)

// InitModule bootstraps the plans/billing domain.
func InitModule(db *mongo.Database, mux *http.ServeMux) {
	billingRepo := pogBillingDB.NewBillingRepository(db)
	service := NewPlansService(billingRepo)
	handler := NewPlansHandler(service)
	handler.RegisterRoutes(mux)
}
