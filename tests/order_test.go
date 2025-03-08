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
			Items: []order.OrderItem{{
				MenuItem: menu.MenuItem{
					ID:          primitive.NewObjectID(),
					Name:        "Pizza",
					Description: "Cheese pizza",
					Price:       10.99,
					Category:    "Food",
					Img:         "pizza.jpg",
				},
				Quantity: 2,
			}},
			TotalPrice:  10.99,
			TableNumber: uint8(5),
			ClosedAt:    nil,
			ServedAt:    nil,
			CreatedAt:   time.Now(),
			HandledBy:   primitive.NewObjectID(),
			ClosedBy:    primitive.NewObjectID(),
		}

		order2 := order.Order{
			ID: primitive.NewObjectID(),
			Items: []order.OrderItem{{
				MenuItem: menu.MenuItem{
					ID:          primitive.NewObjectID(),
					Name:        "Pasta",
					Description: "Spaghetti with tomato sauce",
					Price:       12.50,
					Category:    "Food",
					Img:         "pasta.jpg",
				},
				Quantity: 1,
			}},
			TotalPrice:  12.50,
			TableNumber: uint8(8),
			ClosedAt:    nil,
			ServedAt:    nil,
			CreatedAt:   time.Now(),
			HandledBy:   primitive.NewObjectID(),
			ClosedBy:    primitive.NewObjectID(),
		}

		// Create the first cursor response with key-value BSON pairs
		first := mtest.CreateCursorResponse(1, "testDB.orders", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: order1.ID},
			{Key: "items", Value: order1.Items},
			{Key: "total_price", Value: order1.TotalPrice},
			{Key: "table_number", Value: order1.TableNumber},
			{Key: "closed_at", Value: order1.ClosedAt},
			{Key: "served_at", Value: order1.ServedAt},
			{Key: "created_at", Value: order1.CreatedAt},
			{Key: "handled_ay", Value: order1.HandledBy},
			{Key: "closed_by", Value: order1.ClosedBy},
		})

		// Create the second cursor response with key-value BSON pairs
		second := mtest.CreateCursorResponse(1, "testDB.orders", mtest.NextBatch, bson.D{
			{Key: "_id", Value: order2.ID},
			{Key: "items", Value: order2.Items},
			{Key: "total_price", Value: order2.TotalPrice},
			{Key: "table_number", Value: order2.TableNumber},
			{Key: "closed_at", Value: order2.ClosedAt},
			{Key: "served_at", Value: order2.ServedAt},
			{Key: "created_at", Value: order2.CreatedAt},
			{Key: "handled_by", Value: order2.HandledBy},
			{Key: "closed_by", Value: order2.ClosedBy},
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
		assert.Equal(t, uint8(8), orderResponse.Data[1].TableNumber)
		assert.Equal(t, "Pasta", orderResponse.Data[1].Items[0].MenuItem.Name)
	})
}

func TestCreateOrder(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	o := order.Order{
		Items: []order.OrderItem{{
			MenuItem: menu.MenuItem{
				ID:          primitive.NewObjectID(),
				Name:        "Pizza",
				Description: "Cheese pizza",
				Price:       10.99,
				Category:    "Food",
				Img:         "pizza.jpg",
			},
			Quantity: 3,
		},
		}}

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
			Items: []order.OrderItem{},
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
	o := order.Order{
		Items: []order.OrderItem{{
			MenuItem: menu.MenuItem{
				ID:          primitive.NewObjectID(),
				Name:        "Pizza",
				Description: "Cheese pizza",
				Price:       10.99,
				Category:    "Food",
				Img:         "pizza.jpg",
			},
			Quantity: 3,
		},
		}}
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
		id := o.Items[0].MenuItem.ID.Hex()

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
		id := o.Items[0].MenuItem.ID.Hex()

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
