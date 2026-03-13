package internal

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetPort returns PORT from env or defaults to 8080
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func ConnectMongo() (*mongo.Database, error) {
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("[DB] Attempting to connect to MongoDB URI (redacted)...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetServerSelectionTimeout(10*time.Second))
	if err != nil {
		log.Printf("[ERROR] MongoDB client initialization failed: %v", err)
		return nil, err
	}

	log.Printf("[DB] Pinging MongoDB...")
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("[ERROR] MongoDB ping failed: %v", err)
		return nil, err
	}

	log.Printf("[DB] Successfully connected to database: %s", dbName)
	return client.Database(dbName), nil
}
