package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// MockMongoClient is a mock implementation of the IMongoClient interface.
type MockMongoClient struct {
	collection *mongo.Collection
}

// NewMockMongoClient creates a new instance of MockMongoClient.
func NewMockMongoClient(collection *mongo.Collection) *MockMongoClient {
	return &MockMongoClient{
		collection: collection,
	}
}

// GetCollection mocks the GetCollection method of IMongoClient.
func (m *MockMongoClient) GetCollection(dbName, collectionName string) *mongo.Collection {
	return m.collection
}

// Disconnect mocks the Disconnect method of IMongoClient.
func (m *MockMongoClient) Disconnect() error {
	// No-op for mock
	return nil
}
