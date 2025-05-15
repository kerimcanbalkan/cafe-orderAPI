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
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/table"
)

// SetupRoutes initializes and registers all API routes and middleware for the Gin engine.
func SetupRoutes(r *gin.Engine, client *db.MongoClient) {
	
	r.Use(CORSMiddleware())
	
	// Documentation
	r.GET("api/v1//swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		orderGroup.POST("/:tableID", order.CreateOrder(client))
		orderGroup.GET("/active/:tableID", order.GetActiveOrdersByTableID(client))
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
			"/stats",
			auth.Authenticate([]string{"admin"}),
			order.GetStatistics(client),
		)
		orderGroup.GET(
			"/statsMonthly",
			auth.Authenticate([]string{"admin"}),
			order.GetStatisticsMonthly(client),
		)

	}

	tableGroup := r.Group("/api/v1/table")
	{
		tableGroup.POST("", auth.Authenticate([]string{"admin"}), table.CreateTable(client))
		tableGroup.GET("", auth.Authenticate([]string{"admin","waiter", "cashier"}), table.GetTables(client))
		tableGroup.GET("/:id", table.GetTableById(client))
		tableGroup.DELETE("/:id", auth.Authenticate([]string{"admin"}), table.DeleteTable(client))
	}

	// User Routes
	userGroup := r.Group("/api/v1/user")
	{
		userGroup.POST("", auth.Authenticate([]string{"admin"}), user.CreateUser(client))
		userGroup.GET("", auth.Authenticate([]string{"admin"}), user.GetUsers(client))
		userGroup.GET(
			"/:id/stats",
			auth.Authenticate([]string{"admin", "waiter", "cashier"}),
			user.GetStatistics(client),
		)
		userGroup.GET(
			"/:id",
			auth.Authenticate([]string{"admin", "waiter", "cashier"}),
			user.GetUserById(client),
		)
		userGroup.GET(
			"/me",
			auth.Authenticate([]string{"admin", "waiter", "cashier"}),
			user.GetUserMe(client),
		)
		userGroup.POST("/login", user.Login(client))
		userGroup.DELETE("/:id", auth.Authenticate([]string{"admin"}), user.DeleteUser(client))
	}

	r.GET("/api/v1/events", auth.Authenticate([]string{"admin,cashier,waiter"}), sse.SseHandler)
}

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
