package runtime

import (
	"testing"
)

type profile struct {
	Name     string
	FullName string `json:"full_name"`
}

func TestResolvePath(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Ada",
		},
		"profile": profile{
			Name:     "Grace",
			FullName: "Grace Hopper",
		},
		"items": []string{"a", "b", "c"},
	}
	ctx := NewContext(data)

	if got, ok := ResolvePath(ctx, "user.name"); !ok || got != "Ada" {
		t.Fatalf("ResolvePath user.name = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "profile.Name"); !ok || got != "Grace" {
		t.Fatalf("ResolvePath profile.Name = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "profile.full_name"); !ok || got != "Grace Hopper" {
		t.Fatalf("ResolvePath profile.full_name = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "items.1"); !ok || got != "b" {
		t.Fatalf("ResolvePath items.1 = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "."); !ok || got != data {
		t.Fatalf("ResolvePath . = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "this"); !ok || got != data {
		t.Fatalf("ResolvePath this = (%v, %v)", got, ok)
	}
	if _, ok := ResolvePath(ctx, "missing.path"); ok {
		t.Fatalf("ResolvePath missing.path should be false")
	}
}

func TestResolvePathParentFallback(t *testing.T) {
	parent := NewContext(map[string]any{
		"user": map[string]any{"name": "Parent"},
	})
	child := parent.WithData(map[string]any{
		"value": "child",
	})
	if got, ok := ResolvePath(child, "user.name"); !ok || got != "Parent" {
		t.Fatalf("ResolvePath user.name from parent = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(child, "value"); !ok || got != "child" {
		t.Fatalf("ResolvePath value = (%v, %v)", got, ok)
	}
}
