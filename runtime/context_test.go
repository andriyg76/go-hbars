package runtime

import (
	"reflect"
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
	if got, ok := ResolvePath(ctx, "."); !ok || !reflect.DeepEqual(got, data) {
		t.Fatalf("ResolvePath . = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(ctx, "this"); !ok || !reflect.DeepEqual(got, data) {
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

func TestResolvePathDataVarsAndLocals(t *testing.T) {
	root := NewContext(map[string]any{
		"rootName": "root",
	})
	child := root.WithScope(
		map[string]any{"name": "child"},
		map[string]any{"local": "value"},
		map[string]any{"index": 3, "key": "k"},
	)
	grand := child.WithScope(map[string]any{"name": "grand"}, nil, nil)

	if got, ok := ResolvePath(grand, "../name"); !ok || got != "child" {
		t.Fatalf("ResolvePath ../name = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(grand, "local"); !ok || got != "value" {
		t.Fatalf("ResolvePath local = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(grand, "@index"); !ok || got != 3 {
		t.Fatalf("ResolvePath @index = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(grand, "@key"); !ok || got != "k" {
		t.Fatalf("ResolvePath @key = (%v, %v)", got, ok)
	}
	if got, ok := ResolvePath(grand, "@root.rootName"); !ok || got != "root" {
		t.Fatalf("ResolvePath @root.rootName = (%v, %v)", got, ok)
	}
}

func TestResolvePathParsed(t *testing.T) {
	rootData := map[string]any{
		"rootName": "root",
		"user": map[string]any{
			"name": "Ada",
		},
	}
	root := NewContext(rootData)
	child := root.WithScope(
		map[string]any{"name": "child"},
		map[string]any{"local": "value"},
		map[string]any{"index": 3, "key": "k"},
	)
	grand := child.WithScope(map[string]any{"name": "grand"}, nil, nil)

	tests := []struct {
		name string
		ctx  *Context
		path ParsedPath
		want any
		ok   bool
	}{
		{
			name: "simple path",
			ctx:  root,
			path: ParsedPath{Parts: []string{"user", "name"}},
			want: "Ada",
			ok:   true,
		},
		{
			name: "current",
			ctx:  root,
			path: ParsedPath{Current: true},
			want: rootData,
			ok:   true,
		},
		{
			name: "parent current",
			ctx:  grand,
			path: ParsedPath{Up: 1, Current: true},
			want: map[string]any{"name": "child"},
			ok:   true,
		},
		{
			name: "locals",
			ctx:  grand,
			path: ParsedPath{Parts: []string{"local"}},
			want: "value",
			ok:   true,
		},
		{
			name: "data var",
			ctx:  grand,
			path: ParsedPath{Data: true, Parts: []string{"index"}},
			want: 3,
			ok:   true,
		},
		{
			name: "root data var",
			ctx:  grand,
			path: ParsedPath{Data: true, Parts: []string{"root", "rootName"}},
			want: "root",
			ok:   true,
		},
		{
			name: "missing",
			ctx:  root,
			path: ParsedPath{Parts: []string{"missing", "path"}},
			want: nil,
			ok:   false,
		},
	}

	for _, tc := range tests {
		got, ok := ResolvePathParsed(tc.ctx, tc.path)
		if ok != tc.ok {
			t.Fatalf("ResolvePathParsed %s ok=%v, want %v", tc.name, ok, tc.ok)
		}
		if ok && !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("ResolvePathParsed %s = %v, want %v", tc.name, got, tc.want)
		}
	}
}
