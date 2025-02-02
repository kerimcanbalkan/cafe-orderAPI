package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates necessary indexes for the MongoDB collections to
// enforce uniqueness and improve query performance.
func EnsureIndexes(client IMongoClient, ctx context.Context, dbName string) {
	// Indexes for users collection
	usersCollection := client.GetCollection(dbName, "users")

	userIndexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := usersCollection.Indexes().CreateMany(ctx, userIndexModels)
	if err != nil {
		log.Fatalf("Failed to create indexes for users: %v", err)
	}

	// Indexes for menu collection
	menuCollection := client.GetCollection(dbName, "menu")

	menuIndexModels := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = menuCollection.Indexes().CreateOne(ctx, menuIndexModels)
	if err != nil {
		log.Fatalf("Failed to create indexes for users: %v", err)
	}

	log.Println("Indexes ensured successfully!")
}
