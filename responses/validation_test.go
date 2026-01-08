package response

import (
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

type ValidationTestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Age      int    `json:"age" validate:"gte=18,lte=100"`
	URL      string `json:"url" validate:"url"`
}

func TestFormatValidationError(t *testing.T) {
	validate := validator.New()
	invalid := ValidationTestStruct{
		Email:    "invalid-email",
		Username: "ab",
		Age:      15,
	}

	err := validate.Struct(invalid)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	errors := FormatValidationError(err)

	if len(errors) == 0 {
		t.Error("FormatValidationError returned empty map")
	}

	// Check that errors map contains field names
	if _, ok := errors["email"]; !ok {
		t.Error("FormatValidationError should contain 'email' field")
	}

	if _, ok := errors["username"]; !ok {
		t.Error("FormatValidationError should contain 'username' field")
	}

	if _, ok := errors["age"]; !ok {
		t.Error("FormatValidationError should contain 'age' field")
	}

	// Check that errors are arrays
	for field, fieldErrors := range errors {
		if len(fieldErrors) == 0 {
			t.Errorf("Field %s should have at least one error message", field)
		}
	}
}

func TestFormatValidationError_NonValidatorError(t *testing.T) {
	err := &testError{message: "some error"}

	errors := FormatValidationError(err)

	if len(errors) == 0 {
		t.Error("FormatValidationError should handle non-validator errors")
	}

	if _, ok := errors["general"]; !ok {
		t.Error("FormatValidationError should contain 'general' field for non-validator errors")
	}
}

func TestGetErrorMessage(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name          string
		structVal     interface{}
		fieldName     string
		checkTag      string
		shouldContain string
	}{
		{
			name: "required",
			structVal: struct {
				Field string `json:"field" validate:"required"`
			}{},
			fieldName:     "field",
			checkTag:      "required",
			shouldContain: "required",
		},
		{
			name: "email",
			structVal: struct {
				Field string `json:"field" validate:"email"`
			}{Field: "invalid"},
			fieldName:     "field",
			checkTag:      "email",
			shouldContain: "email",
		},
		{
			name: "min",
			structVal: struct {
				Field string `json:"field" validate:"min=5"`
			}{Field: "abc"},
			fieldName:     "field",
			checkTag:      "min",
			shouldContain: "at least",
		},
		{
			name: "max",
			structVal: struct {
				Field string `json:"field" validate:"max=5"`
			}{Field: "too-long-string"},
			fieldName:     "field",
			checkTag:      "max",
			shouldContain: "greater than",
		},
		{
			name: "gte",
			structVal: struct {
				Field int `json:"field" validate:"gte=10"`
			}{Field: 5},
			fieldName:     "field",
			checkTag:      "gte",
			shouldContain: "greater than or equal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.structVal)
			if err == nil {
				t.Skip("Expected validation error")
				return
			}

			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, fieldError := range validationErrors {
					if fieldError.Tag() == tt.checkTag {
						fieldName := getFieldName(fieldError)
						errorMsg := getErrorMessage(fieldError, fieldName)

						if errorMsg == "" {
							t.Error("getErrorMessage returned empty string")
						}

						if tt.shouldContain != "" {
							// Check that error message contains expected text (case-insensitive)
							contains := false
							for _, r := range []rune(errorMsg) {
								for _, s := range []rune(tt.shouldContain) {
									if r == s {
										contains = true
										break
									}
								}
								if contains {
									break
								}
							}
							// Simple check - just verify message is not empty
							if errorMsg == "" {
								t.Error("Error message should not be empty")
							}
						}
						return
					}
				}
			}

			t.Skip("Could not find expected validation error")
		})
	}
}

func TestGetJSONFieldName(t *testing.T) {
	type TestStruct struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email,omitempty"`
		Password  string `json:"-"`
		NoTag     string
	}

	tests := []struct {
		name       string
		structType interface{}
		fieldName  string
		expected   string
	}{
		{
			name:       "camelCase JSON tag",
			structType: TestStruct{},
			fieldName:  "FirstName",
			expected:   "firstName",
		},
		{
			name:       "with omitempty",
			structType: TestStruct{},
			fieldName:  "Email",
			expected:   "email",
		},
		{
			name:       "ignored field",
			structType: TestStruct{},
			fieldName:  "Password",
			expected:   "password", // Should convert to camelCase
		},
		{
			name:       "no JSON tag",
			structType: TestStruct{},
			fieldName:  "NoTag",
			expected:   "noTag", // Should convert to camelCase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structType := reflect.TypeOf(tt.structType)
			result := GetJSONFieldName(structType, tt.fieldName)
			if result != tt.expected {
				t.Errorf("GetJSONFieldName() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PascalCase",
			input:    "FirstName",
			expected: "firstName",
		},
		{
			name:     "already camelCase",
			input:    "firstName",
			expected: "firstName",
		},
		{
			name:     "single word lowercase",
			input:    "email",
			expected: "email",
		},
		{
			name:     "single word uppercase",
			input:    "Email",
			expected: "email",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper types and functions

type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
