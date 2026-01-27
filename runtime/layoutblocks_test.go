package runtime

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestBlock_EmptyContext(t *testing.T) {
	// {{#block "x"}}default{{/block}} with no Blocks → runs default
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	var defaultRan bool
	opts := BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			defaultRan = true
			_, _ = io.WriteString(w, "default")
			return nil
		},
	}
	err := Block(ctx, []any{"x", opts})
	if err != nil {
		t.Fatalf("Block: %v", err)
	}
	if !defaultRan {
		t.Fatal("expected default Fn to run")
	}
	if got := ctx.Output.(*bytes.Buffer).String(); got != "default" {
		t.Fatalf("got %q", got)
	}
}

func TestBlock_Override(t *testing.T) {
	// {{#block "x"}}default{{/block}} with Blocks["x"] set → writes override
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	ctx.Blocks = map[string]string{"x": "override"}
	opts := BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			t.Fatal("default should not run when override exists")
			return nil
		},
	}
	err := Block(ctx, []any{"x", opts})
	if err != nil {
		t.Fatalf("Block: %v", err)
	}
	if got := ctx.Output.(*bytes.Buffer).String(); got != "override" {
		t.Fatalf("got %q", got)
	}
}

func TestPartial_StoresInBlocks(t *testing.T) {
	// {{#partial "x"}}body{{/partial}} → body in ctx.Blocks["x"], no output
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	opts := BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			_, _ = io.WriteString(w, "body")
			return nil
		},
	}
	err := Partial(ctx, []any{"x", opts})
	if err != nil {
		t.Fatalf("Partial: %v", err)
	}
	if ctx.Blocks == nil {
		t.Fatal("Blocks should be initialized")
	}
	if ctx.Blocks["x"] != "body" {
		t.Fatalf("Blocks[\"x\"] = %q", ctx.Blocks["x"])
	}
	if ctx.Output.(*bytes.Buffer).Len() != 0 {
		t.Fatal("Partial must not write to Output")
	}
}

func TestPartialThenBlock_DirectionA(t *testing.T) {
	// Page does {{#partial "header"}}Custom{{/partial}} then layout uses {{#block "header"}}Default{{/block}}
	ctx := NewContext(nil)
	var out bytes.Buffer
	ctx.Output = &out

	// Simulate partial (page defines slot)
	err := Partial(ctx, []any{"header", BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			_, _ = io.WriteString(w, "Custom")
			return nil
		},
	}})
	if err != nil {
		t.Fatalf("Partial: %v", err)
	}

	// Simulate block (layout renders slot)
	err = Block(ctx, []any{"header", BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			_, _ = io.WriteString(w, "Default")
			return nil
		},
	}})
	if err != nil {
		t.Fatalf("Block: %v", err)
	}

	if got := out.String(); got != "Custom" {
		t.Fatalf("expected Custom, got %q", got)
	}
}

func TestBlock_WithScopePreservesBlocks(t *testing.T) {
	// Child context from WithScope must see parent Blocks (same map)
	ctx := NewContext(nil)
	ctx.Blocks = map[string]string{"x": "from parent"}
	ctx.Output = &bytes.Buffer{}
	child := ctx.WithScope("inner", nil, nil)
	child.Blocks["_"] = "shared"
	if ctx.Blocks["_"] != "shared" {
		t.Fatal("WithScope must preserve Blocks (same map)")
	}
	delete(child.Blocks, "_")
	// Block in child should see parent Blocks
	err := Block(child, []any{"x", BlockOptions{Fn: func(*Context, io.Writer) error { return nil }}})
	if err != nil {
		t.Fatalf("Block: %v", err)
	}
	if got := ctx.Output.(*bytes.Buffer).String(); got != "from parent" {
		t.Fatalf("got %q", got)
	}
}

func TestPartial_StringifyName(t *testing.T) {
	// Name can be non-string (compiler uses Stringify)
	ctx := NewContext(nil)
	opts := BlockOptions{Fn: func(*Context, io.Writer) error { return nil }}
	err := Partial(ctx, []any{123, opts})
	if err != nil {
		t.Fatalf("Partial: %v", err)
	}
	if ctx.Blocks["123"] != "" {
		t.Fatalf("Blocks[\"123\"] = %q", ctx.Blocks["123"])
	}
}

func TestBlock_EmptyName(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	opts := BlockOptions{Fn: func(*Context, io.Writer) error { return nil }}
	_ = Block(ctx, []any{"", opts})
	// No crash, default would run if name empty and no override
}

func TestBlock_NilOutput(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Output = nil
	opts := BlockOptions{Fn: func(*Context, io.Writer) error { return nil }}
	err := Block(ctx, []any{"x", opts})
	if err != nil {
		t.Fatalf("Block with nil Output: %v", err)
	}
}

func TestPartial_MultipleSlots(t *testing.T) {
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	for _, name := range []string{"a", "b"} {
		body := strings.ToUpper(name)
		err := Partial(ctx, []any{name, BlockOptions{
			Fn: func(_ *Context, w io.Writer) error {
				_, _ = io.WriteString(w, body)
				return nil
			},
		}})
		if err != nil {
			t.Fatalf("Partial %q: %v", name, err)
		}
	}
	if ctx.Blocks["a"] != "A" || ctx.Blocks["b"] != "B" {
		t.Fatalf("Blocks = %v", ctx.Blocks)
	}
}

func TestBlock_LazySlots(t *testing.T) {
	// When ctx.LazySlots is set, Block writes placeholder and records; does not run default now
	recorder := NewLazySlotsRecorder()
	ctx := NewContext(nil)
	ctx.Output = &bytes.Buffer{}
	ctx.LazySlots = recorder
	opts := BlockOptions{
		Fn: func(_ *Context, w io.Writer) error {
			t.Fatal("default must not run when LazySlots is set")
			return nil
		},
	}
	err := Block(ctx, []any{"header", opts})
	if err != nil {
		t.Fatalf("Block: %v", err)
	}
	out := ctx.Output.(*bytes.Buffer).String()
	if out == "" || !strings.Contains(out, "SLOT") {
		t.Fatalf("expected placeholder in output, got %q", out)
	}
	slots := recorder.Slots()
	if len(slots) != 1 || slots[0].Name != "header" {
		t.Fatalf("expected one slot named header, got %v", slots)
	}
}

func TestResolveLazySlots(t *testing.T) {
	ctx := NewContext(nil)
	recorder := NewLazySlotsRecorder()
	ph := recorder.Record("h", func(_ *Context, w io.Writer) error {
		_, _ = io.WriteString(w, "default")
		return nil
	})
	buf := &bytes.Buffer{}
	buf.WriteString("prefix" + ph + "suffix")
	blocks := map[string]string{"h": "override"}
	err := ResolveLazySlots(buf, recorder.Slots(), blocks, ctx)
	if err != nil {
		t.Fatalf("ResolveLazySlots: %v", err)
	}
	if got := buf.String(); got != "prefixoverridesuffix" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveLazySlots_UseDefault(t *testing.T) {
	ctx := NewContext(nil)
	recorder := NewLazySlotsRecorder()
	ph := recorder.Record("h", func(_ *Context, w io.Writer) error {
		_, _ = io.WriteString(w, "default")
		return nil
	})
	buf := &bytes.Buffer{}
	buf.WriteString(ph)
	err := ResolveLazySlots(buf, recorder.Slots(), nil, ctx)
	if err != nil {
		t.Fatalf("ResolveLazySlots: %v", err)
	}
	if got := buf.String(); got != "default" {
		t.Fatalf("got %q", got)
	}
}
