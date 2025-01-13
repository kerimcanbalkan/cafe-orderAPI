package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

func SetupRoutes(r *gin.Engine, client *db.MongoClient) {
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
