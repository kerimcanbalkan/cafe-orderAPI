package user

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WaiterStats struct {
	TotalOrdersServed  int     `bson:"total_orders_served"  json:"totalOrdersServed"`
	AverageServingTime float64 `bson:"average_serving_time" json:"averageServingTime"` // minutes
	FastestServingTime int     `bson:"fastest_serving_time" json:"fastestServingTime"` // minutes
}

type CashierStats struct {
	TotalOrdersClosed int     `bson:"total_orders_closed" json:"totalOrdersClosed"`
	TotalRevenue      float64 `bson:"total_revenue"       json:"totalRevenue"`
}

// getStats calculates waiter statistics for a given date range.
// Needs orders collection NOT users collection
func getWaiterStats(
	startDate, endDate time.Time,
	ctx context.Context,
	collection *mongo.Collection,
	userID primitive.ObjectID,
) (WaiterStats, error) {
	matchFilter := bson.M{
		"served_at": bson.M{"$gte": startDate, "$lt": endDate},
		"handled_by": userID,
	}

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: matchFilter}, // Match filter for waiterID and date
		},
		{
			{Key: "$addFields", Value: bson.M{
				"servingTime": bson.M{
					"$subtract": []interface{}{
						"$served_at",
						"$created_at",
					}, // Get the difference between servedAt and createdAt
				},
			}},
		},
		{
			{Key: "$group", Value: bson.M{
				"_id":                 nil,
				"total_orders_served": bson.M{"$sum": 1}, // Count total orders served
				"average_serving_time": bson.M{
					"$avg": "$servingTime",
				}, // Calculate average serving time
				"fastest_serving_time": bson.M{
					"$min": "$servingTime",
				}, // Find the fastest serving time
			}},
		},
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		log.Println("THIS IS THE ERROR", err)
		return WaiterStats{}, err
	}
	defer cursor.Close(ctx)

	// Decode result
	var stats []WaiterStats
	if err := cursor.All(ctx, &stats); err != nil || len(stats) == 0 {
		log.Println("THIS IS THE ERROR", err)
		return WaiterStats{}, err
	}

	stats[0].AverageServingTime = stats[0].AverageServingTime / 60000.0
	stats[0].FastestServingTime = int(stats[0].FastestServingTime / 60000.0)

	return stats[0], nil
}

// getCashierStats calculates cashier statistics for a given date range.
// Needs orders collection NOT users collection
func getCashierStats(
	startDate, endDate time.Time,
	ctx context.Context,
	collection *mongo.Collection,
	userID primitive.ObjectID,
) (CashierStats, error) {
	matchFilter := bson.M{
		"closed_at": bson.M{"$gte": startDate, "$lt": endDate},
		"closed_by":  userID,
	}
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: matchFilter}, // Match filter for cashierID
		},
		{
			{Key: "$group", Value: bson.M{
				"_id":                 nil,               // No grouping key, aggregate over all documents
				"total_orders_closed": bson.M{"$sum": 1}, // Count total orders closed
				"total_revenue": bson.M{
					"$sum": "$total_price",
				}, // Sum of the `totalPrice` field for total revenue
			}},
		},
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return CashierStats{}, err
	}
	defer cursor.Close(ctx)

	// Decode result
	var stats []CashierStats
	if err := cursor.All(ctx, &stats); err != nil || len(stats) == 0 {
		return CashierStats{}, err
	}

	return stats[0], nil
}
