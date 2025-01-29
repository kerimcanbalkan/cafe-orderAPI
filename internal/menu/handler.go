package menu

import (
	"fmt"
	"log"
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
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

func GetMenu(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu []MenuItem

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "menu")

		// Get context from the request
		ctx := c.Request.Context()

		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			handleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &menu); err != nil {
			handleMongoError(c, err)
			return
		}

		// Return the menu in the response
		c.JSON(http.StatusOK, gin.H{
			"menu": menu,
		})
	}
}

// CreateMenu creates
func CreateMenuItem(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		imagePath := "uploads/" + generateImageName()
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
		if err = ValidateMenu(validate, item); err != nil {
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
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Item added successfully",
			"item_id": result.InsertedID,
		})
	}
}

func DeleteMenuItem(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			utils.HandleMongoError(c, err)
			return
		}

		// Delete the image from the /uploads directory
		if menuItem.Img != "" {
			err = os.Remove(menuItem.Img)
			if err != nil {
				if !os.IsNotExist(err) {

					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Could not remove related files",
					})
					return
				}
			}
		}

		// Now delete the menu item from the database
		_, err = collection.DeleteOne(ctx, bson.M{"_id": docID})
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func GetMenuByID(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
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
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"item": menuItem,
		})
	}
}

func GetMenuItemImage(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	filePath := "./uploads/" + filename
	log.Println(os.Stat(filePath))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
		})
	}

	// Determine content type
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open image"})
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read image"})
		return
	}

	contentType := http.DetectContentType(buffer)
	c.Header("Content-Type", contentType)
	c.File("./uploads/" + filename)
}
