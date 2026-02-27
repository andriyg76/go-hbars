package runtime

import (
	"testing"
)

func TestHelperArgs_GetArg(t *testing.T) {
	args := HelperArgs{Args: []any{"a", 42, nil}}
	if got := args.GetArg(0); got != "a" {
		t.Errorf("GetArg(0) = %v, want a", got)
	}
	if got := args.GetArg(1); got != 42 {
		t.Errorf("GetArg(1) = %v, want 42", got)
	}
	if got := args.GetArg(2); got != nil {
		t.Errorf("GetArg(2) = %v, want nil", got)
	}
	if got := args.GetArg(-1); got != nil {
		t.Errorf("GetArg(-1) = %v, want nil", got)
	}
	if got := args.GetArg(3); got != nil {
		t.Errorf("GetArg(3) = %v, want nil", got)
	}
	empty := HelperArgs{Args: nil}
	if got := empty.GetArg(0); got != nil {
		t.Errorf("GetArg(nil Args, 0) = %v, want nil", got)
	}
}

func TestHelperArgs_GetString(t *testing.T) {
	args := HelperArgs{Args: []any{"hello"}}
	if got := args.GetString(0); got != "hello" {
		t.Errorf("GetString(0) = %q, want hello", got)
	}
	args.Args = []any{42}
	if got := args.GetString(0); got != "42" {
		t.Errorf("GetString(42) = %q, want 42", got)
	}
	args.Args = []any{nil}
	if got := args.GetString(0); got != "" {
		t.Errorf("GetString(nil) = %q, want empty", got)
	}
	args.Args = []any{}
	if got := args.GetString(0); got != "" {
		t.Errorf("GetString out of bounds = %q, want empty", got)
	}
}

func TestHelperArgs_GetNumber(t *testing.T) {
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
		a := HelperArgs{Args: tt.args}
		got, err := a.GetNumber(tt.idx)
		if tt.ok && err != nil {
			t.Errorf("GetNumber(%v, %d) err = %v", tt.args, tt.idx, err)
			continue
		}
		if !tt.ok && err == nil {
			t.Errorf("GetNumber(%v, %d) expected error", tt.args, tt.idx)
			continue
		}
		if tt.ok && got != tt.want {
			t.Errorf("GetNumber(%v, %d) = %v, want %v", tt.args, tt.idx, got, tt.want)
		}
	}
}

func TestHelperArgs_GetHash(t *testing.T) {
	args := HelperArgs{HashArgs: map[string]any{"k": "v", "n": 42}}
	if got := args.GetHash("k"); got != "v" {
		t.Errorf("GetHash(k) = %v, want v", got)
	}
	if got := args.GetHash("n"); got != 42 {
		t.Errorf("GetHash(n) = %v, want 42", got)
	}
	if got := args.GetHash("missing"); got != nil {
		t.Errorf("GetHash(missing) = %v, want nil", got)
	}
	args.HashArgs = nil
	if got := args.GetHash("k"); got != nil {
		t.Errorf("GetHash(nil HashArgs) = %v, want nil", got)
	}
}

func TestHelperArgs_GetHashString(t *testing.T) {
	args := HelperArgs{HashArgs: map[string]any{"fmt": "2006-01-02", "n": 99}}
	if got := args.GetHashString("fmt"); got != "2006-01-02" {
		t.Errorf("GetHashString(fmt) = %q, want 2006-01-02", got)
	}
	if got := args.GetHashString("n"); got != "99" {
		t.Errorf("GetHashString(n) = %q, want 99", got)
	}
	if got := args.GetHashString("missing"); got != "" {
		t.Errorf("GetHashString(missing) = %q, want empty", got)
	}
}

func TestHelperArgs_GetHashNumber(t *testing.T) {
	args := HelperArgs{HashArgs: map[string]any{"precision": 2, "sep": "x"}}
	n, err := args.GetHashNumber("precision")
	if err != nil {
		t.Fatalf("GetHashNumber(precision): %v", err)
	}
	if n != 2 {
		t.Errorf("GetHashNumber(precision) = %v, want 2", n)
	}
	_, err = args.GetHashNumber("sep")
	if err == nil {
		t.Error("GetHashNumber(sep) expected error")
	}
	n, err = args.GetHashNumber("missing")
	if err != nil || n != 0 {
		t.Errorf("GetHashNumber(missing) = %v, %v; want 0, nil", n, err)
	}
}
