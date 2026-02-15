package api

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

var tags = map[string]string{
	"required": "%s is required",
	"email":    "Invalid email format for %s",
}

func validationToErrorMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return ""
	}

	var messages []string
	for _, fieldErr := range validationErrors {
		fieldName := fieldErr.Field()

		if message, ok := tags[fieldErr.Tag()]; ok {
			messages = append(messages, fmt.Sprintf(message, fieldName))
			continue
		}
		messages = append(messages, fmt.Sprintf("Invalid value for '%s'", fieldName))
	}

	return joinMessages(messages, " | ")
}

func joinMessages(messages []string, separator string) string {
	if len(messages) == 0 {
		return ""
	}
	return strings.Join(messages, separator)
}
