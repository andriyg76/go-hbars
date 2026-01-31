package runtime

import (
	"encoding/json"
	"testing"
)

func TestIsNumericZero(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"int-zero", 0, true},
		{"int-nonzero", 42, false},
		{"float-zero", 0.0, true},
		{"float-nonzero", 1.5, false},
		{"json.Number-zero", json.Number("0"), true},
		{"json.Number-int-zero", json.Number("0"), true},
		{"json.Number-float-zero", json.Number("0.0"), true},
		{"json.Number-nonzero", json.Number("3"), false},
		{"string", "0", false},
		{"bool-false", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNumericZero(tc.value); got != tc.want {
				t.Errorf("IsNumericZero(%v) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestIncludeZeroTruthy(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"int-zero", 0, true},
		{"int-nonzero", 42, true},
		{"float-zero", 0.0, true},
		{"string-empty", "", false},
		{"string-nonempty", "x", true},
		{"json.Number-zero", json.Number("0"), true},
		{"json.Number-nonzero", json.Number("1"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IncludeZeroTruthy(tc.value); got != tc.want {
				t.Errorf("IncludeZeroTruthy(%v) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func TestIsTruthy(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"bool-false", false, false},
		{"bool-true", true, true},
		{"empty-string", "", false},
		{"string", "ok", true},
		{"safe-string-empty", SafeString(""), false},
		{"safe-string", SafeString("ok"), true},
		{"int-zero", 0, false},
		{"int", 7, true},
		{"float-zero", 0.0, false},
		{"float", 1.25, true},
		{"empty-slice", []int{}, false},
		{"slice", []int{1}, true},
		{"empty-map", map[string]int{}, false},
		{"map", map[string]int{"a": 1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTruthy(tc.value); got != tc.want {
				t.Fatalf("IsTruthy(%v) = %v", tc.value, got)
			}
		})
	}
}

