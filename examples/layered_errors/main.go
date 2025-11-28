package main

import (
	"errors"
	"fmt"

	. "github.com/mikeschinkel/go-doterr"
)

// Define sentinel errors for different layers
var (
	ErrDriver  = errors.New("driver error")
	ErrRepo    = errors.New("repository error")
	ErrService = errors.New("service error")
)

// Simulated database query result
type Result struct {
	Data string
}

// Driver layer - lowest level (simulates database interaction)
func queryDriver(userID int) (*Result, error) {
	// Simulate a database error
	dbErr := errors.New("connection timeout")

	return nil, NewErr(ErrDriver,
		"sql", "SELECT * FROM users WHERE id=?",
		"param", userID,
		dbErr, // trailing cause
	)
}

// Repository layer - middle layer
func getUser(userID int) (*Result, error) {
	result, err := queryDriver(userID)
	if err != nil {
		return nil, NewErr(ErrRepo,
			"table", "users",
			"operation", "fetch",
			err, // trailing cause from driver
		)
	}
	return result, nil
}

// Service layer - top layer
func fetchUserService(userID int) (*Result, error) {
	result, err := getUser(userID)
	if err != nil {
		return nil, NewErr(ErrService,
			"action", "GetUser",
			"user_id", userID,
			err, // trailing cause from repository
		)
	}
	return result, nil
}

func main() {
	fmt.Println("=== Layered Errors Example ===\n")

	// Call the service layer
	_, err := fetchUserService(42)

	if err != nil {
		fmt.Println("Error occurred:")
		fmt.Printf("  Full error: %v\n\n", err)

		// Check each layer using errors.Is
		fmt.Println("Layer detection:")
		if errors.Is(err, ErrDriver) {
			fmt.Println("  ✓ Contains driver layer error")
		}
		if errors.Is(err, ErrRepo) {
			fmt.Println("  ✓ Contains repository layer error")
		}
		if errors.Is(err, ErrService) {
			fmt.Println("  ✓ Contains service layer error")
		}

		// Extract metadata from top-level entry
		fmt.Println("\nTop-level metadata:")
		for _, kv := range ErrMeta(err) {
			fmt.Printf("  %s: %v\n", kv.Key(), kv.Value())
		}
	}
}
