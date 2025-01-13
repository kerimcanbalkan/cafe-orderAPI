package main

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/routes"
)

func main() {
	c := config.LoadConfig()

	db.ConnectMongoDB(c.DatabaseURI)

	r := gin.Default()
	routes.SetupRoutes(r)

	r.Run(":" + c.ServerPort)
}
