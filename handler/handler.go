package handler

import (
	"net/http"
	"pog/internal"

	"go.mongodb.org/mongo-driver/mongo"
)

func CompileHandlers(db *mongo.Database) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+internal.APIPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return mux
}
