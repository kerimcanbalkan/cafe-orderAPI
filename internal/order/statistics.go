package order

import (
	"context"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AggregatedStat struct {
	GroupKey          string  `bson:"group_key" json:"groupKey"` // Will be day/week/month as a string
	TotalOrders       int     `bson:"total_orders" json:"totalOrders"`
	TotalRevenue      float64 `bson:"total_revenue" json:"totalRevenue"`
	AverageOrderValue float64 `bson:"average_order_value" json:"averageOrderValue"`
}

func getStats(
	ctx context.Context,
	collection *mongo.Collection,
	from time.Time,
	to time.Time,
	groupBy string, // "day", "week", or "month"
) ([]AggregatedStat, error) {
	matchFilter := bson.M{
		"created_at": bson.M{"$gte": from, "$lt": to},
		"closed_at":  bson.M{"$exists": true},
	}

	var groupID bson.M
	var groupKeyExpr bson.M

	switch groupBy {
	case "day":
		groupID = bson.M{
			"year":  bson.M{"$year": "$created_at"},
			"month": bson.M{"$month": "$created_at"},
			"day":   bson.M{"$dayOfMonth": "$created_at"},
		}
		groupKeyExpr = bson.M{
			"$dateToString": bson.M{
				"format": "%Y-%m-%d",
				"date":   "$created_at",
			},
		}
	case "week":
		groupID = bson.M{
			"year": bson.M{"$isoWeekYear": "$created_at"},
			"week": bson.M{"$isoWeek": "$created_at"},
		}
		groupKeyExpr = bson.M{
			"$concat": []interface{}{
				bson.M{"$toString": bson.M{"$isoWeekYear": "$created_at"}},
				"-W",
				bson.M{"$toString": bson.M{"$isoWeek": "$created_at"}},
			},
		}
	case "month":
		groupID = bson.M{
			"year":  bson.M{"$year": "$created_at"},
			"month": bson.M{"$month": "$created_at"},
		}
		groupKeyExpr = bson.M{
			"$dateToString": bson.M{
				"format": "%Y-%m",
				"date":   "$created_at",
			},
		}
	default:
		return nil, fmt.Errorf("unsupported groupBy value: %s", groupBy)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{
			Key: "$group", Value: bson.M{
				"_id":                groupID,
				"total_orders":       bson.M{"$sum": 1},
				"total_revenue":      bson.M{"$sum": "$total_price"},
				"average_order_value": bson.M{"$avg": "$total_price"},
				"created_at":          bson.M{"$first": "$created_at"}, // used for formatting
			},
		}},
		{{
			Key: "$project", Value: bson.M{
				"group_key":          groupKeyExpr,
				"total_orders":       1,
				"total_revenue":      1,
				"average_order_value": 1,
			},
		}},
		{{Key: "$sort", Value: bson.M{"group_key": 1}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []AggregatedStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
