package runtime

import "testing"

func TestMergePartialContext(t *testing.T) {
	base := map[string]any{"title": "Hi", "x": 1}
	add := map[string]any{"note": "thanks", "x": 2}
	out := MergePartialContext(base, add)
	if out["title"] != "Hi" {
		t.Errorf("base key: got %v", out["title"])
	}
	if out["note"] != "thanks" {
		t.Errorf("add key: got %v", out["note"])
	}
	if out["x"] != 2 {
		t.Errorf("add overrides base: got %v", out["x"])
	}
	// nil base
	out2 := MergePartialContext(nil, add)
	if out2["note"] != "thanks" {
		t.Errorf("nil base: got %v", out2["note"])
	}
}

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
