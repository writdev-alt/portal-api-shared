package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/writdev-alt/portal-api-shared/responses"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tag names in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return formatValidationError(err)
	}
	return nil
}

// formatValidationError formats validator errors into a user-friendly format
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMessages []string
		for _, fieldError := range validationErrors {
			errorMessages = append(errorMessages, getErrorMessage(fieldError))
		}
		return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
	}
	return err
}

// GetValidationErrors returns a map of field errors (useful for API responses)
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			// Field() already returns the json tag name due to RegisterTagNameFunc
			fieldName := fieldError.Field()
			errors[fieldName] = getErrorMessage(fieldError)
		}
	}
	return errors
}

// getErrorMessage returns a user-friendly error message for a validation error
func getErrorMessage(fieldError validator.FieldError) string {
	// Field() already returns the json tag name due to RegisterTagNameFunc
	fieldName := fieldError.Field()

	// Make field name more readable
	fieldName = strings.ReplaceAll(fieldName, "_", " ")
	fieldName = strings.Title(fieldName)

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fieldName, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fieldName, fieldError.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", fieldName, fieldError.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldName, fieldError.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fieldName, fieldError.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldName, fieldError.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fieldName, fieldError.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", fieldName, fieldError.Param())
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", fieldName, fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fieldName, fieldError.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fieldName)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", fieldName)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldName)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldName)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", fieldName)
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

// NewValidationErrorResponse creates a validation error response from validator errors
func NewValidationErrorResponse(err error) responses.ErrorResponse {
	validationErrors := GetValidationErrors(err)
	return responses.NewValidationErrorResponse(validationErrors)
}
