package handlebars

import (
	"testing"
	"time"

	"github.com/andriyg76/go-hbars/runtime"
)

func TestFormatDate(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test with time.Time
	now := time.Date(2026, 1, 25, 10, 30, 0, 0, time.UTC)
	result, err := FormatDate(ctx, []any{now})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25" {
		t.Errorf("expected '2026-01-25', got %q", result)
	}
	
	// Test with string
	result, err = FormatDate(ctx, []any{"2026-01-25"})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25" {
		t.Errorf("expected '2026-01-25', got %q", result)
	}
	
	// Test with custom format
	hash := runtime.Hash{"format": "2006-01-02 15:04:05"}
	result, err = FormatDate(ctx, []any{now, hash})
	if err != nil {
		t.Fatalf("FormatDate error: %v", err)
	}
	if result != "2026-01-25 10:30:00" {
		t.Errorf("expected '2026-01-25 10:30:00', got %q", result)
	}
}

func TestDefault(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test with truthy value
	result, err := Default(ctx, []any{"value", "default"})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got %q", result)
	}
	
	// Test with falsy value
	result, err = Default(ctx, []any{"", "default"})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "default" {
		t.Errorf("expected 'default', got %q", result)
	}
	
	// Test with hash argument
	hash := runtime.Hash{"value": "fallback"}
	result, err = Default(ctx, []any{"", hash})
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

func TestLookup(t *testing.T) {
	ctx := runtime.NewContext(map[string]any{
		"key": "value",
	})
	
	// Test with map
	data := map[string]any{"name": "test"}
	result, err := Lookup(ctx, []any{data, "name"})
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %q", result)
	}
	
	// Test with context path
	result, err = Lookup(ctx, []any{nil, "key"})
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got %q", result)
	}
}

func TestUpperLower(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	result, err := Upper(ctx, []any{"hello"})
	if err != nil {
		t.Fatalf("Upper error: %v", err)
	}
	if result != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", result)
	}
	
	result, err = Lower(ctx, []any{"WORLD"})
	if err != nil {
		t.Fatalf("Lower error: %v", err)
	}
	if result != "world" {
		t.Errorf("expected 'world', got %q", result)
	}
}

func TestComparisonHelpers(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test Eq
	result, err := Eq(ctx, []any{5, 5})
	if err != nil {
		t.Fatalf("Eq error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
	
	result, err = Eq(ctx, []any{5, 6})
	if err != nil {
		t.Fatalf("Eq error: %v", err)
	}
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
	
	// Test Lt
	result, err = Lt(ctx, []any{5, 10})
	if err != nil {
		t.Fatalf("Lt error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
	
	// Test Gt
	result, err = Gt(ctx, []any{10, 5})
	if err != nil {
		t.Fatalf("Gt error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestMathHelpers(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test Add
	result, err := Add(ctx, []any{5, 3})
	if err != nil {
		t.Fatalf("Add error: %v", err)
	}
	if result != 8.0 {
		t.Errorf("expected 8.0, got %v", result)
	}
	
	// Test Multiply
	result, err = Multiply(ctx, []any{5, 3})
	if err != nil {
		t.Fatalf("Multiply error: %v", err)
	}
	if result != 15.0 {
		t.Errorf("expected 15.0, got %v", result)
	}
	
	// Test Divide
	result, err = Divide(ctx, []any{10, 2})
	if err != nil {
		t.Fatalf("Divide error: %v", err)
	}
	if result != 5.0 {
		t.Errorf("expected 5.0, got %v", result)
	}
}

func TestStringHelpers(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test Capitalize
	result, err := Capitalize(ctx, []any{"hello"})
	if err != nil {
		t.Fatalf("Capitalize error: %v", err)
	}
	if result != "Hello" {
		t.Errorf("expected 'Hello', got %q", result)
	}
	
	// Test Truncate
	result, err = Truncate(ctx, []any{"hello world", 5})
	if err != nil {
		t.Fatalf("Truncate error: %v", err)
	}
	if result != "hello..." {
		t.Errorf("expected 'hello...', got %q", result)
	}
	
	// Test Join
	result, err = Join(ctx, []any{[]any{"a", "b", "c"}, "-"})
	if err != nil {
		t.Fatalf("Join error: %v", err)
	}
	if result != "a-b-c" {
		t.Errorf("expected 'a-b-c', got %q", result)
	}
}

func TestCollectionHelpers(t *testing.T) {
	ctx := runtime.NewContext(nil)
	
	// Test Length
	result, err := Length(ctx, []any{[]any{1, 2, 3}})
	if err != nil {
		t.Fatalf("Length error: %v", err)
	}
	if result != 3 {
		t.Errorf("expected 3, got %v", result)
	}
	
	// Test First
	result, err = First(ctx, []any{[]any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("First error: %v", err)
	}
	if result != "a" {
		t.Errorf("expected 'a', got %q", result)
	}
	
	// Test Last
	result, err = Last(ctx, []any{[]any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("Last error: %v", err)
	}
	if result != "c" {
		t.Errorf("expected 'c', got %q", result)
	}
	
	// Test InArray
	result, err = InArray(ctx, []any{"b", []any{"a", "b", "c"}})
	if err != nil {
		t.Fatalf("InArray error: %v", err)
	}
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

