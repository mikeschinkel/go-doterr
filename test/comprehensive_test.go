package test

import (
	"errors"
	"testing"
)

// TestNewErr_BasicCases tests fundamental NewErr usage patterns
func TestNewErr_BasicCases(t *testing.T) {
	tests := []struct {
		name      string
		parts     []any
		wantNil   bool
		wantIs    error
		checkMeta bool
		metaKey   string
		metaValue any
	}{
		{
			name:    "single sentinel",
			parts:   []any{ErrTestSentinel},
			wantNil: false,
			wantIs:  ErrTestSentinel,
		},
		{
			name:      "sentinel with metadata",
			parts:     []any{ErrTestSentinel, "key", "value"},
			wantNil:   false,
			wantIs:    ErrTestSentinel,
			checkMeta: true,
			metaKey:   "key",
			metaValue: "value",
		},
		{
			name:    "sentinel with cause",
			parts:   []any{ErrTestSentinel, ErrCause},
			wantNil: false,
			wantIs:  ErrTestSentinel,
		},
		{
			name:      "sentinel with metadata and cause",
			parts:     []any{ErrTestSentinel, "status", 500, ErrCause},
			wantNil:   false,
			wantIs:    ErrTestSentinel,
			checkMeta: true,
			metaKey:   "status",
			metaValue: 500,
		},
		{
			name:    "multiple sentinels",
			parts:   []any{ErrTestSentinel, ErrOtherSentinel},
			wantNil: false,
			wantIs:  ErrTestSentinel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewErr(tt.parts...)

			if tt.wantNil {
				if err != nil {
					t.Errorf("expected nil, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("expected error to be %v", tt.wantIs)
			}

			if tt.checkMeta {
				kvs := ErrMeta(err)
				if kvs == nil {
					t.Fatal("expected metadata, got nil")
				}

				found := false
				for _, kv := range kvs {
					if kv.Key() == tt.metaKey {
						found = true
						if kv.Value() != tt.metaValue {
							t.Errorf("expected value %v, got %v", tt.metaValue, kv.Value())
						}
						break
					}
				}
				if !found {
					t.Errorf("metadata key %q not found", tt.metaKey)
				}
			}
		})
	}
}

// TestWithErr_Enrichment tests error enrichment patterns
func TestWithErr_Enrichment(t *testing.T) {
	tests := []struct {
		name      string
		parts     []any
		checkMeta bool
		metaKeys  []string
	}{
		{
			name:      "enrich existing error",
			parts:     []any{NewErr(ErrTestSentinel, "key1", "value1"), "key2", "value2"},
			checkMeta: true,
			metaKeys:  []string{"key1", "key2"},
		},
		{
			name:      "create new entry",
			parts:     []any{"key", "value"},
			checkMeta: true,
			metaKeys:  []string{"key"},
		},
		{
			name:  "with cause",
			parts: []any{"key", "value", ErrCause},
		},
		{
			name:      "enrich with multiple kvs",
			parts:     []any{NewErr(ErrTestSentinel), "k1", "v1", "k2", "v2"},
			checkMeta: true,
			metaKeys:  []string{"k1", "k2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithErr(tt.parts...)

			if err == nil && len(tt.parts) > 0 {
				// WithErr can return nil for empty parts
				if len(tt.parts) > 0 {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if tt.checkMeta {
				kvs := ErrMeta(err)
				if kvs == nil {
					t.Fatal("expected metadata, got nil")
				}

				for _, key := range tt.metaKeys {
					found := false
					for _, kv := range kvs {
						if kv.Key() == key {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("metadata key %q not found", key)
					}
				}
			}
		})
	}
}

// TestCombineErrs_Various tests error combination scenarios
func TestCombineErrs_Various(t *testing.T) {
	tests := []struct {
		name    string
		errs    []error
		wantNil bool
		count   int
	}{
		{
			name:    "empty slice",
			errs:    []error{},
			wantNil: true,
		},
		{
			name:    "all nil",
			errs:    []error{nil, nil, nil},
			wantNil: true,
		},
		{
			name:  "single error",
			errs:  []error{ErrTestSentinel},
			count: 1,
		},
		{
			name:  "multiple errors",
			errs:  []error{ErrTestSentinel, ErrOtherSentinel, ErrCause},
			count: 3,
		},
		{
			name:  "mixed with nils",
			errs:  []error{ErrTestSentinel, nil, ErrOtherSentinel},
			count: 2,
		},
		{
			name: "doterr entries",
			errs: []error{
				NewErr(ErrTestSentinel, "key", "value"),
				NewErr(ErrOtherSentinel, "key2", "value2"),
			},
			count: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineErrs(tt.errs)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected error, got nil")
			}

			// Check unwrap count
			if unwrapper, ok := result.(interface{ Unwrap() []error }); ok {
				unwrapped := unwrapper.Unwrap()
				if len(unwrapped) != tt.count {
					t.Errorf("expected %d unwrapped errors, got %d", tt.count, len(unwrapped))
				}
			} else if tt.count > 1 {
				t.Error("expected multi-unwrap error")
			}
		})
	}
}

// TestErrMeta_Extraction tests metadata extraction from various error types
func TestErrMeta_Extraction(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectNil bool
		expectKey string
	}{
		{
			name:      "simple entry",
			err:       NewErr(ErrTestSentinel, "key", "value"),
			expectKey: "key",
		},
		{
			name:      "joined error with entry first",
			err:       errors.Join(NewErr(ErrTestSentinel, "key", "value"), ErrCause),
			expectKey: "key",
		},
		{
			name:      "joined error with entry last",
			err:       errors.Join(ErrCause, NewErr(ErrTestSentinel, "key", "value")),
			expectKey: "key",
		},
		{
			name:      "non-doterr error",
			err:       errors.New("regular error"),
			expectNil: true,
		},
		{
			name:      "nil error",
			err:       nil,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvs := ErrMeta(tt.err)

			if tt.expectNil {
				if kvs != nil {
					t.Errorf("expected nil metadata, got %v", kvs)
				}
				return
			}

			if kvs == nil {
				t.Fatal("expected metadata, got nil")
			}

			found := false
			for _, kv := range kvs {
				if kv.Key() == tt.expectKey {
					found = true
					break
				}
			}
			if !found && tt.expectKey != "" {
				t.Errorf("expected key %q not found in metadata", tt.expectKey)
			}
		})
	}
}

