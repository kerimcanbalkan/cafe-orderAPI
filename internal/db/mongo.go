package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMongoClient interface {
	GetCollection(dbName, collectionName string) *mongo.Collection
	Disconnect() error
}

// MongoClient is the type of the MongoDB client struct.
type MongoClient struct {
	client *mongo.Client
}

// NewClient initializes and returns a new MongoDB client.
func NewClient(uri string) (*MongoClient, error) {
	clientOptions := options.Client().ApplyURI(uri)

	// Establish connection with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err // Return error if connection fails
	}

	// Ping the database to ensure the connection is successful
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")
	return &MongoClient{client: client}, nil
}

// GetCollection returns a MongoDB collection by name.
func (mc *MongoClient) GetCollection(dbName, collectionName string) *mongo.Collection {
	if mc.client == nil {
		log.Fatal("MongoClient is not initialized.")
	}
	return mc.client.Database(dbName).Collection(collectionName)
}

// Disconnect gracefully closes the MongoDB client connection.
func (mc *MongoClient) Disconnect() error {
	if mc.client == nil {
		return nil // No connection to disconnect
	}

	// Create a timeout for the disconnect operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mc.client.Disconnect(ctx)
}
