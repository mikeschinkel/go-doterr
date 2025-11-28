package main

import (
	"errors"
	"fmt"

	. "github.com/mikeschinkel/go-doterr"
)

// Define sentinel errors for different categories
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

func main() {
	fmt.Println("=== Basic Usage Example ===\n")

	// Example 1: Simple error with sentinel and metadata
	err1 := NewErr(ErrValidation, "field", "email", "value", "invalid@")
	fmt.Println("Example 1 - Simple error with metadata:")
	fmt.Printf("  Error: %v\n", err1)
	fmt.Printf("  Is validation error: %v\n\n", errors.Is(err1, ErrValidation))

	// Example 2: Error with multiple metadata entries
	err2 := NewErr(ErrNotFound, "resource", "user", "id", 42, "query", "SELECT * FROM users")
	fmt.Println("Example 2 - Error with multiple metadata:")
	fmt.Printf("  Error: %v\n", err2)
	fmt.Printf("  Metadata: %v\n\n", ErrMeta(err2))

	// Example 3: Enriching an error with WithErr
	err3 := NewErr(ErrValidation, "field", "age")
	err3 = WithErr(err3, "min", 18, "actual", 15)
	fmt.Println("Example 3 - Enriching error with WithErr:")
	fmt.Printf("  Error: %v\n", err3)
	fmt.Printf("  Metadata: %v\n\n", ErrMeta(err3))

	// Example 4: Error with a cause
	causeErr := errors.New("database connection timeout")
	err4 := NewErr(ErrNotFound, "table", "users", causeErr)
	fmt.Println("Example 4 - Error with underlying cause:")
	fmt.Printf("  Error: %v\n", err4)
	fmt.Printf("  Has cause: %v\n", errors.Unwrap(err4) != nil)
}
