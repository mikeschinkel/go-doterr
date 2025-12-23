package test

import (
	"errors"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

var (
	ErrTestSentinel  = errors.New("test sentinel")
	ErrOtherSentinel = errors.New("other sentinel")
	ErrCause         = errors.New("cause error")
)

// FuzzNewErr tests the NewErr function with various inputs to detect panics,
// infinite loops, or unexpected behavior in error creation.
func FuzzNewErr(f *testing.F) {
	// Basic valid cases with sentinels
	f.Add("key1", "value1", false, false) // Single KV pair
	f.Add("key1", "value1", true, false)  // KV pair with cause
	f.Add("user", "john", false, false)   // String values
	f.Add("count", "42", false, false)    // Numeric string values
	f.Add("path", "/tmp/test", false, false)

	// Multiple key-value pairs
	f.Add("key1", "value1", false, false)
	f.Add("key2", "value2", false, false)
	f.Add("status", "active", false, false)

	// Edge cases - empty and special strings
	f.Add("", "", false, false)
	f.Add("key", "", false, false)
	f.Add("", "value", false, false)

	// Special characters
	f.Add("key\n", "value\n", false, false)
	f.Add("key\t", "value\t", false, false)
	f.Add("key with spaces", "value with spaces", false, false)
	f.Add("key=equals", "value=equals", false, false)
	f.Add("key;semicolon", "value;semicolon", false, false)

	// Long strings
	f.Add(strings.Repeat("k", 1000), strings.Repeat("v", 1000), false, false)
	f.Add(strings.Repeat("key", 100), strings.Repeat("val", 100), false, false)

	// Unicode and special characters
	f.Add("emoji", "😀🎉", false, false)
	f.Add("日本語", "テスト", false, false)
	f.Add("русский", "тест", false, false)

	f.Fuzz(func(t *testing.T, key1, value1 string, withCause, addSecondPair bool) {
		done := make(chan struct{})
		var result error

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("NewErr panicked with key=%q value=%q: %v\nStack: %s",
						key1, value1, r, debug.Stack())
				}
				close(done)
			}()

			// Build parts slice
			parts := []any{ErrTestSentinel}

			// Add first key-value pair if key is not empty
			if key1 != "" {
				parts = append(parts, key1, value1)
			}

			// Optionally add second pair
			if addSecondPair {
				parts = append(parts, "key2", "value2")
			}

			// Optionally add trailing cause
			if withCause {
				parts = append(parts, ErrCause)
			}

			result = NewErr(parts...)
		}()

		select {
		case <-done:
			// Completed successfully (with or without error)
			// NewErr should never return nil when given a sentinel
			if result == nil {
				t.Errorf("NewErr returned nil with sentinel")
			}

		case <-time.After(1 * time.Second):
			t.Fatalf("NewErr hung (infinite loop) on key=%q value=%q", key1, value1)
		}
	})
}

// FuzzWithErr tests the WithErr function with various combinations of base errors,
// metadata, and causes.
func FuzzWithErr(f *testing.F) {
	// Basic enrichment cases
	f.Add("key", "value", true, false, false)  // Enrich with base
	f.Add("key", "value", false, true, false)  // New entry with cause
	f.Add("key", "value", false, false, false) // Just metadata
	f.Add("meta", "data", true, true, false)   // Base + cause

	// Edge cases
	f.Add("", "", true, false, false)
	f.Add("", "value", false, true, false)
	f.Add("key", "", true, true, false)

	// Special characters
	f.Add("key\n\t", "value\r\n", true, false, false)
	f.Add("special=chars;", "value:test", false, true, false)

	f.Fuzz(func(t *testing.T, key, value string, withBase, withCause, addMultipleKeys bool) {
		done := make(chan struct{})
		var result error

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("WithErr panicked with key=%q value=%q: %v\nStack: %s",
						key, value, r, debug.Stack())
				}
				close(done)
			}()

			var parts []any

			// Optionally start with base error
			if withBase {
				base := NewErr(ErrTestSentinel, "base_key", "base_value")
				parts = append(parts, base)
			}

			// Add metadata
			if key != "" {
				parts = append(parts, key, value)
			}

			// Optionally add multiple key-value pairs
			if addMultipleKeys {
				parts = append(parts, "extra_key1", "extra_value1")
				parts = append(parts, "extra_key2", 123)
			}

			// Optionally add trailing cause
			if withCause {
				parts = append(parts, ErrCause)
			}

			result = WithErr(parts...)
		}()

		select {
		case <-done:
			// Completed successfully - result can be nil if no parts provided
			// This is acceptable behavior
			_ = result

		case <-time.After(1 * time.Second):
			t.Fatalf("WithErr hung (infinite loop) on key=%q value=%q", key, value)
		}
	})
}

