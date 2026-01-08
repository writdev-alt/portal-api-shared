package utils

import (
	"strings"
	"testing"
)

type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=20"`
	Age      int    `json:"age" validate:"gte=18,lte=100"`
	URL      string `json:"url" validate:"url"`
}

func TestGetValidator(t *testing.T) {
	v := GetValidator()

	if v == nil {
		t.Error("GetValidator returned nil")
	}

	// Test that validator works
	testStruct := TestStruct{
		Email:    "test@example.com",
		Username: "testuser",
		Age:      25,
		URL:      "https://example.com",
	}

	if err := v.Struct(testStruct); err != nil {
		t.Errorf("GetValidator returned validator that failed on valid struct: %v", err)
	}
}

func TestValidateStruct(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      25,
				URL:      "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "missing required email",
			input: TestStruct{
				Username: "testuser",
				Age:      25,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			input: TestStruct{
				Email:    "invalid-email",
				Username: "testuser",
				Age:      25,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "username too short",
			input: TestStruct{
				Email:    "test@example.com",
				Username: "ab",
				Age:      25,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "username too long",
			input: TestStruct{
				Email:    "test@example.com",
				Username: strings.Repeat("a", 21),
				Age:      25,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "age too young",
			input: TestStruct{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      15,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "age too old",
			input: TestStruct{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      101,
				URL:      "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			input: TestStruct{
				Email:    "test@example.com",
				Username: "testuser",
				Age:      25,
				URL:      "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.input)

			if tt.wantErr && err == nil {
				t.Error("ValidateStruct expected error but got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateStruct unexpected error: %v", err)
			}
		})
	}
}

func TestValidateStruct_JSONTagName(t *testing.T) {
	// Test that validation errors use JSON tag names
	invalid := TestStruct{
		Email: "invalid-email",
	}

	err := ValidateStruct(invalid)
	if err == nil {
		t.Error("ValidateStruct expected error but got nil")
	}

	// Check that error message contains JSON tag name
	errStr := err.Error()
	if !strings.Contains(errStr, "email") {
		t.Errorf("ValidateStruct error should contain JSON tag name 'email', got: %s", errStr)
	}
}

func TestGetValidationErrors(t *testing.T) {
	invalid := TestStruct{
		Email:    "invalid-email",
		Username: "ab",
	}

	err := ValidateStruct(invalid)
	if err == nil {
		t.Error("ValidateStruct expected error but got nil")
	}

	errors := GetValidationErrors(err)

	if len(errors) == 0 {
		t.Error("GetValidationErrors returned empty map")
	}

	// Check that errors map contains field names
	if _, ok := errors["email"]; !ok {
		t.Error("GetValidationErrors should contain 'email' field")
	}

	if _, ok := errors["username"]; !ok {
		t.Error("GetValidationErrors should contain 'username' field")
	}
}

func BenchmarkValidateStruct(b *testing.B) {
	valid := TestStruct{
		Email:    "test@example.com",
		Username: "testuser",
		Age:      25,
		URL:      "https://example.com",
	}

	for i := 0; i < b.N; i++ {
		ValidateStruct(valid)
	}
}
