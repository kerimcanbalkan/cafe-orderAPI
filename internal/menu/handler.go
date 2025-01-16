package menu

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
)

var validate = validator.New()

type MenuItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name"          json:"name"        validate:"required"`
	Description string             `bson:"description"   json:"description" validate:"required"`
	Price       float32            `bson:"price"         json:"price"       validate:"required"`
	Category    string             `bson:"category"      json:"category"    validate:"required"`
	Img         string             `bson:"image"         json:"image"       validate:"required"`
}

func GetMenu(c *gin.Context, client *db.MongoClient) {
	var menu []MenuItem

	// Get the collection from the database
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	// Get context from the request
	ctx := c.Request.Context()

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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	price := c.PostForm("price")
	category := c.PostForm("category")

	// Handle the image upload
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println(err.Error())
		if err.Error() == "multipart: NextPart: http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("Max request body size is %v bytes\n", 2<<20),
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	mimeType := file.Header.Get("Content-Type")

	// Validate image mime-type is allowable
	if valid := isAllowedImageType(mimeType); !valid {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Invalid File format, must be 'image/jpeg' or 'image/png'"},
		)
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

	// Get context from the request
	ctx := c.Request.Context()

	// Insert the item into the database
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

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	// Get collection from db
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	// Get context from the request
	ctx := c.Request.Context()

	// Retrieve the menu item to get the image path
	var menuItem MenuItem
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&menuItem)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Menu item not found",
		})
		return
	}

	// Delete the image from the /uploads directory
	if menuItem.Img != "" {
		err = os.Remove(menuItem.Img)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not remove related files",
			})
			return
		}
	}

	// Now delete the menu item from the database
	_, err = collection.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not delete item",
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetMenuByID(c *gin.Context, client *db.MongoClient) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID!",
		})
		return
	}

	// Get collection from db
	collection := client.GetCollection(config.Env.DatabaseName, "menu")

	// Get context from the request
	ctx := c.Request.Context()

	// Retrieve the menu item
	var menuItem MenuItem
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&menuItem)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Menu item not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"item": menuItem,
	})
}

func GetMenuItemImage(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	filePath := "./uploads/" + filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
		})
	}
	c.Header("Content-Type", "image/jpeg")
	c.File("./uploads/" + filename)
}
