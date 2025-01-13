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

	r.Static("/images", "./uploads")

	r.GET("/menu", func(c *gin.Context) {
		menu.GetMenu(c, client)
	})
	r.POST("/menu", func(c *gin.Context) {
		menu.CreateMenuItem(c, client)
	})
	r.DELETE("/menu/:id", func(c *gin.Context) {
		menu.CreateMenuItem(c, client)
	})
}
