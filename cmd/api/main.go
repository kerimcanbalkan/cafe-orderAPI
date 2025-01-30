// @title Cafe Order API
// @version 1
// @description An API for cafe owners to manage menus, process orders, and oversee fulfillment.
// Admins can create, update, view, and delete menu items. Customers can browse the menu and place orders.
// Waiters can serve orders, and cashiers can complete them upon payment.
// The API implements authentication and role-based access control for secure operations.

// @contact.name Kerimcan Balkan
// @contact.url https://github.com/kerimcanbalkan
// @contact.email kerimcanbalkan@gmail.com

// @securityDefinitions.apikey JwtAuth
// @in header
// @name Authorization
// @description JWT token required to authenticate users. The token must include a valid role (e.g., "admin", "cashier", "waiter").

// @host localhost:8000
// @BasePath /api/v1

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
