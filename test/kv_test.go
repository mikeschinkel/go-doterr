package test

import (
	"errors"
	"strings"
	"testing"
)

// Test typed KV constructors
func TestTypedKVConstructors(t *testing.T) {
	tests := []struct {
		name    string
		kv      ErrKV
		wantKey string
		wantVal any
	}{
		{
			name:    "StringKV",
			kv:      StringKV("name", "Alice"),
			wantKey: "name",
			wantVal: "Alice",
		},
		{
			name:    "IntKV",
			kv:      IntKV("count", 42),
			wantKey: "count",
			wantVal: 42,
		},
		{
			name:    "Int64KV",
			kv:      Int64KV("id", int64(9876543210)),
			wantKey: "id",
			wantVal: int64(9876543210),
		},
		{
			name:    "BoolKV",
			kv:      BoolKV("active", true),
			wantKey: "active",
			wantVal: true,
		},
		{
			name:    "Float64KV",
			kv:      Float64KV("price", 19.99),
			wantKey: "price",
			wantVal: 19.99,
		},
		{
			name:    "AnyKV",
			kv:      AnyKV("data", map[string]int{"a": 1}),
			wantKey: "data",
			wantVal: map[string]int{"a": 1},
		},
		{
			name:    "ErrorKV",
			kv:      ErrorKV("cause", errors.New("test error")),
			wantKey: "cause",
			wantVal: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kv.Key(); got != tt.wantKey {
				t.Errorf("Key() = %v, want %v", got, tt.wantKey)
			}

			// Special handling for error values (can't use == for errors)
			if errVal, ok := tt.wantVal.(error); ok {
				gotErr, ok := tt.kv.Value().(error)
				if !ok {
					t.Errorf("Value() type = %T, want error", tt.kv.Value())
				} else if gotErr.Error() != errVal.Error() {
					t.Errorf("Value() = %v, want %v", gotErr, errVal)
				}
			} else {
				// Use string comparison for maps (can't use == directly)
				if _, isMap := tt.wantVal.(map[string]int); isMap {
					gotMap, ok := tt.kv.Value().(map[string]int)
					if !ok || gotMap["a"] != 1 {
						t.Errorf("Value() = %v, want %v", tt.kv.Value(), tt.wantVal)
					}
				} else if got := tt.kv.Value(); got != tt.wantVal {
					t.Errorf("Value() = %v, want %v", got, tt.wantVal)
				}
			}
		})
	}
}

// Test typed constructors in NewErr
func TestTypedKVInNewErr(t *testing.T) {
	sentinel := errors.New("test error")

	err := NewErr(sentinel,
		StringKV("name", "Alice"),
		IntKV("age", 30),
		BoolKV("active", true),
	)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify metadata
	meta := ErrMeta(err)
	if len(meta) != 3 {
		t.Fatalf("Expected 3 metadata items, got %d", len(meta))
	}

	// Check each metadata item
	if meta[0].Key() != "name" || meta[0].Value() != "Alice" {
		t.Errorf("meta[0] = %v:%v, want name:Alice", meta[0].Key(), meta[0].Value())
	}
	if meta[1].Key() != "age" || meta[1].Value() != 30 {
		t.Errorf("meta[1] = %v:%v, want age:30", meta[1].Key(), meta[1].Value())
	}
	if meta[2].Key() != "active" || meta[2].Value() != true {
		t.Errorf("meta[2] = %v:%v, want active:true", meta[2].Key(), meta[2].Value())
	}
}

// Test AppendKV with KV values
func TestAppendKVWithKVValues(t *testing.T) {
	var kvs []ErrKV

	kvs = AppendKV(kvs, StringKV("name", "Bob"))
	kvs = AppendKV(kvs, IntKV("age", 25))

	if len(kvs) != 2 {
		t.Fatalf("Expected 2 KV values, got %d", len(kvs))
	}

	if kvs[0].Key() != "name" || kvs[0].Value() != "Bob" {
		t.Errorf("kvs[0] = %v:%v, want name:Bob", kvs[0].Key(), kvs[0].Value())
	}
	if kvs[1].Key() != "age" || kvs[1].Value() != 25 {
		t.Errorf("kvs[1] = %v:%v, want age:25", kvs[1].Key(), kvs[1].Value())
	}
}

// Test AppendKV with string pairs
func TestAppendKVWithStringPairs(t *testing.T) {
	var kvs []ErrKV

	kvs = AppendKV(kvs, "city", "NYC")
	kvs = AppendKV(kvs, "country", "USA")

	if len(kvs) != 2 {
		t.Fatalf("Expected 2 KV values, got %d", len(kvs))
	}

	if kvs[0].Key() != "city" || kvs[0].Value() != "NYC" {
		t.Errorf("kvs[0] = %v:%v, want city:NYC", kvs[0].Key(), kvs[0].Value())
	}
	if kvs[1].Key() != "country" || kvs[1].Value() != "USA" {
		t.Errorf("kvs[1] = %v:%v, want country:USA", kvs[1].Key(), kvs[1].Value())
	}
}

