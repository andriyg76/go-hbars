package handlebars

import (
	"testing"
	"time"

	"github.com/andriyg76/go-hbars/runtime"
)

func TestFormatDate(t *testing.T) {
	// Test with time.Time
	now := time.Date(2026, 1, 25, 10, 30, 0, 0, time.UTC)
	result, err := FormatDate([]any{now})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25" {
		t.Errorf("expected '2026-01-25', got %q", result)
	}
	
	// Test with string
	result, err = FormatDate([]any{"2026-01-25"})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25" {
		t.Errorf("expected '2026-01-25', got %q", result)
	}
	
	// Test with custom format
	hash := runtime.Hash{"format": "2006-01-02 15:04:05"}
	result, err = FormatDate([]any{now, hash})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25 10:30:00" {
		t.Errorf("expected '2026-01-25 10:30:00', got %q", result)
	}
}

func TestDefault(t *testing.T) {
	// Test with truthy value
	result, err := Default([]any{"value", "default"})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got %q", result)
	}
	
	// Test with falsy value
	result, err = Default([]any{"", "default"})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}
	
	// Test with hash argument
	hash := runtime.Hash{"value": "fallback"}
	result, err = Default([]any{"", hash})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

func TestLookup(t *testing.T) {
	// Test with map: lookup key in object
	data := map[string]any{"name": "test"}
	result, err := Lookup([]any{data, "name"})
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %q", result)
	}
	
	// Lookup with nil object returns nil (no context in helpers)
	result, err = Lookup([]any{nil, "key"})
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for nil object, got %q", result)
	}
}

func TestUpperLower(t *testing.T) {
	result, err := Upper([]any{"hello"})
	if err != nil {
		t.Fatalf("Upper error: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", result)
	}
	
	result, err = Lower([]any{"WORLD"})
	if err != nil {
		t.Fatalf("Lower error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected 'world', got %q", result)
	}
}

func TestComparisonHelpers(t *testing.T) {
	// Test Eq
	result, err := Eq([]any{5, 5})
	if err != nil {
		t.Fatalf("Eq error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
	
	result, err = Eq([]any{5, 6})
	if err != nil {
		t.Fatalf("Eq error: %v", err)
	}
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
	
	// Test Lt
	result, err = Lt([]any{5, 10})
	if err != nil {
		t.Fatalf("Lt error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
	
	// Test Gt
	result, err = Gt([]any{10, 5})
	if err != nil {
		t.Fatalf("Gt error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestMathHelpers(t *testing.T) {
	// Test Add
	result, err := Add([]any{5, 3})
	if err != nil {
		t.Fatalf("Add error: %v", err)
	}
	if result != 8.0 {
		t.Errorf("expected 8.0, got %v", result)
	}
	
	// Test Multiply
	result, err = Multiply([]any{5, 3})
	if err != nil {
		t.Fatalf("Multiply error: %v", err)
	}
	if result != 15.0 {
		t.Errorf("expected 15.0, got %v", result)
	}
	
	// Test Divide
	result, err = Divide([]any{10, 2})
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}
	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestStringHelpers(t *testing.T) {
	// Test Capitalize
	result, err := Capitalize([]any{"hello"})
	if err != nil {
		t.Fatalf("Capitalize error: %v", err)
	}
	if result != "Hello" {
		t.Errorf("expected 'Hello', got %q", result)
	}
	
	// Test Truncate
	result, err = Truncate([]any{"hello world", 5})
	if err != nil {
		t.Fatalf("Truncate error: %v", err)
	}
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
	
	// Test Join
	result, err = Join([]any{[]any{"a", "b", "c"}, "-"})
	if err != nil {
		t.Fatalf("Join error: %v", err)
	}
	if result != "a-b-c" {
		t.Errorf("expected 'a-b-c', got %q", result)
	}
}

func TestCollectionHelpers(t *testing.T) {
	// Test Length
	result, err := Length([]any{[]any{1, 2, 3}})
	if err != nil {
		t.Fatalf("Length error: %v", err)
	}
	if result != 3 {
		t.Errorf("expected 3, got %v", result)
	}
	
	// Test First
	result, err = First([]any{[]any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("First error: %v", err)
	}
	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
	
	// Test Last
	result, err = Last([]any{[]any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("Last error: %v", err)
	}
	if result != "c" {
		t.Errorf("expected 'c', got %q", result)
	}
	
	// Test InArray
	result, err = InArray([]any{"b", []any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("InArray error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

