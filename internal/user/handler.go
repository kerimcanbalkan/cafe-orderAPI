package user

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
)

var validate = validator.New()

func CreateUser(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return

		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Password",
			})
			return
		}

		if err = ValidateUser(validate, user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}

		user.Password = string(hash)
		user.CreatedAt = time.Now()

		// Get the collection
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Insert the item into the database
		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User created successfuly",
			"user_id": result.InsertedID,
		})
	}
}

func GetUsers(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []User

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}

		// Return the menu in the response
		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	}
}

func SeedAdminUser(client *db.MongoClient, ctx context.Context) {
	collection := client.GetCollection(config.Env.DatabaseName, "users")

	// Check if an admin user already exists
	var existingAdmin User
	err := collection.FindOne(ctx, bson.M{"role": "admin"}).Decode(&existingAdmin)
	if err == nil {
		return
	}

	// Get admin credentials from environment variables
	adminUsername := config.Env.DefaultAdminUsername
	adminPassword := config.Env.DefaultAdminPassword
	if adminUsername == "" || adminPassword == "" {
		panic("ADMIN_USERNAME and ADMIN_PASSWORD must be set in environment variables")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), 10)
	admin := User{
		ID:        primitive.NewObjectID(),
		Name:      "Default",
		Surname:   "Admin",
		Username:  adminUsername,
		Email:     "admin@example.com",
		Password:  string(hash),
		Role:      "admin",
		CreatedAt: time.Now(),
	}

	_, err = collection.InsertOne(ctx, admin)
	if err != nil {
		panic("Failed to seed admin user: " + err.Error())
	}
	log.Println("Admin user created successfully!")
}
