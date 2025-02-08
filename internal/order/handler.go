package order

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/sse"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

type orderRequest struct {
	Items []menu.MenuItem `json:"items" validate:"required"`
}

// CreateOrder creates an order and saves it in the database
//
// @Summary Create a new order
// @Description Creates a new order for a specific table and saves it in the database
// @Tags order
// @Param table path int true "Table number"
// @Param order body orderRequest true "Order details"
// @Success 200 {object} map[string]interface{} "Order created successfully"
// @Failure 400  "Invalid request"
// @Failure 500  "Internal Server Error"
// @Router /order/{table} [post]
func CreateOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tableStr := c.Param("table")
		table, err := convertTableNumber(tableStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid table number",
			})
			return
		}
		var request orderRequest

		// Bind the request body to the order struct
		if err = c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		if len(request.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order should include items",
			})
			return
		}

		totalPrice := float64(0)

		// Validate Items
		for _, item := range request.Items {
			if err = menu.ValidateMenu(validate, item); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"Validation failed for item %s: %s",
						item.Name,
						err.Error(),
					),
				})
				return
			}
			totalPrice += item.Price
		}

		// Validate the struct
		if err = validateOrder(validate, request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order := &Order{}

		order.Items = request.Items
		order.TableNumber = uint8(table)
		order.ClosedAt = nil
		order.CreatedAt = time.Now()
		order.ServedAt = nil
		order.HandledBy = primitive.NilObjectID
		order.ClosedBy = primitive.NilObjectID
		order.TotalPrice = totalPrice

		// Get the collection
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Insert the item into the database
		result, err := collection.InsertOne(ctx, order)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		sse.Notify(table)

		c.JSON(http.StatusOK, gin.H{
			"message": "Order created successfuly",
			"id":      result.InsertedID,
		})
	}
}

// GetOrders retrieves all orders
//
// @Summary Get all orders
// @Description Retrieves all orders for admin, cashier, and waiter roles
// @Tags order
// @Security bearerToken
// @Param is_closed query boolean false "Filter by closed status (true/false)"
// @Param served query boolean false "Filter by served status (true/false)"
// @Param table query int false "Filter by table number"
// @Param date query string false "Filter by order date (YYYY-MM-DD)"
// @Success 200 {array} Order "List of orders with their IDs"
// @Failure 500 "Internal Server Error"
// @Router /order [get]
func GetOrders(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []Order

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Get query parameters
		isClosed := c.Query("is_closed")
		served := c.Query("served")
		table := c.Query("table")
		date := c.Query("date")

		// Build MongoDB query dynamically
		query := bson.D{}

		// Parse "status" query parameter (convert to boolean)
		if isClosed != "" {
			bool, err := strconv.ParseBool(isClosed)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "closedAt", Value: bson.M{"$exists": bool}})
		}

		// Parse "served" query parameter
		if served != "" {
			bool, err := strconv.ParseBool(served)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "servedAt", Value: bson.M{"$exists": bool}})
		}

		// Add table filter
		if table != "" {
			table, err := strconv.Atoi(table)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid table number."},
				)
				return
			}
			query = append(query, bson.E{Key: "tableNumber", Value: table})
		}

		// Parse "date" query parameter
		if date != "" {
			parsedDate, err := time.Parse("2006-01-02", date)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid date format. Use YYYY-MM-DD."},
				)
				return
			}

			// Calculate start and end of the day
			startOfDay := primitive.NewDateTimeFromTime(parsedDate)
			endOfDay := primitive.NewDateTimeFromTime(parsedDate.Add(24 * time.Hour))
			query = append(query, bson.E{
				Key: "createdAt",
				Value: bson.D{
					{Key: "$gte", Value: startOfDay},
					{Key: "$lt", Value: endOfDay},
				},
			})
		}

		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, query)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		// Return the orders in the response
		c.JSON(http.StatusOK, gin.H{
			"data": orders,
		})
	}
}

