package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

		// Log response body for debugging
		mt.Log("response", w.Body.String())

		// Unmarshal into the wrapper struct
		var usersResponse UsersResponse
		err := json.Unmarshal(w.Body.Bytes(), &usersResponse)
		if err != nil {
			mt.Log("Unmarshal error: ", err)
		}
		assert.Nil(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Len(t, usersResponse.Data, 2)
		assert.Equal(t, "Kerim", usersResponse.Data[0].Name)
		assert.Equal(t, "female", usersResponse.Data[1].Gender)
		assert.Equal(t, "Balkan", usersResponse.Data[1].Surname)
	})
}
