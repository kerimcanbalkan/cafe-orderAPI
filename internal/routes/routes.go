package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/order"
)

func SetupRoutes(r *gin.Engine, client *db.MongoClient) {
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Access-Control-Allow-Origin", "*") // Change "*" to specific origin in production
	})

	r.GET("api/v1/images/:filename", func(c *gin.Context) {
		menu.GetMenuItemImage(c)
	})

	r.GET("api/v1/menu", func(c *gin.Context) {
		menu.GetMenu(c, client)
	})

	// Middleware to add headers globally
	r.MaxMultipartMemory = 2 << 20
	r.POST("api/v1/menu", func(c *gin.Context) {
		menu.CreateMenuItem(c, client)
	})
	r.DELETE("api/v1/menu/:id", func(c *gin.Context) {
		menu.DeleteMenuItem(c, client)
	})
	r.GET("api/v1/menu/:id", func(c *gin.Context) {
		menu.GetMenuByID(c, client)
	})

	// Order routes
	r.POST("api/v1/order", func(c *gin.Context) {
		order.CreateOrder(c, client)
	})
	r.GET("api/v1/order", func(c *gin.Context) {
		order.GetOrders(c, client)
	})
	r.PATCH("api/v1/order/serve/:id", func(c *gin.Context) {
		order.ServeOrder(c, client)
	})
	r.PATCH("api/v1/order/complete/:id", func(c *gin.Context) {
		order.CompleteOrder(c, client)
	})
}
