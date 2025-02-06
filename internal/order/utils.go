package order

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func validateOrder(v *validator.Validate, order Order) error {
	// Perform validation
	if err := v.Struct(order); err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return nil
		}

		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				return fmt.Errorf("%s is required", fieldErr.Field())
			default:
				return fmt.Errorf("%s is invalid", fieldErr.Field())
			}
		}
	}
	return nil
}

func convertTableNumber(s string) (uint8, error) {
	// Convert string to an integer
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	// Ensure it's within the uint8 range (0-255)
	if value < 0 || value > 255 {
		return 0, fmt.Errorf("value %d out of range for uint8", value)
	}

	return uint8(value), nil
}
