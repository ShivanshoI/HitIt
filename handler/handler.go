package handler

import (
	"net/http"

	"pog/internal"
	pogActivityFeedSVC "pog/service/activity_feed"
	pogCollaboratorsSVC "pog/service/collaborators"
	"pog/service/collections"
	pogExecutionSVC "pog/service/execution"
	pogPlansSVC "pog/service/plans"
	pogProfileSVC "pog/service/profile"
	pogRequestsSVC "pog/service/requests"
	pogTeamsSVC "pog/service/teams"
	pogUsersSVC "pog/service/users"
	pogWebsocketsSVC "pog/service/websockets"

	"go.mongodb.org/mongo-driver/mongo"
)

func CompileHandlers(db *mongo.Database) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+internal.APIPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Teams Module Setup (must be first — returns authTeam middleware)
	authTeam := pogTeamsSVC.InitModule(db, mux)

	// Users Module Setup
	pogUsersSVC.InitModule(db, mux)

	// Profile Module Setup (stats, activity, update profile/password, sessions)
	pogProfileSVC.InitModule(db, mux)

	// Plans + Billing Module Setup (plans, subscription, payment method, invoices)
	pogPlansSVC.InitModule(db, mux)

	// Collections Module Setup (team-scoped)
	collections.InitModule(db, mux, authTeam)

	// Requests Module Setup (team-scoped)
	pogRequestsSVC.InitModule(db, mux, authTeam)

	// Collaborators Module Setup
	pogCollaboratorsSVC.InitModule(db, mux)

	// Websockets Module Setup
	wsHub := pogWebsocketsSVC.InitModule(mux)

	// Activity Feed Module Setup
	pogActivityFeedSVC.InitModule(db, mux, wsHub)

	// Execution Module Setup (team-scoped)
	pogExecutionSVC.InitModule(db, mux, authTeam)

	return mux
}

