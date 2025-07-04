package order

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kerimcanbalkan/cafe-orderAPI/config"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/db"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/menu"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/sse"
	"github.com/kerimcanbalkan/cafe-orderAPI/internal/utils"
)

var validate = validator.New()

type orderRequest struct {
	Items []OrderItem `json:"items" validate:"required"`
}

// CreateOrder creates an order and saves it in the database
//
// @Summary Create a new order
// @Description Creates a new order for a specific table and saves it in the database
// @Tags order
// @Param table path int true "Table number"
// @Param order body orderRequest true "Order details"
// @Success 200 {object} map[string]interface{} "Order created successfully"
// @Failure 400  "Invalid request"
// @Failure 500  "Internal Server Error"
// @Router /order/{table} [post]
func CreateOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		table := c.Param("tableID")

		if table == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Parameter",
			})
			return
		}

		tableID, err := primitive.ObjectIDFromHex(table)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Table ID",
			})
			return
		}

		ok, err := checkTable(tableID, c, client)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		if ok != true {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Table not found, Please provide a valid Table ID",
			})
			return
		}

		var request orderRequest

		// Bind the request body to the order struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		if len(request.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order should include items",
			})
			return
		}

		totalPrice := int64(0)

		// Validate Items
		for _, orderItem := range request.Items {
			if err := menu.ValidateMenu(validate, orderItem.MenuItem); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"Validation failed for item %s: %s",
						orderItem.MenuItem.Name,
						err.Error(),
					),
				})
				return
			}
			totalPrice += orderItem.MenuItem.Price * int64(orderItem.Quantity)
		}

		// Validate the struct
		if err := validateOrder(validate, request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order := &Order{}

		order.Items = request.Items
		order.TableID = tableID
		order.ClosedAt = nil
		order.CreatedAt = time.Now()
		order.ServedAt = nil
		order.HandledBy = primitive.NilObjectID
		order.ClosedBy = primitive.NilObjectID
		order.TotalPrice = totalPrice

		// Get the collection
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Insert the item into the database
		result, err := collection.InsertOne(ctx, order)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		sse.Notify(table)

		c.JSON(http.StatusOK, gin.H{
			"message": "Order created successfuly",
			"id":      result.InsertedID,
		})
	}
}

// GetOrders retrieves all orders
//
// @Summary Get all orders
// @Description Retrieves all orders for admin, cashier, and waiter roles
// @Tags order
// @Security bearerToken
// @Param is_closed query boolean false "Filter by closed status (true/false)"
// @Param served query boolean false "Filter by served status (true/false)"
// @Param table query int false "Filter by table number"
// @Param date query string false "Filter by order date (YYYY-MM-DD)"
// @Param page query int false "Page number (default is 1)"
// @Param limit query int false "Number of items per page (default is 20)"
// @Success 200 {array} Order "List of orders with their IDs"
// @Failure 500 "Internal Server Error"
// @Router /order [get]
func GetOrders(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []Order

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Get query parameters
		isClosed := c.Query("is_closed")
		served := c.Query("served")
		table := c.Query("table")
		date := c.Query("date")
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "20")

		// Build MongoDB query dynamically
		query := bson.D{}

		page, err := strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid page number."},
			)
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid limit number."},
			)
			return
		}


		// Parse "status" query parameter (convert to boolean)
		if isClosed != "" {
			bool, err := strconv.ParseBool(isClosed)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "closed_at", Value: bson.M{"$exists": bool}})
		}

		// Parse "served" query parameter
		if served != "" {
			bool, err := strconv.ParseBool(served)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid status value. Use true or false."},
				)
				return
			}
			query = append(query, bson.E{Key: "served_at", Value: bson.M{"$exists": bool}})
		}

		// Add table filter
		if table != "" {
			tableID, err := primitive.ObjectIDFromHex(table)
			log.Println("Parsed table ID from query:", tableID.Hex())

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid Table ID",
				})
				return
			}
			query = append(query, bson.E{Key: "table_id", Value: tableID})
		}

		// Parse "date" query parameter
		if date != "" {
			parsedDate, err := time.Parse("2006-01-02", date)
			if err != nil {
				c.JSON(
					http.StatusBadRequest,
					gin.H{"error": "Invalid date format. Use YYYY-MM-DD."},
				)
				return
			}

			// Calculate start and end of the day
			startOfDay := primitive.NewDateTimeFromTime(parsedDate)
			endOfDay := primitive.NewDateTimeFromTime(parsedDate.Add(24 * time.Hour))
			query = append(query, bson.E{
				Key: "created_at",
				Value: bson.D{
					{Key: "$gte", Value: startOfDay},
					{Key: "$lt", Value: endOfDay},
				},
			})
		}

		skip := (page - 1) * limit
		findOptions := options.Find()
		findOptions.SetSkip(int64(skip))
		findOptions.SetLimit(int64(limit))
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})


		// Find all documents in the menu collection
		cursor, err := collection.Find(ctx, query, findOptions)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the menu slice
		if err := cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		totalCount, err := collection.CountDocuments(ctx, query)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		// Return the orders in the response
		c.JSON(http.StatusOK, gin.H{
			"data":  orders,
			"meta": gin.H{
				"total":      totalCount,
				"page":       page,
				"limit":      limit,
				"totalPages": int(math.Ceil(float64(totalCount) / float64(limit))),
			},
		})

	}
}

