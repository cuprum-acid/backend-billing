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

func TestValidatePrice(t *testing.T) {
	type priceStruct struct {
		Price string `validate:"price"`
	}

	cases := []struct {
		value string
		want  bool
	}{
		{"0", true},
		{"19", true},
		{"19.9", true},
		{"19.99", true},
		{"0.05", true},
		{"", false},
		{"banana", false},
		{".99", false},
		{"-1.00", false},
		{"19.999", false},
		{"19,99", false},
	}

	v := GetValidator()
	for _, tc := range cases {
		err := v.Struct(priceStruct{Price: tc.value})
		got := err == nil
		if got != tc.want {
			t.Errorf("validatePrice(%q): got accepted=%v, want %v (err=%v)", tc.value, got, tc.want, err)
		}
	}
}

func TestValidateBillingPeriod(t *testing.T) {
	type periodStruct struct {
		Period string `validate:"billing_period"`
	}

	v := GetValidator()
	for _, valid := range []string{"monthly", "yearly"} {
		if err := v.Struct(periodStruct{Period: valid}); err != nil {
			t.Errorf("validateBillingPeriod(%q) rejected unexpectedly: %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "weekly", "Monthly", "MONTHLY"} {
		if err := v.Struct(periodStruct{Period: invalid}); err == nil {
			t.Errorf("validateBillingPeriod(%q) accepted unexpectedly", invalid)
		}
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
