package order

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderStats represents aggregated order statistics for a given time period.
type OrderStats struct {
	TotalOrders       int     `bson:"total_orders"        json:"totalOrders"`
	TotalRevenue      float64 `bson:"total_revenue"       json:"totalRevenue"`
	AverageOrderValue float64 `bson:"average_order_value" json:"averageOrderValue"`
}

// getStats calculates order statistics for a given date range.
func getStats(
	startDate, endDate time.Time,
	ctx context.Context,
	collection *mongo.Collection,
) (OrderStats, error) {
	matchFilter := bson.M{
		"created_at": bson.M{"$gte": startDate, "$lt": endDate},
		"closed_at":  bson.M{"$exists": true},
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{
			{Key: "$group", Value: bson.M{
				"_id":                 nil,
				"total_orders":        bson.M{"$sum": 1},
				"total_revenue":       bson.M{"$sum": "$total_price"},
				"average_order_value": bson.M{"$avg": "$total_price"},
			}},
		},
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		return OrderStats{}, err
	}
	defer cursor.Close(ctx)

	// Decode result
	var stats []OrderStats
	if err := cursor.All(ctx, &stats); err != nil || len(stats) == 0 {
		return OrderStats{}, err
	}

	return stats[0], nil
}
