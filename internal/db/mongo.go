package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// ConnectMongoDB establishes a connection to the MongoDB database.
func ConnectMongoDB(uri string) {
	clientOptions := options.Client().ApplyURI(uri)

	// Establish connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database to ensure connection is successful
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Could not ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	MongoClient = client
}

// GetCollection returns a MongoDB collection by name.
func GetCollection(dbName, collectionName string) *mongo.Collection {
	if MongoClient == nil {
		log.Fatal("MongoClient is not initialized. Call ConnectMongoDB first.")
	}
	return MongoClient.Database(dbName).Collection(collectionName)
}
