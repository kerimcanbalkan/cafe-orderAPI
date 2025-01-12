package main

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/routes"
)

func main() {
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run()
}
