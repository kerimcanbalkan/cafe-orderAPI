package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/routes"
)

func main() {
	// Initialize MongoDB client
	client, err := db.NewClient(config.Env.DatabaseURI)
	if err != nil {
		log.Fatalf("Error initializing MongoDB client %v", err)
	}

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