// Test AppendKV with mixed inputs
func TestAppendKVWithMixedInputs(t *testing.T) {
	var kvs []ErrKV

	kvs = AppendKV(kvs,
		StringKV("name", "Charlie"),
		"age", 35,
		BoolKV("active", false),
	)

	if len(kvs) != 3 {
		t.Fatalf("Expected 3 KV values, got %d", len(kvs))
	}

	if kvs[0].Key() != "name" || kvs[0].Value() != "Charlie" {
		t.Errorf("kvs[0] = %v:%v, want name:Charlie", kvs[0].Key(), kvs[0].Value())
	}
	if kvs[1].Key() != "age" || kvs[1].Value() != 35 {
		t.Errorf("kvs[1] = %v:%v, want age:35", kvs[1].Key(), kvs[1].Value())
	}
	if kvs[2].Key() != "active" || kvs[2].Value() != false {
		t.Errorf("kvs[2] = %v:%v, want active:false", kvs[2].Key(), kvs[2].Value())
	}
}

// Test AppendKV panic on trailing key
func TestAppendKVPanicOnTrailingKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on trailing key, got none")
		} else {
			msg := r.(string)
			if !strings.Contains(msg, "trailing key") {
				t.Errorf("Expected panic message about trailing key, got: %v", msg)
			}
		}
	}()

	var kvs []ErrKV
	kvs = AppendKV(kvs, "incomplete")
}

// Test AppendKV panic on invalid type
func TestAppendKVPanicOnInvalidType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on invalid type, got none")
		} else {
			msg := r.(string)
			if !strings.Contains(msg, "invalid type") {
				t.Errorf("Expected panic message about invalid type, got: %v", msg)
			}
		}
	}()

	var kvs []ErrKV
	kvs = AppendKV(kvs, 123)
}

// Test lazy evaluation timing
func TestLazyEvaluationTiming(t *testing.T) {
	evaluationCount := 0

	var kvs []ErrKV
	kvs = AppendKV(kvs, func() ErrKV {
		evaluationCount++
		return StringKV("expensive", "computed")
	})

	// After AppendKV, function should NOT have been evaluated yet
	if evaluationCount != 0 {
		t.Errorf("Lazy function evaluated at append time, expected 0 evaluations, got %d", evaluationCount)
	}

	// Happy path - no error created
	// Function should still not be evaluated
	if evaluationCount != 0 {
		t.Errorf("Lazy function evaluated in happy path, expected 0 evaluations, got %d", evaluationCount)
	}

	// Error path - create error with kvs
	sentinel := errors.New("test error")
	err := NewErr(sentinel, kvs)

	// Now function should have been evaluated exactly once
	if evaluationCount != 1 {
		t.Errorf("Lazy function not evaluated at error creation, expected 1 evaluation, got %d", evaluationCount)
	}

	// Verify metadata contains the lazy value
	meta := ErrMeta(err)
	if len(meta) != 1 {
		t.Fatalf("Expected 1 metadata item, got %d", len(meta))
	}
	if meta[0].Key() != "expensive" || meta[0].Value() != "computed" {
		t.Errorf("meta[0] = %v:%v, want expensive:computed", meta[0].Key(), meta[0].Value())
	}

	// Calling Error() again should not re-evaluate (sync.Once)
	_ = err.Error()
	if evaluationCount != 1 {
		t.Errorf("Lazy function evaluated multiple times, expected 1 evaluation, got %d", evaluationCount)
	}
}

// Test []KV in NewErr
func TestSliceKVInNewErr(t *testing.T) {
	var kvs []ErrKV
	kvs = AppendKV(kvs, "user_id", 123)
	kvs = AppendKV(kvs, "action", "login")

	sentinel := errors.New("auth error")
	err := NewErr(sentinel, kvs)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	meta := ErrMeta(err)
	if len(meta) != 2 {
		t.Fatalf("Expected 2 metadata items, got %d", len(meta))
	}

	if meta[0].Key() != "user_id" || meta[0].Value() != 123 {
		t.Errorf("meta[0] = %v:%v, want user_id:123", meta[0].Key(), meta[0].Value())
	}
	if meta[1].Key() != "action" || meta[1].Value() != "login" {
		t.Errorf("meta[1] = %v:%v, want action:login", meta[1].Key(), meta[1].Value())
	}
}

