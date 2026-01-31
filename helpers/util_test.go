package helpers

import (
	"testing"
)

func TestGetArg(t *testing.T) {
	args := []any{"a", 42, nil}
	if got := GetArg(args, 0); got != "a" {
		t.Errorf("GetArg(args, 0) = %v, want a", got)
	}
	if got := GetArg(args, 1); got != 42 {
		t.Errorf("GetArg(args, 1) = %v, want 42", got)
	}
	if got := GetArg(args, 2); got != nil {
		t.Errorf("GetArg(args, 2) = %v, want nil", got)
	}
	if got := GetArg(args, -1); got != nil {
		t.Errorf("GetArg(args, -1) = %v, want nil", got)
	}
	if got := GetArg(args, 3); got != nil {
		t.Errorf("GetArg(args, 3) = %v, want nil", got)
	}
	if got := GetArg(nil, 0); got != nil {
		t.Errorf("GetArg(nil, 0) = %v, want nil", got)
	}
}

func TestGetStringArg(t *testing.T) {
	if got := GetStringArg([]any{"hello"}, 0); got != "hello" {
		t.Errorf("GetStringArg = %q, want hello", got)
	}
	if got := GetStringArg([]any{42}, 0); got != "42" {
		t.Errorf("GetStringArg(42) = %q, want 42", got)
	}
	if got := GetStringArg([]any{nil}, 0); got != "" {
		t.Errorf("GetStringArg(nil) = %q, want empty", got)
	}
	if got := GetStringArg([]any{}, 0); got != "" {
		t.Errorf("GetStringArg out of bounds = %q, want empty", got)
	}
}

func TestGetNumberArg(t *testing.T) {
	tests := []struct {
		args []any
		idx  int
		want float64
		ok   bool
	}{
		{[]any{42}, 0, 42, true},
		{[]any{int64(10)}, 0, 10, true},
		{[]any{3.14}, 0, 3.14, true},
		{[]any{"1.5"}, 0, 1.5, true},
		{[]any{nil}, 0, 0, true},
		{[]any{"x"}, 0, 0, false},
		{[]any{}, 0, 0, true},
	}
	for _, tt := range tests {
		got, err := GetNumberArg(tt.args, tt.idx)
		if tt.ok && err != nil {
			t.Errorf("GetNumberArg(%v, %d) err = %v", tt.args, tt.idx, err)
			continue
		}
		if !tt.ok && err == nil {
			t.Errorf("GetNumberArg(%v, %d) expected error", tt.args, tt.idx)
			continue
		}
		if tt.ok && got != tt.want {
			t.Errorf("GetNumberArg(%v, %d) = %v, want %v", tt.args, tt.idx, got, tt.want)
		}
	}
}

func TestIsTruthy(t *testing.T) {
	if IsTruthy(nil) {
		t.Error("IsTruthy(nil) want false")
	}
	if !IsTruthy(true) {
		t.Error("IsTruthy(true) want true")
	}
	if IsTruthy(false) {
		t.Error("IsTruthy(false) want false")
	}
	if !IsTruthy("x") {
		t.Error("IsTruthy(\"x\") want true")
	}
	if IsTruthy("") {
		t.Error("IsTruthy(\"\") want false")
	}
	if !IsTruthy([]any{1}) {
		t.Error("IsTruthy([]any{1}) want true")
	}
	if IsTruthy([]any{}) {
		t.Error("IsTruthy([]any{}) want false")
	}
	if !IsTruthy(map[string]any{"a": 1}) {
		t.Error("IsTruthy(map) want true")
	}
	if IsTruthy(map[string]any{}) {
		t.Error("IsTruthy(empty map) want false")
	}
	if !IsTruthy(1) {
		t.Error("IsTruthy(1) want true")
	}
	if IsTruthy(0) {
		t.Error("IsTruthy(0) want false")
	}
	if !IsTruthy(3.14) {
		t.Error("IsTruthy(3.14) want true")
	}
	if IsTruthy(0.0) {
		t.Error("IsTruthy(0.0) want false")
	}
}

func TestIsEmpty(t *testing.T) {
	if !IsEmpty(nil) {
		t.Error("IsEmpty(nil) want true")
	}
	if !IsEmpty("") {
		t.Error("IsEmpty(\"\") want true")
	}
	if IsEmpty("x") {
		t.Error("IsEmpty(\"x\") want false")
	}
	if !IsEmpty([]any{}) {
		t.Error("IsEmpty([]any{}) want true")
	}
	if IsEmpty([]any{1}) {
		t.Error("IsEmpty([]any{1}) want false")
	}
	if !IsEmpty(map[string]any{}) {
		t.Error("IsEmpty(empty map) want true")
	}
	if IsEmpty(map[string]any{"a": 1}) {
		t.Error("IsEmpty(non-empty map) want false")
	}
	if !IsEmpty(0) {
		t.Error("IsEmpty(0) want true")
	}
	if IsEmpty(1) {
		t.Error("IsEmpty(1) want false")
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
