package errs

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"strings"
)

func FormatValidationError(err error) map[string]string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		errorMap := make(map[string]string)
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errorMap[field] = "is required"
			case "min":
				errorMap[field] = "must be at least " + e.Param()
			case "max":
				errorMap[field] = "must be at most " + e.Param()
			case "len":
				errorMap[field] = "must be exactly " + e.Param() + " characters long"
			case "email":
				errorMap[field] = "must be a valid email address"
			case "url":
				errorMap[field] = "must be a valid URL"
			case "numeric":
				errorMap[field] = "must be a valid number"
			case "alpha":
				errorMap[field] = "must contain only letters"
			case "alphanum":
				errorMap[field] = "must contain only letters and numbers"
			case "datetime":
				errorMap[field] = "must be a valid datetime"
			case "gt":
				errorMap[field] = "must be greater than " + e.Param()
			case "gte":
				errorMap[field] = "must be greater than or equal to " + e.Param()
			case "lt":
				errorMap[field] = "must be less than " + e.Param()
			case "lte":
				errorMap[field] = "must be less than or equal to " + e.Param()
			default:
				errorMap[field] = "is invalid"
			}
		}
		return errorMap
	}

	return map[string]string{"parse_error": "-"}
}
