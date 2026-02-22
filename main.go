package main

import (
	"log"
	"net/http"

	"pog/handler"
	"pog/internal"
	"pog/middleware"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system env")
	}

	db, err := internal.ConnectMongo()
	if err != nil {
		log.Fatalf("[FATAL] MongoDB connection failed: %v", err)
	}
	log.Println("[INFO] Connected to MongoDB")

	mux := handler.CompileHandlers(db)

	chain := middleware.Chain(
		middleware.Recovery,
		middleware.Logger,
		middleware.CORS,
		middleware.JSONContentType,
	)
	app := chain(mux)

	port := internal.GetPort()
	log.Printf("[%s] Server starting on :%s", internal.AppName, port)
	if err := http.ListenAndServe(":"+port, app); err != nil {
		log.Fatalf("[FATAL] Server failed: %v", err)
	}
}
