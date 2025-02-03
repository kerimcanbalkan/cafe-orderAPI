package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/user"
)

type UsersResponse struct {
	Data []user.User `json:"data"`
}

type UserErrorResponse struct {
	Error string `json:"error"`
}

func TestGetUsers(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create sample user instances
		user1 := user.User{
			ID:        primitive.NewObjectID(),
			Name:      "Kerim",
			Surname:   "Balkan",
			Gender:    "male",
			Email:     "kerimcanbalkan@gmail.com",
			Username:  "ballkan",
			Password:  "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92",
			Role:      "waiter",
			CreatedAt: time.Now(),
		}

		user2 := user.User{
			ID:        primitive.NewObjectID(),
			Name:      "Eda",
			Surname:   "Balkan",
			Gender:    "female",
			Email:     "edabalkan@gmail.com",
			Username:  "eda",
			Password:  "b79ea17b7c5ca8fe9cccd8cdba6e8f8ed0b3c948f9f709ed0f47d2fd47fcba82",
			Role:      "cashier",
			CreatedAt: time.Now(),
		}

		// Create mock responses for users
		first := mtest.CreateCursorResponse(1, "testDB.users", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: user1.ID.Hex()},
			{Key: "name", Value: user1.Name},
			{Key: "surname", Value: user1.Surname},
			{Key: "gender", Value: user1.Gender},
			{Key: "email", Value: user1.Email},
			{Key: "username", Value: user1.Username},
			{Key: "password", Value: user1.Password},
			{Key: "role", Value: user1.Role},
			{Key: "created_at", Value: user1.CreatedAt},
		})

		second := mtest.CreateCursorResponse(1, "testDB.users", mtest.NextBatch, bson.D{
			{Key: "_id", Value: user2.ID.Hex()},
			{Key: "name", Value: user2.Name},
			{Key: "surname", Value: user2.Surname},
			{Key: "gender", Value: user2.Gender},
			{Key: "email", Value: user2.Email},
			{Key: "username", Value: user2.Username},
			{Key: "password", Value: user2.Password},
			{Key: "role", Value: user2.Role},
			{Key: "created_at", Value: user2.CreatedAt},
		})

		// Simulate cursor close
		killCursors := mtest.CreateCursorResponse(0, "testDB.users", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// Test route
		r := gin.Default()
		r.GET("/test/user", user.GetUsers(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/user", nil)
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var usersResponse UsersResponse
		err := json.Unmarshal(w.Body.Bytes(), &usersResponse)
		assert.Nil(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Len(t, usersResponse.Data, 2)
		assert.Equal(t, "Kerim", usersResponse.Data[0].Name)
		assert.Equal(t, "female", usersResponse.Data[1].Gender)
		assert.Equal(t, "Balkan", usersResponse.Data[1].Surname)
	})
}

type CreateResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

func TestCreateUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	usr := user.User{
		Name:     "Eda",
		Surname:  "Balkan",
		Gender:   "female",
		Email:    "edabalkan@gmail.com",
		Username: "eda",
		Password: "123456321",
		Role:     "cashier",
	}

	mt.Run("success", func(mt *mtest.T) {
		body, _ := json.Marshal(usr)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// Test route
		r := gin.Default()
		r.POST("/test/user", user.CreateUser(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var createResponse CreateResponse
		err := json.Unmarshal(w.Body.Bytes(), &createResponse)

		assert.Nil(t, err)
		assert.Equal(t, "User created successfuly", createResponse.Message)
	})

	mt.Run("custom error duplicate", func(mt *mtest.T) {
		body, _ := json.Marshal(usr)
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// Test route
		r := gin.Default()
		r.POST("/test/user", user.CreateUser(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var userErrorResponse UserErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &userErrorResponse)
		assert.Nil(t, err)
		assert.Equal(t,
			"Username or email already exists. Please choose a different one.",
			userErrorResponse.Error,
		)
	})
}

func TestUserValidation(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("custom error validation", func(mt *mtest.T) {
		mockClient := db.NewMockMongoClient(mt.Coll)

		r := gin.Default()
		r.POST("/test/user", user.CreateUser(mockClient))

		cases := []struct {
			name          string
			input         user.User
			expectedError string
		}{
			{
				"Missing Name",
				user.User{
					Surname:  "Doe",
					Gender:   "male",
					Email:    "john@example.com",
					Username: "johndoe",
					Password: "password123",
					Role:     "admin",
				},
				"Name is required",
			},
			{
				"Name Too Short",
				user.User{
					Name:     "Jo",
					Surname:  "Doe",
					Gender:   "male",
					Email:    "john@example.com",
					Username: "johndoe",
					Password: "password123",
					Role:     "admin",
				},
				"Name must be at least 3 characters",
			},
			{
				"Password Too Short",
				user.User{
					Name:     "John",
					Surname:  "Doe",
					Gender:   "male",
					Email:    "john@example.com",
					Username: "johndoe",
					Password: "123456",
					Role:     "admin",
				},
				"Password must be at least 8 characters",
			},
			{
				"Invalid Email",
				user.User{
					Name:     "John",
					Surname:  "Doe",
					Gender:   "male",
					Email:    "invalid-email",
					Username: "johndoe",
					Password: "password123",
					Role:     "admin",
				},
				"Email must be a valid email",
			},
			{
				"Invalid Gender",
				user.User{
					Name:     "John",
					Surname:  "Doe",
					Gender:   "other",
					Email:    "john@example.com",
					Username: "johndoe",
					Password: "password123",
					Role:     "admin",
				},
				"Gender must be male or female",
			},
			{
				"Invalid Role",
				user.User{
					Name:     "John",
					Surname:  "Doe",
					Gender:   "male",
					Email:    "john@example.com",
					Username: "johndoe",
					Password: "password123",
					Role:     "manager",
				},
				"Role should be one of the following [admin, waiter, cashier]",
			},
		}

		for _, tc := range cases {
			body, _ := json.Marshal(tc.input)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test/user", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			var response UserErrorResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			if tc.expectedError == "" {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Equal(t, tc.expectedError, response.Error)
			}
		}
	})
}

func TestDeleteUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		id := primitive.NewObjectID().Hex()
		// Mock a deleted user response
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: id},
				{Key: "name", Value: "John"},
				{Key: "surname", Value: "Doe"},
				{Key: "gender", Value: "male"},
				{Key: "email", Value: "johndoe@example.com"},
				{Key: "username", Value: "johndoe"},
				{
					Key:   "password",
					Value: "$2a$10$r/hOOqj2CLcVweIvPz23ZOlB8PLRI84pxG4ZqM.fbIYAjqRJcxh82",
				},
				{Key: "role", Value: "waiter"},
				{Key: "created_at", Value: time.Now()},
			}},
		})

		// Test route
		r := gin.Default()
		r.DELETE("/test/user/:id", user.DeleteUser(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/test/user/"+id, nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	mt.Run("custom not found error", func(mt *mtest.T) {
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		id := primitive.NewObjectID().Hex()
		// Mock a deleted user response
		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 0}},
		)

		// Test route
		r := gin.Default()
		r.DELETE("/test/user/:id", user.DeleteUser(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/test/user/"+id, nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestLogin(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		id := primitive.NewObjectID().Hex()

		// Mock a FindOne user response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "testDB.users", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: id},
			{Key: "name", Value: "John"},
			{Key: "surname", Value: "Doe"},
			{Key: "gender", Value: "male"},
			{Key: "email", Value: "johndoe@example.com"},
			{Key: "username", Value: "johndoe"},
			{
				Key:   "password",
				Value: "$2a$10$r/hOOqj2CLcVweIvPz23ZOlB8PLRI84pxG4ZqM.fbIYAjqRJcxh82",
			},
			{Key: "role", Value: "waiter"},
			{Key: "created_at", Value: time.Now()},
		}))

		// Test route
		r := gin.Default()
		r.POST("/test/user/login", user.Login(mockClient))
		loginBody := user.LoginBody{
			Username: "johndoe",
			Password: "password",
		}

		body, _ := json.Marshal(loginBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/user/login", bytes.NewBuffer(body))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Parse the response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert the response contains the expected fields
		assert.Contains(t, response, "expires_in")
		assert.Equal(t, float64(2592000), response["expires_in"])

		assert.Contains(t, response, "token")
		token, ok := response["token"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, token)

		// Check if the token has a valid format (JWT has 3 parts)
		tokenParts := strings.Split(token, ".")
		assert.Len(t, tokenParts, 3, "token should have 3 parts")
	})
}
