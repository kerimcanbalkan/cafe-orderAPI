package order

import (
	"fmt"

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
