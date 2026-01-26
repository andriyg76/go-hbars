package runtime

import "testing"

func TestHashArg(t *testing.T) {
	if _, ok := HashArg(nil); ok {
		t.Fatalf("expected no hash for nil args")
	}
	args := []any{"x", Hash{"a": 1}}
	hash, ok := HashArg(args)
	if !ok {
		t.Fatalf("expected hash arg")
	}
	if hash["a"].(int) != 1 {
		t.Fatalf("unexpected hash value: %v", hash)
	}
}
