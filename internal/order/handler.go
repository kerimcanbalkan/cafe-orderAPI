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
func CreateOrder(client *db.MongoClient) gin.HandlerFunc {
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

func GetOrders(client *db.MongoClient) gin.HandlerFunc {
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

// ServeOrder changes served status to true
func ServeOrder(client *db.MongoClient) gin.HandlerFunc {
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

// CompleteOrder changes the status of order to true
func CompleteOrder(client *db.MongoClient) gin.HandlerFunc {
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

// CompleteOrder changes the status of order to true
func UpdateOrder(client *db.MongoClient) gin.HandlerFunc {
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
