package runtime

import "testing"

func TestBlocks_GetSet(t *testing.T) {
	b := NewBlocks()
	if b == nil {
		t.Fatal("NewBlocks() returned nil")
	}
	if got, ok := b.Get("x"); ok || got != "" {
		t.Errorf("Get empty: got %q, ok=%v", got, ok)
	}
	b.Set("x", "content")
	got, ok := b.Get("x")
	if !ok || got != "content" {
		t.Errorf("Get after Set: got %q, ok=%v", got, ok)
	}
	b.Set("y", "other")
	got2, ok2 := b.Get("y")
	if !ok2 || got2 != "other" {
		t.Errorf("Get y: got %q, ok=%v", got2, ok2)
	}
}

func TestBlocks_Get_nil(t *testing.T) {
	var b *Blocks
	got, ok := b.Get("x")
	if ok || got != "" {
		t.Errorf("Get on nil Blocks: got %q, ok=%v", got, ok)
	}
	b.Set("x", "y") // no-op, should not panic
}
