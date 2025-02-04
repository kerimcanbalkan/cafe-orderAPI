package menu

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lithammer/shortuuid/v3"
	"go.mongodb.org/mongo-driver/mongo"
)

var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// IsAllowedImageType determines if image is among types defined
// in map of allowed images
func isAllowedImageType(mimeType string) bool {
	_, exists := validImageTypes[mimeType]

	return exists
}

func ValidateMenu(v *validator.Validate, item MenuItem) error {
	// Perform validation
	if err := v.Struct(item); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return nil
		}

		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				return fmt.Errorf("%s is required", fieldErr.Field())
			case "min":
				return fmt.Errorf(
					"%s must be at least %s characters",
					fieldErr.Field(),
					fieldErr.Param(),
				)
			case "max":
				return fmt.Errorf(
					"%s must be at most %s characters",
					fieldErr.Field(),
					fieldErr.Param(),
				)
			case "gt":
				return fmt.Errorf(
					"%s must be greater than %s",
					fieldErr.Field(),
					fieldErr.Param(),
				)
			case "number":
				return fmt.Errorf("%s must be a valid number", fieldErr.Field())
			default:
				return fmt.Errorf("%s is invalid", fieldErr.Field())
			}
		}
	}
	return nil
}

func handleMongoError(c *gin.Context, err error) {
	// Handle duplicate key error (for example, unique constraint violations)
	if mongo.IsDuplicateKeyError(err) {
		c.JSON(http.StatusConflict, gin.H{"error": "Item with this name already exists"})
		return
	}

	// Handle timeout errors
	if err.Error() == "context deadline exceeded" {
		c.JSON(
			http.StatusRequestTimeout,
			gin.H{"error": "Request timed out. Please try again later."},
		)
		return
	}

	// Handle unauthorized access errors
	if err.Error() == "unauthorized" {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "You do not have permission to perform this action."},
		)
		return
	}

	// Handle 'not found' errors
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Generic internal server error for other unexpected errors
	c.JSON(
		http.StatusInternalServerError,
		gin.H{"error": "An unexpected error occurred. Please try again later."},
	)
}

func generateImageName() string {
	return shortuuid.New()
}
