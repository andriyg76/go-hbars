package helpers

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "x", false},
		{"empty slice", []any{}, true},
		{"non-empty slice", []any{1}, false},
		{"empty map", map[string]any{}, true},
		{"non-empty map", map[string]any{"a": 1}, false},
		{"zero", 0, true},
		{"non-zero", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsEmpty(tt.v)
			if err != nil {
				t.Fatalf("IsEmpty(%v): %v", tt.v, err)
			}
			if got != tt.want {
				t.Errorf("IsEmpty(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}
	for _, s := range tests {
		_, err := ParseTime(s)
		if err != nil {
			t.Errorf("ParseTime(%q) = %v", s, err)
		}
	}
	if _, err := ParseTime("not-a-date"); err == nil {
		t.Error("ParseTime(\"not-a-date\") expected error")
	}
}
