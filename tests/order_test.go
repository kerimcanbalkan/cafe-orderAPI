package test

import (
	"bytes"
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
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/order"
)

func TestGetOrders(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		// Create sample Order and MenuItem instances
		order1 := order.Order{
			ID: primitive.NewObjectID(),
			Items: []menu.MenuItem{
				{
					ID:          primitive.NewObjectID(),
					Name:        "Pizza",
					Description: "Cheese pizza",
					Price:       10.99,
					Category:    "Food",
					Img:         "pizza.jpg",
				},
			},
			TotalPrice:  10.99,
			TableNumber: 5,
			ClosedAt:    nil,
			ServedAt:    nil,
			CreatedAt:   time.Now(),
			HandledBy:   primitive.NewObjectID(),
		}

		order2 := order.Order{
			ID: primitive.NewObjectID(),
			Items: []menu.MenuItem{
				{
					ID:          primitive.NewObjectID(),
					Name:        "Pasta",
					Description: "Spaghetti with tomato sauce",
					Price:       12.50,
					Category:    "Food",
					Img:         "pasta.jpg",
				},
			},
			TotalPrice:  12.50,
			TableNumber: 3,
			ClosedAt:    nil,
			ServedAt:    nil,
			CreatedAt:   time.Now(),
			HandledBy:   primitive.NewObjectID(),
		}

		// Create the first cursor response with key-value BSON pairs
		first := mtest.CreateCursorResponse(1, "testDB.orders", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: order1.ID},
			{Key: "items", Value: order1.Items},
			{Key: "totalPrice", Value: order1.TotalPrice},
			{Key: "tableNumber", Value: order1.TableNumber},
			{Key: "closedAt", Value: order1.ClosedAt},
			{Key: "servedAt", Value: order1.ServedAt},
			{Key: "createdAt", Value: order1.CreatedAt},
			{Key: "handledBy", Value: order1.HandledBy},
		})

		// Create the second cursor response with key-value BSON pairs
		second := mtest.CreateCursorResponse(1, "testDB.orders", mtest.NextBatch, bson.D{
			{Key: "_id", Value: order2.ID},
			{Key: "items", Value: order2.Items},
			{Key: "totalPrice", Value: order2.TotalPrice},
			{Key: "tableNumber", Value: order2.TableNumber},
			{Key: "closedAt", Value: order2.ClosedAt},
			{Key: "servedAt", Value: order2.ServedAt},
			{Key: "createdAt", Value: order2.CreatedAt},
			{Key: "handledBy", Value: order2.HandledBy},
		})

		// Simulate cursor close
		killCursors := mtest.CreateCursorResponse(0, "testDB.orders", mtest.NextBatch)

		mt.AddMockResponses(first, second, killCursors)
		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// test route
		r := gin.Default()
		r.GET("/test/order", order.GetOrders(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test/order", nil)
		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var orderResponse OrderResponse
		err := json.Unmarshal(w.Body.Bytes(), &orderResponse)
		assert.Nil(t, err)

		assert.Equal(t, 200, w.Code)
		assert.Len(t, orderResponse.Data, 2)
		assert.Equal(t, 10.99, orderResponse.Data[0].TotalPrice)
		assert.Equal(t, 3, orderResponse.Data[1].TableNumber)
		assert.Equal(t, "Pasta", orderResponse.Data[1].Items[0].Name)
	})
}

func TestCreateOrder(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	o := order.Order{
		Items: []menu.MenuItem{
			{
				ID:          primitive.NewObjectID(),
				Name:        "Pizza",
				Description: "Cheese pizza",
				Price:       10.99,
				Category:    "Food",
				Img:         "pizza.jpg",
			},
		},
	}

	mt.Run("success", func(mt *mtest.T) {
		body, _ := json.Marshal(o)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Create mock client
		mockClient := db.NewMockMongoClient(mt.Coll)

		// Test route
		r := gin.Default()
		r.POST("/test/order/:table", order.CreateOrder(mockClient))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/order/5", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Unmarshal into the wrapper struct
		var createResponse CreateResponse
		err := json.Unmarshal(w.Body.Bytes(), &createResponse)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Order created successfuly", createResponse.Message)
	})
}

func TestOrderValidation(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("custom error validation", func(mt *mtest.T) {
		o := order.Order{
			Items: []menu.MenuItem{},
		}
		mockClient := db.NewMockMongoClient(mt.Coll)

		mt.Log("item lenght", len(o.Items))

		r := gin.Default()
		r.POST("/test/order/:table", order.CreateOrder(mockClient))

		body, _ := json.Marshal(o)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test/order/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		mt.Log("ACTUAL RESPONSE", w.Body.String())

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "Order should include items", response.Error)
	})
}

func TestUpdateOrder(t *testing.T) {
	o := []menu.MenuItem{
		{
			ID:          primitive.NewObjectID(),
			Name:        "Pizza",
			Description: "Cheese pizza",
			Price:       10.99,
			Category:    "Food",
			Img:         "pizza.jpg",
		},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		mockClient := db.NewMockMongoClient(mt.Coll)

		mt.AddMockResponses(
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "acknowledged", Value: true},
				{Key: "n", Value: 1},
			},
		)

		r := gin.Default()
		r.PATCH("/test/order/:id", order.UpdateOrder(mockClient))
		id := o[0].ID.Hex()

		body, _ := json.Marshal(o)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/test/order/"+id, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		mt.Log("ACTUAL RESPONSE", w.Body.String())

		var response SuccessResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Order updated succesfully", response.Message)
	})

	mt.Run("custom error not found", func(mt *mtest.T) {
		mockClient := db.NewMockMongoClient(mt.Coll)

		mt.AddMockResponses(
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "acknowledged", Value: true},
				{Key: "n", Value: int32(0)},
			},
		)

		r := gin.Default()
		r.PATCH("/test/order/:id", order.UpdateOrder(mockClient))
		id := o[0].ID.Hex()

		body, _ := json.Marshal(o)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PATCH", "/test/order/"+id, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)
		mt.Log("ACTUAL RESPONSE", w.Body.String())

		var response ErrorResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "Order not found.", response.Error)
	})
}
