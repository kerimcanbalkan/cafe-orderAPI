package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/menu", menu.GetMenu)
}
