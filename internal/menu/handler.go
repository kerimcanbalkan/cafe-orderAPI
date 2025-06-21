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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

// GetMenu retrieves all menu items.
//
// @Summary Get all menu items
// @Description Fetches the entire menu from the database
// @Tags menu
// @Produce json
// @Success 200 {object} []MenuItem "List of menu items"
// @Failure 500
// @Router /menu [get]
func GetMenu(client db.IMongoClient) gin.HandlerFunc {
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		// Return the menu in the response
		c.JSON(http.StatusOK, gin.H{
			"data": menu,
		})
	}
}

// CreateMenuItem creates a new menu item.
//
// @Summary Create a new menu item
// @Description Adds a new item to the menu with an image upload. Only accessible by users with the "admin" role.
// @Tags menu
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Name of the item"
// @Param description formData string true "Description of the item"
// @Param price formData number true "Price of the item"
// @Param category formData string true "Category of the item"
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]interface{} "Item added successfully"
// @Failure 400  "Bad Request"
// @Failure 500 "Internal Server Error"
// @Security bearerToken
// @Router /menu [post]
func CreateMenuItem(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
		// Parse form data
		name := c.PostForm("name")
		description := c.PostForm("description")
		priceStr := c.PostForm("price")
		category := c.PostForm("category")
		currency := c.PostForm("currency")

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
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format"})
			return
		}

		img := filepath.Base(imagePath)

		// Create a new menu item
		item := MenuItem{
			Name:        name,
			Description: description,
			Price:       int64(price),
			Category:    category,
			Currency:    currency,
			Img:         img,
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
			if mongo.IsDuplicateKeyError(err) {
				c.JSON(http.StatusConflict, gin.H{
					"error": fmt.Sprintf("Menu item named %s already exists", item.Name),
				})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Item added successfully",
			"id":      result.InsertedID,
		})
	}
}

// DeleteMenuItem deletes a menu item by its ID.
// @Summary Delete a menu item
// @Description Deletes a menu item and its related image. Only accessible by users with the "admin" role.
// @Tags menu
// @Param id path string true "ID of the menu item to delete"
// @Success 200 {object} nil
// @Failure 400 "Bad Request"
// @Failure 404 "Menu item not found"
// @Failure 500 "Internal Server Error"
// @Security bearerToken
// @Router /menu/{id} [delete]
func DeleteMenuItem(client db.IMongoClient) gin.HandlerFunc {
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

		// Delete menu item from database
		result := collection.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: docID}})

		var deletedItem MenuItem
		err = result.Decode(&deletedItem)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		_ = os.Remove(deletedItem.Img)

		c.JSON(http.StatusOK, nil)
	}
}

// @Summary Get the image of a menu item
// @Description Retrieves the image of a menu item by filename. This route is publicly accessible.
// @Tags menu
// @Param filename path string true "Filename of the image"
// @Success 200 {file} File "Image file"
// @Failure 404  "Image not found"
// @Failure 500 "Internal Server Error"
// @Router /menu/images/{filename} [get]
func GetMenuItemImage(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	filePath := "./uploads/" + filename
	log.Println(os.Stat(filePath))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
		})
		return
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