// Test []KV in WithErr
func TestSliceKVInWithErr(t *testing.T) {
	var kvs []ErrKV
	kvs = AppendKV(kvs, "retry_count", 3)

	sentinel1 := errors.New("base error")
	err1 := NewErr(sentinel1, "request_id", "abc123")

	sentinel2 := errors.New("enriched error")
	err2 := WithErr(err1, sentinel2, kvs)

	meta := ErrMeta(err2)
	if len(meta) < 2 {
		t.Fatalf("Expected at least 2 metadata items, got %d", len(meta))
	}

	// Should have both original and new metadata
	hasRequestID := false
	hasRetryCount := false
	for _, kv := range meta {
		if kv.Key() == "request_id" && kv.Value() == "abc123" {
			hasRequestID = true
		}
		if kv.Key() == "retry_count" && kv.Value() == 3 {
			hasRetryCount = true
		}
	}

	if !hasRequestID {
		t.Error("Missing original metadata request_id:abc123")
	}
	if !hasRetryCount {
		t.Error("Missing new metadata retry_count:3")
	}
}

// Test direct func()KV in NewErr
func TestDirectFuncKVInNewErr(t *testing.T) {
	evaluationCount := 0

	sentinel := errors.New("test error")
	err := NewErr(sentinel, func() ErrKV {
		evaluationCount++
		return StringKV("lazy", "value")
	})

	// Function should have been evaluated once at error creation
	if evaluationCount != 1 {
		t.Errorf("Expected 1 evaluation, got %d", evaluationCount)
	}

	meta := ErrMeta(err)
	if len(meta) != 1 {
		t.Fatalf("Expected 1 metadata item, got %d", len(meta))
	}

	if meta[0].Key() != "lazy" || meta[0].Value() != "value" {
		t.Errorf("meta[0] = %v:%v, want lazy:value", meta[0].Key(), meta[0].Value())
	}
}

// Test mixed old and new styles
func TestMixedOldAndNewStyles(t *testing.T) {
	sentinel := errors.New("mixed error")
	err := NewErr(sentinel,
		"old_key", "old_value", // Old style
		StringKV("new_key", "new_value"), // New style
		"another_old", 42,                // Old style
		IntKV("another_new", 99), // New style
	)

	meta := ErrMeta(err)
	if len(meta) != 4 {
		t.Fatalf("Expected 4 metadata items, got %d", len(meta))
	}

	// Verify all metadata is present
	expected := map[string]any{
		"old_key":     "old_value",
		"new_key":     "new_value",
		"another_old": 42,
		"another_new": 99,
	}

	for _, kv := range meta {
		wantVal, ok := expected[kv.Key()]
		if !ok {
			t.Errorf("Unexpected key: %v", kv.Key())
			continue
		}
		if kv.Value() != wantVal {
			t.Errorf("For key %v: got value %v, want %v", kv.Key(), kv.Value(), wantVal)
		}
	}
}

// Test parity checking with []KV
func TestParityCheckingWithSliceKV(t *testing.T) {
	var kvs []ErrKV
	kvs = AppendKV(kvs, "field", "value")

	sentinel := errors.New("test error")
	cause := errors.New("cause error")

	// []KV should be treated as single item, so this should have trailing cause
	err := NewErr(sentinel, kvs, cause)

	// Verify cause is present
	if !errors.Is(err, cause) {
		t.Error("Cause error not found in error chain")
	}
}

// Test parity checking with func()KV
func TestParityCheckingWithFuncKV(t *testing.T) {
	sentinel := errors.New("test error")
	cause := errors.New("cause error")

	// func()KV should be treated as single item, so this should have trailing cause
	err := NewErr(sentinel, func() ErrKV {
		return StringKV("lazy", "value")
	}, cause)

	// Verify cause is present
	if !errors.Is(err, cause) {
		t.Error("Cause error not found in error chain")
	}
}

// Test real-world usage pattern: accumulating metadata
func TestRealWorldAccumulateMetadata(t *testing.T) {
	// Simulate a function that accumulates metadata
	processFile := func(path string, size int, shouldFail bool) error {
		var kvs []ErrKV
		kvs = AppendKV(kvs, "path", path)
		kvs = AppendKV(kvs, "size", size)

		if size > 1000 {
			kvs = AppendKV(kvs, "truncated", true)
		}

		if shouldFail {
			sentinel := errors.New("processing failed")
			return NewErr(sentinel, kvs)
		}

		return nil
	}

	// Happy path - no error, kvs not used
	err := processFile("/tmp/small.txt", 500, false)
	if err != nil {
		t.Errorf("Expected no error in happy path, got %v", err)
	}

	// Error path with metadata
	err = processFile("/tmp/large.txt", 2000, true)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	meta := ErrMeta(err)
	if len(meta) != 3 {
		t.Fatalf("Expected 3 metadata items, got %d", len(meta))
	}

	// Verify metadata
	expected := map[string]any{
		"path":      "/tmp/large.txt",
		"size":      2000,
		"truncated": true,
	}

	for _, kv := range meta {
		wantVal, ok := expected[kv.Key()]
		if !ok {
			t.Errorf("Unexpected key: %v", kv.Key())
			continue
		}
		if kv.Value() != wantVal {
			t.Errorf("For key %v: got value %v, want %v", kv.Key(), kv.Value(), wantVal)
		}
	}
}
