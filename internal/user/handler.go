package user

import (
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
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

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
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User created successfuly",
			"user_id": result.InsertedID,
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
			utils.HandleMongoError(c, err)
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

type loginBody struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

// Login authenticates a user and returns a JWT token
//
// @Summary User login
// @Description Allows users to log in by providing username and password
// @Tags user
// @Param loginBody body loginBody true "Login details"
// @Success 200 {object} map[string]interface{} "JWT token and expiration time"
// @Failure 400  "Invalid request"
// @Failure 401 "Unauthorized"
// @Failure 500  "Internal Server Error"
// @Router /user/login [post]
func Login(client *db.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body loginBody
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
func DeleteUser(client *db.MongoClient) gin.HandlerFunc {
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
		_, err = collection.DeleteOne(ctx, bson.M{"_id": docID})
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, nil)
	}
}
