package table

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

// CreateTable creates a new table and saves it in the database
//
// @Summary Create a new table
// @Description Creates a new table with a name and assigns a uniqueCode name to it
// @Tags table
// @Param table body Table
// @Success 200 {object} map[string]interface{} "Table created successfully"
// @Failure 400  "Invalid request"
// @Failure 500  "Internal Server Error"
// @Router /table [post]
func CreateTable(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var table Table

		// Bind the request body to the order struct
		if err := c.ShouldBindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		// Validate the struct
		if err := validateTable(validate, table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		table.CreatedAt = time.Now()
		
		// Get the collection
		collection := client.GetCollection(config.Env.DatabaseName, "tables")

		// Get context from the request
		ctx := c.Request.Context()

		// Insert the item into the database
		result, err := collection.InsertOne(ctx, table)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}


		c.JSON(http.StatusOK, gin.H{
			"message": "Table created successfuly",
			"id":      result.InsertedID,
		})
	}
}

// GetTables retrieves all menu items.
//
// @Summary Get all tables
// @Description Fetches the tables from the database
// @Tags table
// @Produce json
// @Success 200 {object} []Table "List of tables"
// @Failure 500
// @Router /table [get]
func GetTables(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tables []Table

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "tables")

		// Get context from the request
		ctx := c.Request.Context()

		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &tables); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		// Return the menu in the response
		c.JSON(http.StatusOK, gin.H{
			"data": tables,
		})
	}
}
