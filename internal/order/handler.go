package order

import (
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
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Items       []menu.MenuItem    `bson:"items"         json:"items"       validate:"required"`
	TotalPrice  float32            `bson:"totalPrice"    json:"totalPrice"`
	TableNumber int                `bson:"tableNumber"   json:"tableNumber"`
	Status      bool               `bson:"status"        json:"status"`
	Served      bool               `bson:"served"        json:"served"`
	CreatedAt   time.Time          `bson:"createdAt"     json:"createdAt"`
}

var validate = validator.New()

// CreateOrder creates an order and saves it in the database
func CreateOrder(c *gin.Context, client *db.MongoClient) {
	tableStr := c.Param("table")
	table, err := strconv.Atoi(tableStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid table number",
		})
		return
	}
	var order Order

	// Bind the request body to the order structure
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	// Validate the order fields
	if err := validate.Struct(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Order created successfuly",
		"order_id": result.InsertedID,
	})
}

func GetOrders(c *gin.Context, client *db.MongoClient) {
	var orders []Order

	// Get the collection from the database
	collection := client.GetCollection(config.Env.DatabaseName, "orders")

	// Get context from the request
	ctx := c.Request.Context()

	// Find all documents in the menu collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
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

func ServeOrder(c *gin.Context, client *db.MongoClient) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	id, _ := primitive.ObjectIDFromHex(idParam)
	filter := bson.D{{"_id", id}}

	update := bson.D{{"$set", bson.D{{"served", true}}}}
	// Get the collection from the database
	collection := client.GetCollection(config.Env.DatabaseName, "orders")

	// Get context from the request
	ctx := c.Request.Context()

	// Updates the first document that has the specified "_id" value
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Serve status updated successfuly",
	})
}

func CompleteOrder(c *gin.Context, client *db.MongoClient) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	id, _ := primitive.ObjectIDFromHex(idParam)
	filter := bson.D{{"_id", id}}

	update := bson.D{{"$set", bson.D{{"status", true}}}}
	// Get the collection from the database
	collection := client.GetCollection(config.Env.DatabaseName, "orders")

	// Get context from the request
	ctx := c.Request.Context()

	// Updates the first document that has the specified "_id" value
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Complete status updated successfuly",
	})
}
