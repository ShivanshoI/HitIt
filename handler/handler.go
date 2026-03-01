package handler

import (
	"net/http"

	"pog/internal"
	pogCollaboratorsSVC "pog/service/collaborators"
	"pog/service/collections"
	pogRequestsSVC "pog/service/requests"
	pogUsersSVC "pog/service/users"

	"go.mongodb.org/mongo-driver/mongo"
)

func CompileHandlers(db *mongo.Database) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+internal.APIPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Users Module Setup
	pogUsersSVC.InitModule(db, mux)

	// Collections Module Setup
	collections.InitModule(db, mux)

	// Requests Module Setup
	pogRequestsSVC.InitModule(db, mux)

	// Collaborators Module Setup
	pogCollaboratorsSVC.InitModule(db, mux)

	return mux
}
