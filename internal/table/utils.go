package table

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func validateTable(v *validator.Validate, table Table) error {
	// Perform validation
	if err := v.Struct(table); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return nil
		}

		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			switch fieldErr.Tag() {
			case "required":
				return fmt.Errorf("%s is required", fieldErr.Field())
			}
		}
	}
	return nil
}
