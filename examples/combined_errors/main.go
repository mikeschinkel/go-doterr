package main

import (
	"errors"
	"fmt"

	. "github.com/mikeschinkel/go-doterr"
)

// Define sentinel errors for different validation types
var (
	ErrValidation = errors.New("validation error")
	ErrNetwork    = errors.New("network error")
	ErrConfig     = errors.New("configuration error")
)

// Validate multiple fields independently
func validateUserInput(name, email string, age int) error {
	var errs []error

	// Validate name
	if name == "" {
		errs = append(errs, NewErr(ErrValidation, "field", "name", "error", "cannot be empty"))
	}

	// Validate email
	if email == "" || !contains(email, "@") {
		errs = append(errs, NewErr(ErrValidation, "field", "email", "error", "invalid format"))
	}

	// Validate age
	if age < 18 {
		errs = append(errs, NewErr(ErrValidation, "field", "age", "min", 18, "actual", age))
	}

	// Combine all validation errors into one
	return CombineErrs(errs)
}

// Simulate multiple independent operations that might fail
func processUserRegistration(name, email string, age int) error {
	var errs []error

	// Validation errors
	if validationErr := validateUserInput(name, email, age); validationErr != nil {
		errs = append(errs, validationErr)
	}

	// Simulate network check failure
	networkErr := NewErr(ErrNetwork, "service", "email-verification", "status", "unreachable")
	errs = append(errs, networkErr)

	// Simulate config error
	configErr := NewErr(ErrConfig, "file", "app.yaml", "error", "missing api_key")
	errs = append(errs, configErr)

	return CombineErrs(errs)
}

func main() {
	fmt.Println("=== Combined Errors Example ===\n")

	// Process a registration with multiple errors
	err := processUserRegistration("", "invalid-email", 15)

	if err != nil {
		fmt.Println("Registration failed with multiple errors:")
		fmt.Printf("\nCombined error string:\n%v\n\n", err)

		// Check for specific error types
		fmt.Println("Error type detection:")
		if errors.Is(err, ErrValidation) {
			fmt.Println("  ✓ Contains validation errors")
		}
		if errors.Is(err, ErrNetwork) {
			fmt.Println("  ✓ Contains network errors")
		}
		if errors.Is(err, ErrConfig) {
			fmt.Println("  ✓ Contains configuration errors")
		}

		// Unwrap to get individual errors
		fmt.Println("\nIndividual errors:")
		if unwrapper, ok := err.(interface{ Unwrap() []error }); ok {
			for i, e := range unwrapper.Unwrap() {
				fmt.Printf("  %d. %v\n", i+1, e)
			}
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
