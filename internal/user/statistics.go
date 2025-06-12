package user

import (
	"context"
	"time"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AggregatedWaiterStats struct {
	GroupKey           string  `bson:"group_key" json:"groupKey"`                      // day, week or year
	TotalOrdersServed  int     `bson:"total_orders_served" json:"totalOrdersServed"`   // minutes
	AverageServingTime float64 `bson:"average_serving_time" json:"averageServingTime"` // minutes
	FastestServingTime float64     `bson:"fastest_serving_time" json:"fastestServingTime"` // minutesn
}

type AggregatedCashierStats struct {
	GroupKey          string  `bson:"group_key" json:"groupKey"` // day, week or year
	TotalOrdersClosed int     `bson:"total_orders_closed" json:"totalOrdersClosed"`
	TotalRevenue      float64 `bson:"total_revenue"       json:"totalRevenue"`
}

type WaiterStats struct {
	TotalOrdersServed  int                     `bson:"total_orders_served"  json:"totalOrdersServed"`
	AverageServingTime float64                 `bson:"average_serving_time" json:"averageServingTime"` // minutes
	FastestServingTime int                     `bson:"fastest_serving_time" json:"fastestServingTime"` // minutes
	AggregatedStats    []AggregatedWaiterStats `json:"aggregatedStats"`
}

type CashierStats struct {
	TotalOrdersClosed int                      `bson:"total_orders_closed" json:"totalOrdersClosed"`
	TotalRevenue      float64                  `bson:"total_revenue"       json:"totalRevenue"`
	AggregatedStats   []AggregatedCashierStats `json:"aggregatedStats"`
}

// getStats calculates waiter statistics for a given date range.
// Needs orders collection NOT users collection
func getWaiterStats(
	startDate, endDate time.Time,
	ctx context.Context,
	collection *mongo.Collection,
	userID primitive.ObjectID,
	groupBy string,
) (WaiterStats, error) {
	matchFilter := bson.M{
		"served_at":  bson.M{"$gte": startDate, "$lt": endDate},
		"handled_by": userID,
	}

	var groupID bson.M
	var groupKeyExpr bson.M

	switch groupBy {
	case "day":
		groupID = bson.M{"year": bson.M{"$year": "$created_at"}, "month": bson.M{"$month": "$created_at"}, "day": bson.M{"$dayOfMonth": "$created_at"}}
		groupKeyExpr = bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"}}
	case "week":
		groupID = bson.M{"year": bson.M{"$isoWeekYear": "$created_at"}, "week": bson.M{"$isoWeek": "$created_at"}}
		groupKeyExpr = bson.M{
			"$concat": []interface{}{
				bson.M{"$toString": bson.M{"$isoWeekYear": "$created_at"}},
				"-W",
				bson.M{"$toString": bson.M{"$isoWeek": "$created_at"}},
			},
		}
	case "month":
		groupID = bson.M{"year": bson.M{"$year": "$created_at"}, "month": bson.M{"$month": "$created_at"}}
		groupKeyExpr = bson.M{"$dateToString": bson.M{"format": "%Y-%m", "date": "$created_at"}}
	default:
		return WaiterStats{}, fmt.Errorf("unsupported groupBy value: %s", groupBy)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$addFields", Value: bson.M{
			"servingTime": bson.M{"$subtract": []interface{}{"$served_at", "$created_at"}},
		}}},
		{{Key: "$facet", Value: bson.M{
			"overall": []bson.M{
				{"$group": bson.M{
					"_id": nil,
					"total_orders_served":  bson.M{"$sum": 1},
					"average_serving_time": bson.M{"$avg": "$servingTime"},
					"fastest_serving_time": bson.M{"$min": "$servingTime"},
				}},
			},
			"grouped": []bson.M{
				{"$group": bson.M{
					"_id": groupID,
					"total_orders_served":  bson.M{"$sum": 1},
					"average_serving_time": bson.M{"$avg": "$servingTime"},
					"fastest_serving_time": bson.M{"$min": "$servingTime"},
					"created_at":           bson.M{"$first": "$created_at"},
				}},
				{"$project": bson.M{
					"group_key":           groupKeyExpr,
					"total_orders_served": 1,
					"average_serving_time": bson.M{"$divide": []interface{}{"$average_serving_time", 60000}}, // ms to min
					"fastest_serving_time": bson.M{"$divide": []interface{}{"$fastest_serving_time", 60000}}, // ms to min
				}},
				{"$sort": bson.M{"group_key": 1}},
			},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return WaiterStats{}, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Overall []struct {
			TotalOrdersServed  int     `bson:"total_orders_served"`
			AverageServingTime float64 `bson:"average_serving_time"`
			FastestServingTime int     `bson:"fastest_serving_time"`
		} `bson:"overall"`
		Grouped []AggregatedWaiterStats `bson:"grouped"`
	}

	if err := cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return WaiterStats{}, err
	}

	var stats WaiterStats
	if len(results[0].Overall) > 0 {
		stats.TotalOrdersServed = results[0].Overall[0].TotalOrdersServed
		stats.AverageServingTime = results[0].Overall[0].AverageServingTime / 60000.0
		stats.FastestServingTime = int(float64(results[0].Overall[0].FastestServingTime) / 60000.0)
	}
	stats.AggregatedStats = results[0].Grouped

	return stats, nil
}


