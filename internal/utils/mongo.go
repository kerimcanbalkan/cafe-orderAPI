package utils

import (
	"errors"
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

	var mongoErr mongo.CommandError
	if errors.As(err, &mongoErr) {
		switch mongoErr.Code {
		case 11000: // Duplicate key error
			c.JSON(
				http.StatusConflict,
				gin.H{
					"error": "Duplicate entry detected. A record with the same unique value already exists.",
				},
			)
		case 121: // Document validation failure
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "Invalid data format. Please ensure all required fields are correctly provided.",
				},
			)
		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "Database operation failed. Please try again later."},
			)
		}
		return
	}

	switch err {
	case mongo.ErrNoDocuments:
		c.JSON(http.StatusNotFound, gin.H{"error": "Requested resource not found."})
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
		// Handle duplicate key errors dynamically
		if strings.Contains(err.Error(), "duplicate key error collection") {
			c.JSON(
				http.StatusConflict,
				gin.H{"error": "Duplicate key error. Ensure unique values for required fields."},
			)
		} else if strings.Contains(err.Error(), "document validation failure") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Document validation failed. Check your input data."})
		} else if strings.Contains(err.Error(), "network timeout") {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Database request timed out. Please try again."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected database error occurred."})
		}
	}
}
