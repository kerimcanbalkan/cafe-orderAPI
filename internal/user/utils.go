package user

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidateUser(v *validator.Validate, user User) error {
	// Perform validation
	if err := v.Struct(user); err != nil {

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
			case "oneof":
				return fmt.Errorf(fieldErr.Error())
			case "email":
				return fmt.Errorf("%s must be a valid email", fieldErr.Field())
			default:
				return fmt.Errorf("%s is invalid", fieldErr.Field())
			}
		}
	}
	return nil
}
