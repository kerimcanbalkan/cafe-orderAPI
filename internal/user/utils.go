package user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
)

func ValidateUser(v *validator.Validate, user User) error {
	// Perform validation
	if err := v.Struct(user); err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return nil
		}

		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				return fmt.Errorf("%s is required", fieldErr.Field())
			case "min":
				return fmt.Errorf(
					"%s must be at least %s characters",
					fieldErr.Field(),
					fieldErr.Param(),
				)
			case "max":
				return fmt.Errorf(
					"%s must be at most %s characters",
					fieldErr.Field(),
					fieldErr.Param(),
				)
			case "oneof":
				if fieldErr.Field() == "Gender" {
					return fmt.Errorf("%s must be male or female", fieldErr.Field())
				} else if fieldErr.Field() == "Role" {
					return fmt.Errorf("%s should be one of the following [admin, waiter, cashier]", fieldErr.Field())
				}
			case "email":
				return fmt.Errorf("%s must be a valid email", fieldErr.Field())
			default:
				return fmt.Errorf("%s is invalid", fieldErr.Field())
			}
		}
	}
	return nil
}

// SeedAdminUser ensures that an admin user exists in the database.
// If an admin user is not found, it creates one with credentials from environment variables.
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

func AuthSameUserOrAdmin(
	requestID primitive.ObjectID,
	c *gin.Context,
) bool {
	// Get claims from Gin context
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return false
	}

	// Type assert to jwt.MapClaims
	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token data"})
		return false
	}

	// Extract UserID
	userIDHex, ok := jwtClaims["UserID"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return false
	}

	// Convert the string back to primitive.ObjectID
	clientID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
		return false
	}

	// Extract role from jwt
	role, ok := jwtClaims["Role"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Role"})
		return false
	}

	if role == "admin" {
		return true
	}

	return clientID == requestID
}
