package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/order"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/user"
)

func SetupRoutes(r *gin.Engine, client *db.MongoClient) {
	// Middleware to add headers globally
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Access-Control-Allow-Origin", "*") // Change "*" to specific origin in production
	})

	// Serving Images
	r.GET("api/v1/images/:filename", menu.GetMenuItemImage)

	// Menu routes
	r.GET("api/v1/menu", menu.GetMenu(client))

	r.MaxMultipartMemory = 2 << 20
	r.POST("api/v1/menu", menu.CreateMenuItem(client))
	r.DELETE("api/v1/menu/:id", menu.DeleteMenuItem(client))
	r.GET("api/v1/menu/:id", menu.GetMenuByID(client))

	// Order routes
	r.POST("api/v1/order/:table", order.CreateOrder(client))
	r.GET("api/v1/order", order.GetOrders(client))
	r.PATCH("api/v1/order/:id", order.UpdateOrder(client))
	r.PATCH("api/v1/order/serve/:id", order.ServeOrder(client))
	r.PATCH("api/v1/order/complete/:id", order.CompleteOrder(client))

	// User routes
	r.POST("api/v1/user", user.CreateUser(client))
	r.GET("api/v1/user", user.GetUsers(client))
	r.POST("api/v1/user/login", user.Login(client))
}
