package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/routes"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/user"
)

func main() {
	// Initialize MongoDB client
	client, err := db.NewClient(config.Env.DatabaseURI)
	if err != nil {
		log.Fatalf("Error initializing MongoDB client %v", err)
	}

	// Create a root context for the application
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user.SeedAdminUser(client, rootCtx)

	// Setup gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, client)

	// Start the server
	r.Run(":" + config.Env.ServerPort)

	// Disconnect after server shutdown
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
	}()
}