// ServeOrder sets the servedAt and HandledBy fields.
// @Summary Mark an order as served
// @Description Allows admin and waiter roles to mark an order as served
// @Tags order
// @Param id path string true "Order ID"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order served successfully"
// @Failure 404 "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/serve/{id} [patch]
func ServeOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		// Convert idParam to ObjectID
		id, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Order ID!",
			})
			return
		}

		// Get claims from Gin context
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Type assert to jwt.MapClaims
		jwtClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token data"})
			return
		}

		// Extract UserID
		userIDHex, ok := jwtClaims["UserID"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		// Convert the string back to primitive.ObjectID
		userID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
			return
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")
		ctx := c.Request.Context()

		// Prepare the update statement
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "servedAt", Value: time.Now()},
				{Key: "handledBy", Value: userID},
			}},
		}

		// Find order and if exists update
		result := collection.FindOneAndUpdate(ctx, bson.D{{Key: "_id", Value: id}}, update)

		var order Order
		err = result.Decode(&order)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order served successfully"})
	}
}

// CloseOrder marks an order as complete
//
// @Summary Mark an order as complete
// @Description Allows admin and cashier roles to mark an order as complete
// @Tags order
// @Param id path string true "Order ID"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order completed successfully"
// @Failure 404  "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/complete/{id} [patch]
func CloseOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		id, _ := primitive.ObjectIDFromHex(idParam)
		// Filters by id and checks if its served
		filter := bson.D{
			{Key: "_id", Value: id},
			{Key: "servedAt", Value: bson.M{"$exists": true}},
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "closedAt", Value: time.Now()},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		// Check if a document was actually updated
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or must be served first"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order closed succesfully",
		})
	}
}

// UpdateOrder updates an existing order
//
// @Summary Update an existing order
// @Description Allows admin, cashier, and waiter roles to update an order
// @Tags order
// @Param id path string true "Order ID"
// @Param order body orderRequest true "Order update details"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order updated successfully"
// @Failure 400  "Invalid request"
// @Failure 404  "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/{id} [patch]
func UpdateOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		id, _ := primitive.ObjectIDFromHex(idParam)
		filter := bson.D{{Key: "_id", Value: id}}

		var request orderRequest

		// Bind the request body to the order struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		totalPrice := float64(0)

		// Validate Items
		for _, item := range request.Items {
			if err := menu.ValidateMenu(validate, item); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"Validation failed for item %s: %s",
						item.Name,
						err.Error(),
					),
				})
				return
			}
			totalPrice += item.Price
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "items", Value: request.Items},
				{Key: "totalPrice", Value: totalPrice},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		// Check if a document was actually updated
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Order updated succesfully",
		})
	}
}

// GetMonthlyStatistics retrieves monthly order statistics.
//
// @Summary Get monthly statistics
// @Description Fetches order statistics for a given year and month. If no year or month is provided, the current year and month are used as defaults.
// @Tags Statistics
// @Security bearerToken
// @Accept json
// @Produce json
// @Param year query string false "Year in YYYY format (defaults to current year)" example(2025)
// @Param month query string false "Month in M format (1-12, defaults to current month)" example(1)
// @Success 200 {object} map[string]interface{} "Monthly statistics data"
// @Failure 400 {object} map[string]string "Invalid month/year format"
// @Failure 500 {object} map[string]string "Failed to fetch statistics"
// @Router /order/stats/monthly [get]
func GetMonthlyStatistics(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		yearStr := c.DefaultQuery("year", time.Now().Format("2006"))
		monthStr := c.DefaultQuery("month", time.Now().Format("1"))

		yearInt, err := strconv.Atoi(yearStr)
		monthInt, err := strconv.Atoi(monthStr)
		if err != nil || monthInt < 1 || monthInt > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month/year format"})
			return
		}

		startDate := time.Date(yearInt, time.Month(monthInt), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0)

		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		stats, err := monthlyStats(startDate, endDate, c, collection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
			return
		}

		// Return the stats in the response
		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}
