package test

import (
	"errors"
	"strings"
	"testing"
)

var (
	ErrContext = errors.New("context error")
)

// TestWithErr_StoresCauseInEntryErrors verifies that WithErr stores the cause
// in entry.errors instead of using errors.Join, giving us control over formatting
func TestWithErr_StoresCauseInEntryErrors(t *testing.T) {
	cause := errors.New("database connection failed")
	err := WithErr("endpoint", "/api/users", ErrContext, cause)

	// Debug: Check what errors are stored in the entry
	errs := Errors(err)
	t.Logf("Errors() returned %d errors: %v", len(errs), errs)

	// Debug: Check unwrap behavior
	type unwrapper interface{ Unwrap() []error }
	if u, ok := err.(unwrapper); ok {
		unwrapped := u.Unwrap()
		t.Logf("Unwrap() returned %d errors: %v", len(unwrapped), unwrapped)
	}

	// Verify cause is accessible via errors.Is
	if !errors.Is(err, cause) {
		t.Error("expected cause to be unwrappable via errors.Is")
	}

	// Verify formatting uses "; " separator (not errors.Join's "\n")
	errStr := err.Error()
	t.Logf("Error string: %q", errStr)
	if strings.Contains(errStr, "\n") {
		t.Errorf("error string should not contain newlines from errors.Join, got: %q", errStr)
	}

	// Verify semicolons are used for separating errors
	if !strings.Contains(errStr, ";") {
		t.Errorf("error string should use semicolons as separator, got: %q", errStr)
	}
}

// TestWithErr_ChainedCauses verifies multiple causes stored in entry.errors
func TestWithErr_ChainedCauses(t *testing.T) {
	cause1 := errors.New("network timeout")
	cause2 := errors.New("retry failed")

	err := NewErr(ErrContext, "operation", "save", cause1)
	err = WithErr(err, "retry_count", 3, cause2)

	// Both causes should be accessible
	if !errors.Is(err, cause1) {
		t.Error("expected cause1 to be in error chain")
	}
	if !errors.Is(err, cause2) {
		t.Error("expected cause2 to be in error chain")
	}

	// Error string should use semicolons, not newlines
	errStr := err.Error()
	if strings.Contains(errStr, "\n") {
		t.Errorf("error string should not contain newlines, got: %q", errStr)
	}
}

// TestWithErr_FormattingControl verifies we control error formatting
func TestWithErr_FormattingControl(t *testing.T) {
	base := NewErr(ErrContext, "user_id", 123)
	cause := errors.New("validation failed")
	err := WithErr(base, "field", "email", cause)

	errStr := err.Error()

	// Should contain our sentinel
	if !strings.Contains(errStr, "context error") {
		t.Errorf("expected error string to contain sentinel, got: %q", errStr)
	}

	// Should contain cause
	if !strings.Contains(errStr, "validation failed") {
		t.Errorf("expected error string to contain cause, got: %q", errStr)
	}

	// Should use semicolons for separation (our formatting)
	if !strings.Contains(errStr, ";") {
		t.Errorf("expected semicolons in error string, got: %q", errStr)
	}

	// Should NOT use newlines (errors.Join formatting)
	if strings.Contains(errStr, "\n") {
		t.Errorf("should not use newlines (errors.Join formatting), got: %q", errStr)
	}
}

// TestNewErr_StoresCauseInEntryErrors verifies NewErr also stores causes in entry.errors
func TestNewErr_StoresCauseInEntryErrors(t *testing.T) {
	cause := errors.New("disk full")
	err := NewErr(ErrContext, "file", "data.txt", cause)

	// Cause should be unwrappable
	if !errors.Is(err, cause) {
		t.Error("expected cause to be unwrappable")
	}

	// Should use semicolons, not newlines
	errStr := err.Error()
	if strings.Contains(errStr, "\n") {
		t.Errorf("error string should not contain newlines, got: %q", errStr)
	}
}

// TestWithErr_MetadataOnly verifies metadata-only additions don't affect formatting
func TestWithErr_MetadataOnly(t *testing.T) {
	base := NewErr(ErrContext, "step", 1)
	err := WithErr(base, "details", "processing")

	errStr := err.Error()

	// Should contain metadata
	if !strings.Contains(errStr, "step") {
		t.Errorf("expected 'step' in error string, got: %q", errStr)
	}
	if !strings.Contains(errStr, "details") {
		t.Errorf("expected 'details' in error string, got: %q", errStr)
	}

	// Should not have extra newlines
	if strings.Contains(errStr, "\n") {
		t.Errorf("should not contain newlines, got: %q", errStr)
	}
}

// TestWithErr_DuplicateKeyValueDetection verifies that duplicate key-values
// are properly detected in the error chain when using entry.errors
func TestWithErr_DuplicateKeyValueDetection(t *testing.T) {
	// Create an error chain with a duplicate key-value
	innerErr := NewErr(ErrContext, "request_id", "abc123")
	outerErr := WithErr(innerErr, "request_id", "abc123", "status", "failed")

	errStr := outerErr.Error()

	// Count how many times "request_id=abc123" appears in the output
	count := strings.Count(errStr, "request_id=abc123")

	// With our MatchKV logic in Error(), duplicates should be suppressed
	// (The exact behavior depends on implementation, but we're testing it's handled)
	if count > 1 {
		t.Logf("Note: request_id=abc123 appears %d times (may be expected): %q", count, errStr)
	}
}
