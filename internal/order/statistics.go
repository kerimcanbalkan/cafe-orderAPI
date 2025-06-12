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

type Stats struct {
	TotalOrders int `json:"totalOrders"`
	TotalRevenue int `json:"totelRevenue"`
	AverageOrderValue int `json:"averageOrderValue"`
	AggregatedStats []AggregatedStat `json:"aggregatedStats"`
}

func getStats(
	ctx context.Context,
	collection *mongo.Collection,
	from time.Time,
	to time.Time,
	groupBy string, // "day", "week", or "month"
) (Stats, error) {
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
		return Stats{}, fmt.Errorf("unsupported groupBy value: %s", groupBy)
	}

	// Define the facet stage with two pipelines
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{
			Key: "$facet", Value: bson.M{
				"grouped": []bson.M{
					{"$group": bson.M{
						"_id":                 groupID,
						"total_orders":        bson.M{"$sum": 1},
						"total_revenue":       bson.M{"$sum": "$total_price"},
						"average_order_value": bson.M{"$avg": "$total_price"},
						"created_at":          bson.M{"$first": "$created_at"},
					}},
					{"$project": bson.M{
						"group_key":           groupKeyExpr,
						"total_orders":        1,
						"total_revenue":       1,
						"average_order_value": 1,
					}},
					{"$sort": bson.M{"group_key": 1}},
				},
				"overall": []bson.M{
					{"$group": bson.M{
						"_id":                 nil,
						"total_orders":        bson.M{"$sum": 1},
						"total_revenue":       bson.M{"$sum": "$total_price"},
						"average_order_value": bson.M{"$avg": "$total_price"},
					}},
				},
			},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return Stats{}, err
	}
	defer cursor.Close(ctx)

	var facetResult []struct {
		Grouped []AggregatedStat `bson:"grouped"`
		Overall []struct {
			TotalOrders       int     `bson:"total_orders"`
			TotalRevenue      float64 `bson:"total_revenue"`
			AverageOrderValue float64 `bson:"average_order_value"`
		} `bson:"overall"`
	}

	if err := cursor.All(ctx, &facetResult); err != nil {
		return Stats{}, err
	}

	var finalStats Stats
	if len(facetResult) > 0 {
		if len(facetResult[0].Overall) > 0 {
			finalStats.TotalOrders = facetResult[0].Overall[0].TotalOrders
			finalStats.TotalRevenue = int(facetResult[0].Overall[0].TotalRevenue)
			finalStats.AverageOrderValue = int(facetResult[0].Overall[0].AverageOrderValue)
		}
		finalStats.AggregatedStats = facetResult[0].Grouped
	}

	return finalStats, nil
}