// ServeOrder sets the servedAt and HandledBy fields.
// @Summary Mark an order as served
// @Description Allows admin and waiter roles to mark an order as served
// @Tags order
// @Param id path string true "Order ID"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order served successfully"
// @Failure 404 "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/serve/{id} [patch]
func ServeOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		// Convert idParam to ObjectID
		id, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Order ID!",
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
		userID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
			return
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")
		ctx := c.Request.Context()

		// Filters by id and checks if its not already been served
		filter := bson.D{
			{Key: "_id", Value: id},
			{Key: "served_at", Value: bson.M{"$exists": false}},
		}

		// Prepare the update statement
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "served_at", Value: time.Now()},
				{Key: "handled_by", Value: userID},
			}},
		}

		// Find order and if exists update
		result := collection.FindOneAndUpdate(ctx, filter, update)

		var order Order
		err = result.Decode(&order)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(
					http.StatusNotFound,
					gin.H{"error": "Wrong order ID or Order already served."},
				)
				return
			}
			utils.HandleMongoError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order served successfully"})
	}
}

// CloseOrder marks an order as complete
//
// @Summary Mark an order as complete
// @Description Allows admin and cashier roles to marks all orders complete for a given table ID
// @Tags order
// @Param id path string true "Table ID"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order completed successfully"
// @Failure 404  "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/close/{tableID} [patch]
func CloseOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("tableID")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		// Convert parameter ID
		id, _ := primitive.ObjectIDFromHex(idParam)

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
		userID, err := primitive.ObjectIDFromHex(userIDHex)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid ObjectID"})
			return
		}

		// Filters by id and checks if its served
		filter := bson.D{
			{Key: "table_id", Value: id},
			{Key: "served_at", Value: bson.M{"$exists": true}},
			{Key: "closed_at", Value: bson.M{"$exists": false}},
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "closed_at", Value: time.Now()},
				{Key: "closed_by", Value: userID},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		result, err := collection.UpdateMany(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		// Check if a document was actually updated
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or must be served first"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Order closed succesfully",
		})
	}
}

// UpdateOrder updates an existing order
//
// @Summary Update an existing order
// @Description Allows admin, cashier, and waiter roles to update an order
// @Tags order
// @Param id path string true "Order ID"
// @Param order body orderRequest true "Order update details"
// @Security bearerToken
// @Success 200 {object} map[string]interface{} "Order updated successfully"
// @Failure 400  "Invalid request"
// @Failure 404  "Order not found"
// @Failure 500  "Internal Server Error"
// @Router /order/{id} [patch]
func UpdateOrder(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID!",
			})
			return
		}

		id, _ := primitive.ObjectIDFromHex(idParam)
		filter := bson.D{{Key: "_id", Value: id}}

		var request orderRequest

		// Bind the request body to the order struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
			})
			return
		}

		totalPrice := int64(0)

		// Validate Items
		for _, orderItem := range request.Items {
			if err := menu.ValidateMenu(validate, orderItem.MenuItem); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf(
						"Validation failed for item %s: %s",
						orderItem.MenuItem.Name,
						err.Error(),
					),
				})
				return
			}
			totalPrice += orderItem.MenuItem.Price * int64(orderItem.Quantity)
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "items", Value: request.Items},
				{Key: "total_price", Value: totalPrice},
			}},
		}

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Updates the first document that has the specified "_id" value
		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}

		// Check if a document was actually updated
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Order updated succesfully",
		})
	}
}

