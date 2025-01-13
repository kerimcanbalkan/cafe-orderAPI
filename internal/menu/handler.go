package menu

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
)

var validate = validator.New()

type MenuItem struct {
	ID          string  `bson:"_id,omitempty" json:"id"`
	Name        string  `bson:"name"          json:"name"        validate:"required"`
	Description string  `bson:"description"   json:"description" validate:"required"`
	Price       float32 `bson:"price"         json:"price"       validate:"required"`
	Category    string  `bson:"category"      json:"category"    validate:"required"`
	Img         string  `bson:"image"         json:"image"       validate:"required"`
}

func GetMenu(c *gin.Context, client *db.MongoClient) {
	var menu []MenuItem

	// Get the collection from the database
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	// Set a context with a timeout for the query
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Find all documents in the menu collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not fetch the menu",
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode the results into the menu slice
	if err := cursor.All(ctx, &menu); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Could not find any menus",
		})
		return
	}

	// Return the menu in the response
	c.JSON(http.StatusOK, gin.H{
		"menu": menu,
	})
}

// CreateMenu creates
func CreateMenuItem(c *gin.Context, client *db.MongoClient) {
	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	price := c.PostForm("price")
	category := c.PostForm("category")

	// Handle the image upload
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	// Save the image to the server
	imagePath := "uploads/" + file.Filename
	if err = c.SaveUploadedFile(file, imagePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save image"})
		return
	}

	// Convert price to float
	priceFloat, err := strconv.ParseFloat(price, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
		return
	}

	// Create a new menu item
	item := MenuItem{
		Name:        name,
		Description: description,
		Price:       float32(priceFloat),
		Category:    category,
		Img:         imagePath,
	}

	// Validate the struct
	if err = validate.Struct(item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the collection
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	// Insert the item into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add the item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item added successfully",
		"item_id": result.InsertedID,
	})
}

func DeleteMenuItem(c *gin.Context, client *db.MongoClient) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	// Get collection from db
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Retrieve the menu item to get the image path
	var menuItem MenuItem
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&menuItem)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Menu item not found",
		})
		return
	}

	// Delete the image from the /uploads directory
	if menuItem.Img != "" {
		err = os.Remove("./uploads/" + menuItem.Img)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not delete image file",
			})
			return
		}
	}

	// Now delete the menu item from the database
	_, err = collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not delete item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Menu item deleted successfully",
	})
}
