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
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

// CreateOrder creates an order and saves it in the database
//
// @Summary Create a new order
// @Description Creates a new order for a specific table and saves it in the database
// @Tags order
// @Param table path int true "Table number"
// @Param order body Order true "Order details"
// @Success 200 {object} map[string]interface{} "Order created successfully"
// @Failure 400  "Invalid request"
// @Failure 500  "Internal Server Error"
// @Router /order/{table} [post]
func CreateOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tableStr := c.Param("table")
		table, err := strconv.Atoi(tableStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid table number",
			})
			return
		}
		var order Order

		// Bind the request body to the order struct
		if err = c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		// Validate the struct
		if err = validateOrder(validate, order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.TableNumber = table
		order.IsClosed = false
		order.CreatedAt = time.Now()
		order.ServedAt = nil
		order.HandledBy = primitive.NilObjectID

		var totalPrice float32 = 0.0
		for _, p := range order.Items {
			totalPrice += p.Price
		}
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
			isClosedBool, err := strconv.ParseBool(isClosed)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "isClosed", Value: isClosedBool})
		}

		// Parse "served" query parameter
		if served != "" {
			servedBool, err := strconv.ParseBool(served)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			if servedBool {
				// Orders that have been served (servedAt exists)
				query = append(query, bson.E{Key: "servedAt", Value: bson.M{"$exists": true}})
			} else {
				// Orders that have NOT been served (servedAt does not exist)
				query = append(query, bson.E{Key: "servedAt", Value: bson.M{"$exists": false}})
			}
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

		// Return the menu in the response
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

		// Find the existing order to check the current values
		var order Order
		err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&order)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		// If servedAt and handledBy are already set as desired, skip update
		if order.ServedAt != nil && order.HandledBy == userID {
			c.JSON(http.StatusOK, gin.H{"message": "Order already served"})
			return
		}

		// Prepare the update statement
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "servedAt", Value: time.Now()},
				{Key: "handledBy", Value: userID},
			}},
		}

		// Perform the update
		_, err = collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order served successfully"})
	}
}

// CompleteOrder marks an order as complete
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
func CompleteOrder(client db.IMongoClient) gin.HandlerFunc {
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

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "isClosed", Value: true},
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
			"message": "Complete status updated successfuly",
		})
	}
}

// UpdateOrder updates an existing order
//
// @Summary Update an existing order
// @Description Allows admin, cashier, and waiter roles to update an order
// @Tags order
// @Param id path string true "Order ID"
// @Param order body Order true "Order update details"
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

		var orderItems []menu.MenuItem

		// Bind the request body to the order struct
		if err := c.ShouldBindJSON(&orderItems); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		totalPrice := float32(0)

		// Validate Items
		for _, item := range orderItems {
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
				{Key: "items", Value: orderItems},
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