// getCashierStats calculates cashier statistics for a given date range.
// Needs orders collection NOT users collection
func getCashierStats(
	startDate, endDate time.Time,
	ctx context.Context,
	collection *mongo.Collection,
	userID primitive.ObjectID,
	groupBy string,
) (CashierStats, error) {
	matchFilter := bson.M{
		"closed_at": bson.M{"$gte": startDate, "$lt": endDate},
		"closed_by": userID,
	}

	var groupID bson.M
	var groupKeyExpr bson.M

	switch groupBy {
	case "day":
		groupID = bson.M{"year": bson.M{"$year": "$closed_at"}, "month": bson.M{"$month": "$closed_at"}, "day": bson.M{"$dayOfMonth": "$closed_at"}}
		groupKeyExpr = bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$closed_at"}}
	case "week":
		groupID = bson.M{"year": bson.M{"$isoWeekYear": "$closed_at"}, "week": bson.M{"$isoWeek": "$closed_at"}}
		groupKeyExpr = bson.M{
			"$concat": []interface{}{
				bson.M{"$toString": bson.M{"$isoWeekYear": "$closed_at"}},
				"-W",
				bson.M{"$toString": bson.M{"$isoWeek": "$closed_at"}},
			},
		}
	case "month":
		groupID = bson.M{"year": bson.M{"$year": "$closed_at"}, "month": bson.M{"$month": "$closed_at"}}
		groupKeyExpr = bson.M{"$dateToString": bson.M{"format": "%Y-%m", "date": "$closed_at"}}
	default:
		return CashierStats{}, fmt.Errorf("unsupported groupBy value: %s", groupBy)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchFilter}},
		{{Key: "$facet", Value: bson.M{
			"overall": []bson.M{
				{"$group": bson.M{
					"_id":                nil,
					"total_orders_closed": bson.M{"$sum": 1},
					"total_revenue":       bson.M{"$sum": "$total_price"},
				}},
			},
			"grouped": []bson.M{
				{"$group": bson.M{
					"_id":               groupID,
					"total_orders_closed": bson.M{"$sum": 1},
					"total_revenue":       bson.M{"$sum": "$total_price"},
				}},
				{"$project": bson.M{
					"group_key":          groupKeyExpr,
					"total_orders_closed": 1,
					"total_revenue":       1,
				}},
				{"$sort": bson.M{"group_key": 1}},
			},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return CashierStats{}, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Overall []struct {
			TotalOrdersClosed int     `bson:"total_orders_closed"`
			TotalRevenue      float64 `bson:"total_revenue"`
		} `bson:"overall"`
		Grouped []AggregatedCashierStats `bson:"grouped"`
	}

	if err := cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return CashierStats{}, err
	}

	var stats CashierStats
	if len(results[0].Overall) > 0 {
		stats.TotalOrdersClosed = results[0].Overall[0].TotalOrdersClosed
		stats.TotalRevenue = results[0].Overall[0].TotalRevenue
	}
	stats.AggregatedStats = results[0].Grouped

	return stats, nil
}
