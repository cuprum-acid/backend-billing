package validator

import (
	"errors"
	"testing"
)

func TestGetValidator(t *testing.T) {
	v := GetValidator()
	if v == nil {
		t.Fatal("GetValidator returned nil")
	}

	// Test that validator works
	type TestStruct struct {
		Name string `validate:"required"`
	}

	valid := TestStruct{Name: "test"}
	invalid := TestStruct{Name: ""}

	if err := v.Struct(valid); err != nil {
		t.Errorf("Expected no error for valid struct, got %v", err)
	}

	if err := v.Struct(invalid); err == nil {
		t.Error("Expected error for invalid struct, got nil")
	}
}

func TestValidationErrors(t *testing.T) {
	type TestStruct struct {
		Name     string `validate:"required,min=3"`
		Email    string `validate:"required,email"`
		Age      int    `validate:"required,min=18"`
		Currency string `validate:"required,currency"`
	}

	tests := []struct {
		name          string
		input         TestStruct
		expectedError bool
		expectedField string
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:     "John",
				Email:    "john@example.com",
				Age:      25,
				Currency: "USD",
			},
			expectedError: false,
		},
		{
			name: "missing name",
			input: TestStruct{
				Email:    "john@example.com",
				Age:      25,
				Currency: "USD",
			},
			expectedError: true,
			expectedField: "Name",
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:     "John",
				Email:    "invalid",
				Age:      25,
				Currency: "USD",
			},
			expectedError: true,
			expectedField: "Email",
		},
		{
			name: "invalid currency",
			input: TestStruct{
				Name:     "John",
				Email:    "john@example.com",
				Age:      25,
				Currency: "INVALID",
			},
			expectedError: true,
			expectedField: "Currency",
		},
	}

	v := GetValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)

			if tt.expectedError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.expectedError && tt.expectedField != "" {
				var validationErrors ValidationErrors
				if errors.As(err, &validationErrors) {
					found := false
					for _, ve := range validationErrors {
						if ve.Field() == tt.expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field %s, got errors for: %v", tt.expectedField, validationErrors)
					}
				}
			}
		})
	}
}