// GetStatistics calculates and serves order statistics for a given date range
// @Summary Get statistics for a given date range.
// @Description Fetches statistics for a specific date range.
// @Tags Statistics
// @Security bearerToken
// @Accept json
// @Produce json
// @Param from query string false "The date for which to fetch the statistics (format: yyyy-mm-dd).
// @Param to query string false "The date for which to fetch the statistics (format: yyyy-mm-dd).
// @Param group_by query string false "Grouping interval: one of 'day', 'week', or 'month'"
// @Success 200 {object} map[string]interface{} "Order statistics data"
// @Failure 400 {object} map[string]string "Invalid date format"
// @Failure 500 {object} map[string]string "Failed to fetch statistics"
// @Router /order/stats [get]
func GetStatistics(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := client.GetCollection(config.Env.DatabaseName, "orders")
		// Get date range from query params
		
		fromStr := c.Query("from")
		toStr := c.Query("to")
		groupBy := c.Query("group_by")

		if groupBy != "day" && groupBy != "week" && groupBy != "month" {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid 'group_by' parameter. Allowed values are 'day', 'week', or 'month'."},
			)
			return
		}

		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "Invalid 'from' date format use YYYY-MM-DD"},
			)
			return
		}

		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' date format YYYY-MM-DD"})
			return
		}

		stats, err := getStats(c, collection, from, to, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
			return
		}

		// Return the stats in the response
		c.JSON(http.StatusOK, gin.H{
			"data": stats,
		})
	}
}

// GetActiveOrdersByTableID gets active orders for a specific table
//
// @Summary Get table specific active orders
// @Description Retrieves all active (not signed as closed) orders for a table
// @Tags order
// @Param tableID path string true "Table ID"
// @Success 200 {array} Order "List of orders with their IDs"
// @Failure 500 "Internal Server Error"
// @Router /order/:tableID [get]
func GetActiveOrdersByTableID(client db.IMongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []Order

		id := c.Param("tableID")
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

		// Get the collection from the database
		collection := client.GetCollection(config.Env.DatabaseName, "orders")

		// Get context from the request
		ctx := c.Request.Context()

		// Build the query
		query := bson.D{}

		// Check for closed_at status
		query = append(query, bson.E{Key: "closed_at", Value: bson.M{"$exists": false}})

		// Match the table ID
		query = append(query, bson.E{Key: "table_id", Value: docID})

		// Find all documents that fits the query
		cursor, err := collection.Find(ctx, query)
		if err != nil {
			utils.HandleMongoError(c, err)
			return
		}
		defer cursor.Close(ctx)

		// Decode the results into the order slice
		if err := cursor.All(ctx, &orders); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse database response.",
			})
			return
		}

		var total OrderTotal
		total.TableID = docID
		total.Items = []OrderItem{}
		total.TotalPrice = 0

		itemIndexMap := make(map[primitive.ObjectID]int)

		for _, order := range orders {
			for _, item := range order.Items {
				menuItemID := item.MenuItem.ID
				if idx, exists := itemIndexMap[menuItemID]; exists {
					total.Items[idx].Quantity += item.Quantity
				} else {
					total.Items = append(total.Items, item)
					itemIndexMap[menuItemID] = len(total.Items) - 1
				}
			}
		}

		// Calculate total price
		var totalPrice int64
		for _, item := range total.Items {
			totalPrice += item.MenuItem.Price * int64(item.Quantity)
		}
		total.TotalPrice = totalPrice

		// Return the orders in the response
		c.JSON(http.StatusOK, gin.H{
			"data": total,
		})
	}
}
