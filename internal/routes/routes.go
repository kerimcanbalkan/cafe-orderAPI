package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/kerimcanbalkan/cafe-orderAPI/docs"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/auth"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/order"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/sse"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/user"
)

// SetupRoutes initializes and registers all API routes and middleware for the Gin engine.
func SetupRoutes(r *gin.Engine, client *db.MongoClient) {
	// Middleware to add headers globally
	r.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Access-Control-Allow-Origin", "*") // Change "*" to specific origin in production
	})

	// Documentation
	r.GET("api/v1//swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serving Images
	r.GET("api/v1/images/:filename", menu.GetMenuItemImage)

	// Menu Routes
	menuGroup := r.Group("/api/v1/menu")
	{
		menuGroup.GET("", menu.GetMenu(client))
		menuGroup.POST("", auth.Authenticate([]string{"admin"}), menu.CreateMenuItem(client))
		menuGroup.DELETE("/:id", auth.Authenticate([]string{"admin"}), menu.DeleteMenuItem(client))
		menuGroup.GET("/images/:filename", menu.GetMenuItemImage)
	}

	// Order Routes
	orderGroup := r.Group("/api/v1/order")
	{
		orderGroup.POST("/:table", order.CreateOrder(client))
		orderGroup.GET(
			"",
			auth.Authenticate([]string{"admin", "cashier", "waiter"}),
			order.GetOrders(client),
		)
		orderGroup.PATCH(
			"/:id",
			auth.Authenticate([]string{"admin", "cashier", "waiter"}),
			order.UpdateOrder(client),
		)
		orderGroup.PATCH(
			"/serve/:id",
			auth.Authenticate([]string{"admin", "waiter"}),
			order.ServeOrder(client),
		)
		orderGroup.PATCH(
			"/close/:id",
			auth.Authenticate([]string{"admin", "cashier"}),
			order.CloseOrder(client),
		)
		orderGroup.GET(
			"/stats/monthly",
			auth.Authenticate([]string{"admin"}),
			order.GetMonthlyStatistics(client),
		)
		orderGroup.GET(
			"/stats/yearly",
			auth.Authenticate([]string{"admin"}),
			order.GetYearlyStatistics(client),
		)
		orderGroup.GET(
			"/stats/daily",
			auth.Authenticate([]string{"admin"}),
			order.GetDailyStatistics(client),
		)

	}

	// User Routes
	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("", auth.Authenticate([]string{"admin"}), user.CreateUser(client))
		userGroup.GET("", auth.Authenticate([]string{"admin"}), user.GetUsers(client))
		userGroup.GET(
			":id/stats/yearly",
			auth.Authenticate([]string{"admin", "waiter", "cashier"}),
			user.GetYearlyStatistics(client),
		)
		userGroup.GET(
			":id",
			auth.Authenticate([]string{"admin", "waiter", "cashier"}),
			user.GetUserById(client),
		)
		userGroup.POST("/login", user.Login(client))
		userGroup.DELETE("/:id", auth.Authenticate([]string{"admin"}), user.DeleteUser(client))
	}

	r.GET("/api/v1/events", sse.SseHandler)
}