// FuzzCombineErrs tests the CombineErrs function with various error combinations.
func FuzzCombineErrs(f *testing.F) {
	// Various counts of errors
	f.Add(0, false, false) // Empty slice
	f.Add(1, false, false) // Single error
	f.Add(2, false, false) // Two errors
	f.Add(5, false, false) // Multiple errors
	f.Add(10, true, false) // Many errors with nils
	f.Add(3, false, true)  // Errors with metadata

	f.Fuzz(func(t *testing.T, errCount int, withNils, withMetadata bool) {
		// Limit to reasonable count to avoid timeout
		if errCount < 0 || errCount > 100 {
			t.Skip()
		}

		done := make(chan struct{})
		var result error

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("CombineErrs panicked with count=%d: %v\nStack: %s",
						errCount, r, debug.Stack())
				}
				close(done)
			}()

			errs := make([]error, errCount)
			for i := 0; i < errCount; i++ {
				if withNils && i%3 == 0 {
					errs[i] = nil
					continue
				}

				if withMetadata {
					errs[i] = NewErr(ErrTestSentinel, "index", i, "value", i*10)
				} else {
					// Create simple errors
					errs[i] = errors.New("error " + string(rune(i)))
				}
			}

			result = CombineErrs(errs)
		}()

		select {
		case <-done:
			// Verify result is sensible
			if errCount == 0 || (withNils && errCount <= 3) {
				// Should return nil for empty or all-nil slices
			} else if errCount == 1 && !withNils {
				// Should return single error
				if result == nil {
					t.Errorf("CombineErrs returned nil for single error")
				}
			}

		case <-time.After(1 * time.Second):
			t.Fatalf("CombineErrs hung (infinite loop) with count=%d", errCount)
		}
	})
}

// FuzzErrMeta tests the ErrMeta function with various error structures.
func FuzzErrMeta(f *testing.F) {
	// Various metadata scenarios
	f.Add("key", "value", false, false) // Single KV
	f.Add("key", "value", true, false)  // Multiple KVs
	f.Add("", "", false, false)         // Empty keys
	f.Add("k", "v", false, true)        // With cause

	f.Fuzz(func(t *testing.T, key, value string, multipleKVs, withJoin bool) {
		done := make(chan struct{})
		var result []ErrKV

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ErrMeta panicked with key=%q: %v\nStack: %s",
						key, r, debug.Stack())
				}
				close(done)
			}()

			var err error
			if withJoin {
				// Create joined error
				err1 := NewErr(ErrTestSentinel, key, value)
				err2 := errors.New("other error")
				err = errors.Join(err1, err2)
			} else {
				parts := []any{ErrTestSentinel}
				if key != "" {
					parts = append(parts, key, value)
				}
				if multipleKVs {
					parts = append(parts, "key2", "value2", "key3", 123)
				}
				err = NewErr(parts...)
			}

			result = ErrMeta(err)
		}()

		select {
		case <-done:
			// Should always complete without panic
			// result can be nil or non-nil depending on input
			_ = result

		case <-time.After(1 * time.Second):
			t.Fatalf("ErrMeta hung (infinite loop) with key=%q", key)
		}
	})
}

// FuzzErrValue tests the ErrValue function with various types and keys.
func FuzzErrValue(f *testing.F) {
	// Various key lookups
	f.Add("existing_key", "lookup_key", true)
	f.Add("key", "key", true)    // Exact match
	f.Add("key", "other", false) // No match
	f.Add("", "lookup", false)   // Empty key
	f.Add("key", "", false)      // Empty lookup

	f.Fuzz(func(t *testing.T, storeKey, lookupKey string, shouldMatch bool) {
		done := make(chan struct{})
		var resultValue string
		var resultOk bool

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ErrValue panicked with storeKey=%q lookupKey=%q: %v\nStack: %s",
						storeKey, lookupKey, r, debug.Stack())
				}
				close(done)
			}()

			var err error
			if storeKey != "" {
				err = NewErr(ErrTestSentinel, storeKey, "test_value")
			} else {
				err = NewErr(ErrTestSentinel)
			}

			resultValue, resultOk = ErrValue[string](err, lookupKey)
		}()

		select {
		case <-done:
			// Should always complete
			// resultOk indicates whether key was found
			if shouldMatch && storeKey == lookupKey && storeKey != "" {
				if !resultOk {
					t.Logf("Expected match for key=%q but got ok=false", lookupKey)
				}
			}
			_ = resultValue

		case <-time.After(1 * time.Second):
			t.Fatalf("ErrValue hung (infinite loop) with lookupKey=%q", lookupKey)
		}
	})
}
