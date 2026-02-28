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

// ConnectMongo connects to MongoDB and returns the database instance
// called once in main.go and passed down to all modules
func ConnectMongo() (*mongo.Database, error) {
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB")

	// 10 second timeout for initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	log.Printf("[DB] Successfully connected to database: %s", dbName)
	return client.Database(dbName), nil
}
