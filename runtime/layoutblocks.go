package runtime

import (
	"bytes"
	"io"
)

// Block implements {{#block "name"}}default{{/block}}. It writes the content
// for name from ctx.Blocks if set; otherwise runs the default block (opts.Fn)
// to ctx.Output. Used by layouts to render slots that the page may override.
//
// When ctx.LazySlots is set (layout-first / Direction B), Block writes a
// placeholder and records the slot instead of resolving now. Call
// ResolveLazySlots after layout and content have run.
func Block(ctx *Context, args []any) error {
	if len(args) == 0 {
		return nil
	}
	name := Stringify(args[0])
	opts, ok := GetBlockOptions(args)
	if !ok || opts.Fn == nil {
		return nil
	}
	if ctx.Output == nil {
		return nil
	}
	if ctx.LazySlots != nil {
		ph := ctx.LazySlots.Record(name, opts.Fn)
		_, err := io.WriteString(ctx.Output, ph)
		return err
	}
	if ctx.Blocks != nil {
		if s, ok := ctx.Blocks[name]; ok && s != "" {
			_, err := io.WriteString(ctx.Output, s)
			return err
		}
	}
	return opts.Fn(ctx, ctx.Output)
}

// Partial implements {{#partial "name"}}body{{/partial}}. It renders the block
// body into a buffer and stores the result in ctx.Blocks[name]. Does not write
// to ctx.Output. Used by pages to define slot content that the layout reads
// via {{#block "name"}}.
func Partial(ctx *Context, args []any) error {
	if len(args) == 0 {
		return nil
	}
	name := Stringify(args[0])
	opts, ok := GetBlockOptions(args)
	if !ok || opts.Fn == nil {
		return nil
	}
	if ctx.Blocks == nil {
		ctx.Blocks = make(map[string]string)
	}
	var buf bytes.Buffer
	if err := opts.Fn(ctx, &buf); err != nil {
		return err
	}
	ctx.Blocks[name] = buf.String()
	return nil
}