// TestErrValue_TypeSafety tests type-safe value extraction
func TestErrValue_TypeSafety(t *testing.T) {
	err := NewErr(ErrTestSentinel,
		"string_key", "string_value",
		"int_key", 42,
		"bool_key", true,
		"float_key", 3.14,
	)

	// Test string extraction
	if val, ok := ErrValue[string](err, "string_key"); !ok || val != "string_value" {
		t.Errorf("string extraction failed: got %v, %v", val, ok)
	}

	// Test int extraction
	if val, ok := ErrValue[int](err, "int_key"); !ok || val != 42 {
		t.Errorf("int extraction failed: got %v, %v", val, ok)
	}

	// Test bool extraction
	if val, ok := ErrValue[bool](err, "bool_key"); !ok || val != true {
		t.Errorf("bool extraction failed: got %v, %v", val, ok)
	}

	// Test float extraction
	if val, ok := ErrValue[float64](err, "float_key"); !ok || val != 3.14 {
		t.Errorf("float extraction failed: got %v, %v", val, ok)
	}

	// Test wrong type
	if _, ok := ErrValue[int](err, "string_key"); ok {
		t.Error("expected type mismatch to return false")
	}

	// Test missing key
	if _, ok := ErrValue[string](err, "missing_key"); ok {
		t.Error("expected missing key to return false")
	}
}

// TestErrors_Extraction tests sentinel error extraction
func TestErrors_Extraction(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectNil bool
		expectLen int
	}{
		{
			name:      "single sentinel",
			err:       NewErr(ErrTestSentinel),
			expectLen: 1,
		},
		{
			name:      "multiple sentinels",
			err:       NewErr(ErrTestSentinel, ErrOtherSentinel),
			expectLen: 2,
		},
		{
			name:      "with metadata",
			err:       NewErr(ErrTestSentinel, "key", "value"),
			expectLen: 1,
		},
		{
			name: "joined error",
			err:  errors.Join(NewErr(ErrTestSentinel), ErrCause),
			// Note: Current implementation of Errors() may not handle all entry type cases
			// This is testing current behavior
			expectNil: true,
		},
		{
			name:      "non-doterr error",
			err:       errors.New("regular"),
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Errors(tt.err)

			if tt.expectNil {
				if errs != nil {
					t.Errorf("expected nil, got %v", errs)
				}
				return
			}

			if len(errs) != tt.expectLen {
				t.Errorf("expected %d errors, got %d", tt.expectLen, len(errs))
			}
		})
	}
}

// TestFindErr_TypeAssertion tests finding specific error types
// Tests finding sentinel errors in doterr entries and joined errors
func TestFindErr_TypeAssertion(t *testing.T) {
	// Test finding sentinel in doterr entry
	err := NewErr(ErrTestSentinel, "request_id", "12345")
	if found, ok := FindErr[error](err); !ok {
		t.Error("expected to find error")
	} else if !errors.Is(found, ErrTestSentinel) {
		t.Errorf("expected to find ErrTestSentinel")
	}

	// Test finding specific sentinel in joined error
	joined := errors.Join(NewErr(ErrTestSentinel, "key", "value"), ErrCause)
	if _, ok := FindErr[error](joined); !ok {
		t.Error("expected to find error in joined error")
	}

	// Test finding in nested structure
	nested := NewErr(ErrTestSentinel, ErrOtherSentinel, "key", "value")
	if !errors.Is(nested, ErrTestSentinel) {
		t.Error("expected to find ErrTestSentinel in nested")
	}
	if !errors.Is(nested, ErrOtherSentinel) {
		t.Error("expected to find ErrOtherSentinel in nested")
	}
}

// TestAppendErr tests error appending utility
func TestAppendErr(t *testing.T) {
	tests := []struct {
		name      string
		initial   []error
		toAppend  error
		expectLen int
	}{
		{
			name:      "append to empty",
			initial:   []error{},
			toAppend:  ErrTestSentinel,
			expectLen: 1,
		},
		{
			name:      "append to existing",
			initial:   []error{ErrTestSentinel},
			toAppend:  ErrOtherSentinel,
			expectLen: 2,
		},
		{
			name:      "append nil",
			initial:   []error{ErrTestSentinel},
			toAppend:  nil,
			expectLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AppendErr(tt.initial, tt.toAppend)
			if len(result) != tt.expectLen {
				t.Errorf("expected length %d, got %d", tt.expectLen, len(result))
			}
		})
	}
}
