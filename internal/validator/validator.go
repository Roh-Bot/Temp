package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func NewValidator() *validator.Validate {
	v := validator.New()

	// Register custom uuid7 validator
	v.RegisterValidation("uuid7", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return validateUUIDv7(value)
	})

	return v
}

func validateUUIDv7(uuidString string) bool {
	if _, err := uuid.Parse(uuidString); err != nil {
		return false
	}
	return true
}
