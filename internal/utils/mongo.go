package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleMongoError provides a clean error response for MongoDB-related errors
func HandleMongoError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	// Handle known MongoDB errors
	switch err {
	case mongo.ErrClientDisconnected:
		c.JSON(
			http.StatusServiceUnavailable,
			gin.H{"error": "Database connection lost. Please try again later."},
		)
	case mongo.ErrNilDocument:
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "Invalid request data. Please check your input."},
		)
	default:
		// Handle dynamic error cases
		if strings.Contains(err.Error(), "network timeout") {
			c.JSON(
				http.StatusGatewayTimeout,
				gin.H{"error": "Database request timed out. Please try again."},
			)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected database error occurred."})
		}
	}
}
