package handler

import (
	"net/http"

	"pog/internal"
	pogActivityFeedSVC "pog/service/activity_feed"
	pogCollaboratorsSVC "pog/service/collaborators"
	"pog/service/collections"
	pogExecutionSVC "pog/service/execution"
	pogOrganizationsSVC "pog/service/organizations"
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

	// Organizations Module Setup (returns authOrg middleware)
	authOrg := pogOrganizationsSVC.InitModule(db, mux)

	// Teams Module Setup (must be first — returns authTeam middleware)
	authTeam := pogTeamsSVC.InitModule(db, mux, authOrg)

	// Users Module Setup
	pogUsersSVC.InitModule(db, mux)

	// Profile Module Setup (stats, activity, update profile/password, sessions)
	pogProfileSVC.InitModule(db, mux)

	// Plans + Billing Module Setup (plans, subscription, payment method, invoices)
	pogPlansSVC.InitModule(db, mux)

	// Combined middleware for team-scoped modules
	authOrgAndTeam := func(h http.Handler) http.Handler {
		return authOrg(authTeam(h))
	}

	// Collections Module Setup (team-scoped)
	collections.InitModule(db, mux, authOrgAndTeam)

	// Requests Module Setup (team-scoped)
	pogRequestsSVC.InitModule(db, mux, authOrgAndTeam)

	// Collaborators Module Setup
	pogCollaboratorsSVC.InitModule(db, mux)

	// Websockets Module Setup
	wsHub := pogWebsocketsSVC.InitModule(mux)

	// Activity Feed Module Setup
	pogActivityFeedSVC.InitModule(db, mux, wsHub)

	// Execution Module Setup (team-scoped)
	pogExecutionSVC.InitModule(db, mux, authOrgAndTeam)

	return mux
}

