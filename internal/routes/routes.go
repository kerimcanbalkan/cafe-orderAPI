package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
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

	r.Static("api/v1/images", "./uploads")

	r.GET("api/v1/menu", func(c *gin.Context) {
		menu.GetMenu(c, client)
	})
	r.POST("api/v1/menu", func(c *gin.Context) {
		menu.CreateMenuItem(c, client)
	})
	r.DELETE("api/v1/menu/:id", func(c *gin.Context) {
		menu.DeleteMenuItem(c, client)
	})
	r.GET("api/v1/menu/:id", func(c *gin.Context) {
		menu.GetMenuByID(c, client)
	})
}
