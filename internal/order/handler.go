package order

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
		order.Status = false
		order.Served = false
		order.CreatedAt = time.Now()

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
			"message":  "Order created successfuly",
			"order_id": result.InsertedID,
		})
	}
}

// GetOrders retrieves all orders
//
// @Summary Get all orders
// @Description Retrieves all orders for admin, cashier, and waiter roles
// @Tags order
// @Security bearerToken
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
		status := c.Query("status")
		served := c.Query("served")
		table := c.Query("table")
		date := c.Query("date")

		// Build MongoDB query dynamically
		query := bson.D{}

		// Parse "status" query parameter (convert to boolean)
		if status != "" {
			status, err := strconv.ParseBool(status)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "status", Value: status})
		}

		// Parse "served" query parameter
		if served != "" {
			served, err := strconv.ParseBool(served)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "served", Value: served})
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
				"error": err,
			})
			return
		}

		// Return the menu in the response
		c.JSON(http.StatusOK, gin.H{
			"orders": orders,
		})
	}
}

// ServeOrder marks an order as served
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

		id, _ := primitive.ObjectIDFromHex(idParam)
		filter := bson.D{{Key: "_id", Value: id}}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "served", Value: true},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Serve status updated successfuly",
		})
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
				{Key: "status", Value: true},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
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
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Order updated succesfully",
		})
	}
}
