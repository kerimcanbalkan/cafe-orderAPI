package menu_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
)

type MenuResponse struct {
	Data []menu.MenuItem `json:"data"`
}

type MenuErrorResponse struct {
	Error string `json:"error"`
}

func TestGetMenu(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		first := mtest.CreateCursorResponse(1, "testDB.menu", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Burger"},
			{Key: "description", Value: "A delicious beef burger"},
			{Key: "price", Value: 5.99},
			{Key: "category", Value: "Main"},
			{Key: "image", Value: "burger_image_url"},
		})
		second := mtest.CreateCursorResponse(1, "testDB.menu", mtest.NextBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Pizza"},
			{Key: "description", Value: "Cheese and tomato pizza"},
			{Key: "price", Value: 8.99},
			{Key: "category", Value: "Main"},
			{Key: "image", Value: "pizza_image_url"},
		})

		// Simulate cursor close
		killCursors := mtest.CreateCursorResponse(0, "testDB.menu", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// test route
		r := gin.Default()
		r.GET("/test/menu", menu.GetMenu(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/menu", nil)
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var menuResponse MenuResponse
		err := json.Unmarshal(w.Body.Bytes(), &menuResponse)
		assert.Nil(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Len(t, menuResponse.Data, 2) // Two items expected
		assert.Equal(t, "Burger", menuResponse.Data[0].Name)
		assert.Equal(t, "Pizza", menuResponse.Data[1].Name)
	})
}
