package user

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

// For excluding sensetive data
type userResponse struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name"          json:"name"      validate:"required,min=3,max=20"`
	Surname   string             `bson:"surname"       json:"surname"   validate:"required,min=3,max=20"`
	Gender    string             `bson:"gender"        json:"gender"    validate:"required,oneof=male female"`
	Email     string             `bson:"email"         json:"email"     validate:"required,email"`
	Username  string             `bson:"username"      json:"username"  validate:"required,min=3,max=20"`
	Role      string             `bson:"role"          json:"role"      validate:"required,oneof=admin cashier waiter"`
	CreatedAt time.Time          `bson:"created_at"    json:"createdAt"`
}

// CreateUser creates a new user and saves it to the database
//
// @Summary Create a new user
// @Description Allows admin role to create a new user
// @Tags user
// @Param user body User true "User details"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "User created successfully"
// @Failure 400 "Invalid request"
// @Failure 500 "Internal Server Error"
// @Router /user [post]
func CreateUser(client db.IMongoClient) gin.HandlerFunc {
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
			if mongo.IsDuplicateKeyError(err) {
				c.JSON(
					http.StatusConflict,
					gin.H{
						"error": "Username or email already exists. Please choose a different one.",
					},
				)
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User created successfuly",
			"id":      result.InsertedID,
		})
	}
}

// GetUsers retrieves all users from the database
//
// @Summary Retrieve all users
// @Description Allows admin role to retrieve a list of all users
// @Tags user
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 500  "Internal Server Error"
// @Router /user [get]
func GetUsers(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []userResponse

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, bson.D{})
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found.",
				})
			}
			utils.HandleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		// Return the user in the response
		c.JSON(http.StatusOK, gin.H{
			"data": users,
		})
	}
}

type LoginBody struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

// Login authenticates a user and returns a JWT token
//
// @Summary User login
// @Description Allows users to log in by providing username and password
// @Tags user
// @Param loginBody body LoginBody true "Login details"
// @Success 200 {object} map[string]interface{} "JWT token and expiration time"
// @Failure 400  "Invalid request"
// @Failure 401 "Unauthorized"
// @Failure 500  "Internal Server Error"
// @Router /user/login [post]
func Login(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body LoginBody
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return

		}
		var user User
		// Get the collection
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		filter := bson.D{{Key: "username", Value: body.Username}}

		err := collection.FindOne(ctx, filter).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found."})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Wrong password or username!",
			})
			return
		}

		claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"UserID":    user.ID,
			"Role":      user.Role,
			"ExpiresAt": time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		// Sign and get the complete encoded token as a string using the secret
		token, err := claims.SignedString([]byte(config.Env.Secret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not login",
			})
			return

		}
		// Return the token in a JSON response
		c.JSON(http.StatusOK, gin.H{
			"token":      token,
			"expires_in": 30 * 24 * 60 * 60, // 30 days
		})
	}
}

// DeleteUser deletes a user by their ID
//
// @Summary Delete a user
// @Description Allows admin role to delete a user by their ID
// @Tags user
// @Param id path string true "User ID"
// @Security bearerToken
// @Success 200 {object} nil "User deleted successfully"
// @Failure 400  "Invalid ID"
// @Failure 404  "User not found"
// @Failure 500  "Internal Server Error"
// @Router /user/{id} [delete]
func DeleteUser(client db.IMongoClient) gin.HandlerFunc {
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
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Delete user from database
		result := collection.FindOneAndDelete(ctx, bson.D{{Key: "_id", Value: docID}})

		var deletedUser User
		err = result.Decode(&deletedUser)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}

func GetUserById(client db.IMongoClient) gin.HandlerFunc {
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

		// Get claims from Gin context
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Type assert to jwt.MapClaims
		jwtClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token data"})
			return
		}

		// Extract UserID
		userIDHex, ok := jwtClaims["UserID"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		// Convert the string back to primitive.ObjectID
		clientID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
			return
		}

		// Extract role from jwt
		role, ok := jwtClaims["Role"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Role"})
			return
		}

		// Get collection from db
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Delete user from database
		result := collection.FindOne(ctx, bson.D{{Key: "_id", Value: docID}})

		var user userResponse
		err = result.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		if clientID != user.ID && role != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		// Return the user in the response
		c.JSON(http.StatusOK, gin.H{
			"data": user,
		})
	}
}

func GetYearlyStatistics(client db.IMongoClient) gin.HandlerFunc {
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

		// Get claims from Gin context
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Type assert to jwt.MapClaims
		jwtClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token data"})
			return
		}

		// Extract UserID
		userIDHex, ok := jwtClaims["UserID"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}

		// Convert the string back to primitive.ObjectID
		clientID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
			return
		}

		// Extract role from jwt
		role, ok := jwtClaims["Role"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Role"})
			return
		}

		// Get collection from db
		collection := client.GetCollection(config.Env.DatabaseName, "users")

		// Get context from the request
		ctx := c.Request.Context()

		// Delete user from database
		result := collection.FindOne(ctx, bson.D{{Key: "_id", Value: docID}})

		var user userResponse
		err = result.Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		if clientID != user.ID && role != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		yearStr := c.DefaultQuery("year", time.Now().Format("2006"))

		yearInt, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year format"})
			return
		}

		// Set the start date to January 1st of the given year
		startDate := time.Date(yearInt, time.January, 1, 0, 0, 0, 0, time.UTC)

		// Set the end date to December 31st of the given year at 23:59:59
		endDate := time.Date(yearInt, time.December, 31, 23, 59, 59, 999999999, time.UTC)

		collection = client.GetCollection(config.Env.DatabaseName, "orders")

		var stats interface{}

		if user.Role == "waiter" {
			log.Println(
				"Fetching waiter stats for:",
				user.ID,
			) // Add this log to check execution flow
			stats, err = getWaiterStats(startDate, endDate, c, collection, user.ID)
			log.Println("THIS IS THE STATS RETURN", stats)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
				return
			}
		} else if user.Role == "cashier" {
			log.Println(
				"Fetching cashier stats for:",
				user.ID,
			) // Add this log to check execution flow

			stats, err = getCashierStats(startDate, endDate, c, collection, user.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
				return
			}

		}

		// Return the stats in the response
		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}
