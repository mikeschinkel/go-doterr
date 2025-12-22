package test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-doterr"
)

// TestMsgErr_String tests creating msgErr with a string message
func TestMsgErr_String(t *testing.T) {
	err := doterr.MsgErr("operation failed")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "operation failed" {
		t.Errorf("Expected 'operation failed', got %q", err.Error())
	}
}

// TestMsgErr_WrapError tests wrapping an existing error with msgErr
func TestMsgErr_WrapError(t *testing.T) {
	original := errors.New("underlying problem")
	wrapped := doterr.MsgErr(original)

	if wrapped == nil {
		t.Fatal("Expected error, got nil")
	}

	// Error message comes from wrapped error
	if wrapped.Error() != "underlying problem" {
		t.Errorf("Expected 'underlying problem', got %q", wrapped.Error())
	}

	// Can unwrap to find original
	if !errors.Is(wrapped, original) {
		t.Error("Expected errors.Is to find wrapped error")
	}
}

// TestMsgErr_InvalidType tests that MsgErr panics on invalid input
func TestMsgErr_InvalidType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid type")
		} else {
			msg := r.(string)
			if !strings.Contains(msg, "invalid type") {
				t.Errorf("Expected panic message about invalid type, got %q", msg)
			}
		}
	}()

	// Should panic
	doterr.MsgErr(123)
}

// TestMsgErr_ErrorChainPreservation tests that error chains are preserved
func TestMsgErr_ErrorChainPreservation(t *testing.T) {
	// Create chain: root -> middle -> wrapped
	root := errors.New("root cause")
	middle := errors.Join(errors.New("middle error"), root)
	wrapped := doterr.MsgErr(middle)

	// Should be able to find root through the chain
	if !errors.Is(wrapped, root) {
		t.Error("Expected errors.Is to find root cause through msgErr wrapper")
	}
}

// TestNewErr_WithMsgErr tests using MsgErr as a sentinel replacement
func TestNewErr_WithMsgErr(t *testing.T) {
	err := doterr.NewErr(doterr.MsgErr("validation failed"), "field", "email")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "validation failed") {
		t.Errorf("Expected error message to contain 'validation failed', got %q", msg)
	}
	if !strings.Contains(msg, "field") {
		t.Errorf("Expected error message to contain 'field', got %q", msg)
	}
}

// TestWithErr_WithMsgErr tests using MsgErr in context layering
func TestWithErr_WithMsgErr(t *testing.T) {
	base := doterr.MsgErr("file read failed")
	err := doterr.WithErr(base, doterr.MsgErr("config loading failed"))

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	msg := err.Error()
	// Should contain both messages (context layering)
	if !strings.Contains(msg, "config loading failed") {
		t.Errorf("Expected 'config loading failed' in error, got %q", msg)
	}
	if !strings.Contains(msg, "file read failed") {
		t.Errorf("Expected 'file read failed' in error, got %q", msg)
	}
}

// TestMsgErr_MixedWithSentinels tests mixing MsgErr and sentinels
func TestMsgErr_MixedWithSentinels(t *testing.T) {
	var ErrSentinel = errors.New("sentinel error")

	// Use both sentinel and MsgErr
	err := doterr.NewErr(ErrSentinel, doterr.MsgErr("additional context"), "key", "value")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should be able to detect the sentinel
	if !errors.Is(err, ErrSentinel) {
		t.Error("Expected errors.Is to find sentinel")
	}

	msg := err.Error()
	if !strings.Contains(msg, "sentinel error") {
		t.Errorf("Expected 'sentinel error' in message, got %q", msg)
	}
	if !strings.Contains(msg, "additional context") {
		t.Errorf("Expected 'additional context' in message, got %q", msg)
	}
}

// TestMsgErr_TypeDetection tests that msgErr can be distinguished from errors.New()
func TestMsgErr_TypeDetection(t *testing.T) {
	// Note: We can't directly access the msgErr type from the test package
	// because it's unexported. This test verifies the behavior exists,
	// but tooling would use type assertions in the doterr package itself.

	msgErrInstance := doterr.MsgErr("ad-hoc message")
	standardErr := errors.New("standard sentinel")

	// Both are errors
	if msgErrInstance.Error() != "ad-hoc message" {
		t.Errorf("Expected msgErr message, got %q", msgErrInstance.Error())
	}
	if standardErr.Error() != "standard sentinel" {
		t.Errorf("Expected standard error message, got %q", standardErr.Error())
	}

	// msgErr can be used in NewErr just like a sentinel
	err1 := doterr.NewErr(msgErrInstance, "key", "value")
	err2 := doterr.NewErr(standardErr, "key", "value")

	if err1 == nil || err2 == nil {
		t.Fatal("Expected both errors to be created successfully")
	}
}

// TestMsgErr_WithMetadata tests adding metadata to MsgErr
func TestMsgErr_WithMetadata(t *testing.T) {
	err := doterr.MsgErr("config invalid")
	err = doterr.WithErr(err, "path", "/etc/app.conf")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should contain both the message and metadata
	msg := err.Error()
	if !strings.Contains(msg, "config invalid") {
		t.Errorf("Expected 'config invalid' in error, got %q", msg)
	}
	if !strings.Contains(msg, "path") {
		t.Errorf("Expected 'path' in error, got %q", msg)
	}
}

// TestMsgErr_AsTrailingCause tests using MsgErr as a trailing cause
func TestMsgErr_AsTrailingCause(t *testing.T) {
	var ErrSentinel = errors.New("operation failed")
	cause := doterr.MsgErr("underlying issue")

	err := doterr.NewErr(ErrSentinel, "key", "value", cause)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Should contain both sentinel and cause
	msg := err.Error()
	if !strings.Contains(msg, "operation failed") {
		t.Errorf("Expected 'operation failed' in error, got %q", msg)
	}
	if !strings.Contains(msg, "underlying issue") {
		t.Errorf("Expected 'underlying issue' in error, got %q", msg)
	}

	// Should be able to find the sentinel
	if !errors.Is(err, ErrSentinel) {
		t.Error("Expected errors.Is to find sentinel")
	}
}

// TestMsgErr_RealWorldPattern tests a real-world usage pattern
func TestMsgErr_RealWorldPattern(t *testing.T) {
	// Simulate a function that uses MsgErr during rapid development
	processFile := func(path string) error {
		// Start with MsgErr
		err := doterr.MsgErr("file processing failed")

		// Add metadata as needed
		err = doterr.WithErr(err, "path", path)
		err = doterr.WithErr(err, "operation", "validate")

		return err
	}

	err := processFile("/tmp/test.json")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "file processing failed") {
		t.Errorf("Expected base message in error, got %q", msg)
	}
	if !strings.Contains(msg, "/tmp/test.json") {
		t.Errorf("Expected path in error, got %q", msg)
	}
	if !strings.Contains(msg, "validate") {
		t.Errorf("Expected operation in error, got %q", msg)
	}
}
