package runtime

import (
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Context holds the current data scope for rendering.
type Context struct {
	Data   any
	parent *Context
	locals map[string]any
	data   map[string]any
	root   any

	// Output is the current writer for this render frame. Set at render start,
	// preserved through WithScope. Used by block helpers (e.g. block) to write.
	Output io.Writer

	// Blocks is the shared layout-blocks map when using partial/block. Passed
	// down through WithScope so layout and content see the same map. Nil when
	// not using layout blocks.
	Blocks map[string]string

	// LazySlots enables layout-first (Direction B) lazy slots: when set, Block()
	// writes a placeholder and records the slot instead of resolving now.
	// Preserved through WithScope. Call ResolveLazySlots after layout + content.
	LazySlots *LazySlotsRecorder
}

// ParsedPath represents a pre-parsed path expression.
type ParsedPath struct {
	Parts   []string
	Up      int
	Data    bool
	Current bool
}

// NewContext creates a new rendering context.
func NewContext(data any) *Context {
	return &Context{Data: data, root: data}
}

// WithData creates a child context with new data.
func (c *Context) WithData(data any) *Context {
	return c.WithScope(data, nil, nil)
}

// WithScope creates a child context with new data and optional locals/data vars.
// Output and Blocks are preserved from the parent so layout partial/block share state.
func (c *Context) WithScope(data any, locals map[string]any, dataVars map[string]any) *Context {
	if c == nil {
		return &Context{Data: data, locals: locals, data: dataVars, root: data}
	}
	return &Context{Data: data, parent: c, locals: locals, data: dataVars, root: c.root, Output: c.Output, Blocks: c.Blocks, LazySlots: c.LazySlots}
}

// ResolvePath looks up a dotted path in the current context.
func ResolvePath(ctx *Context, path string) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	path = strings.TrimSpace(path)
	if path == "" || path == "." || path == "this" {
		return ctx.Data, true
	}
	for path == ".." || strings.HasPrefix(path, "../") {
		if ctx.parent == nil {
			return nil, false
		}
		ctx = ctx.parent
		if path == ".." {
			path = ""
			break
		}
		path = strings.TrimPrefix(path, "../")
	}
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	if path == "" || path == "." || path == "this" {
		return ctx.Data, true
	}
	if strings.HasPrefix(path, "@") {
		return resolveDataVar(ctx, path)
	}
	parts := strings.Split(path, ".")
	if val, ok := resolveLocals(ctx, parts); ok {
		return val, true
	}
	return resolveData(ctx, parts)
}

// ResolvePathParsed resolves a pre-parsed path expression.
func ResolvePathParsed(ctx *Context, path ParsedPath) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	for i := 0; i < path.Up; i++ {
		if ctx.parent == nil {
			return nil, false
		}
		ctx = ctx.parent
	}
	if path.Current {
		return ctx.Data, true
	}
	if path.Data {
		return resolveDataVarParts(ctx, path.Parts)
	}
	if val, ok := resolveLocals(ctx, path.Parts); ok {
		return val, true
	}
	return resolveData(ctx, path.Parts)
}

// ResolvePathValueParsed resolves a pre-parsed path and returns the value only.
func ResolvePathValueParsed(ctx *Context, path ParsedPath) any {
	val, _ := ResolvePathParsed(ctx, path)
	return val
}

func resolveLocals(ctx *Context, parts []string) (any, bool) {
	if len(parts) == 0 {
		return nil, false
	}
	name := parts[0]
	for cur := ctx; cur != nil; cur = cur.parent {
		if cur.locals == nil {
			continue
		}
		val, ok := cur.locals[name]
		if !ok {
			continue
		}
		return resolveParts(val, parts[1:])
	}
	return nil, false
}

func resolveData(ctx *Context, parts []string) (any, bool) {
	for cur := ctx; cur != nil; cur = cur.parent {
		val, ok := resolveParts(cur.Data, parts)
		if ok {
			return val, true
		}
	}
	return nil, false
}

func resolveParts(value any, parts []string) (any, bool) {
	if len(parts) == 0 {
		return value, true
	}
	val := value
	ok := true
	for _, part := range parts {
		if part == "" {
			continue
		}
		val, ok = lookupValue(val, part)
		if !ok {
			return nil, false
		}
	}
	return val, true
}

func resolveDataVar(ctx *Context, path string) (any, bool) {
	path = strings.TrimPrefix(path, "@")
	if path == "" {
		return nil, false
	}
	parts := strings.Split(path, ".")
	return resolveDataVarParts(ctx, parts)
}

func resolveDataVarParts(ctx *Context, parts []string) (any, bool) {
	if len(parts) == 0 {
		return nil, false
	}
	name := parts[0]
	if name == "" {
		return nil, false
	}
	if name == "root" {
		return resolveParts(ctx.root, parts[1:])
	}
	for cur := ctx; cur != nil; cur = cur.parent {
		if cur.data == nil {
			continue
		}
		val, ok := cur.data[name]
		if !ok {
			continue
		}
		return resolveParts(val, parts[1:])
	}
	return nil, false
}

func lookupValue(val any, key string) (any, bool) {
	if val == nil {
		return nil, false
	}
	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		mv := v.MapIndex(reflect.ValueOf(key))
		if !mv.IsValid() {
			return nil, false
		}
		return mv.Interface(), true
	case reflect.Struct:
		if field := v.FieldByName(key); field.IsValid() && field.CanInterface() {
			return field.Interface(), true
		}
		// Try json tag matches (best-effort).
		vt := v.Type()
		for i := 0; i < vt.NumField(); i++ {
			sf := vt.Field(i)
			if sf.PkgPath != "" {
				continue
			}
			tag := sf.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			tagName := strings.Split(tag, ",")[0]
			if tagName == key {
				fv := v.Field(i)
				if fv.IsValid() && fv.CanInterface() {
					return fv.Interface(), true
				}
			}
		}
		return nil, false
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(key)
		if err != nil || idx < 0 || idx >= v.Len() {
			return nil, false
		}
		return v.Index(idx).Interface(), true
	default:
		return nil, false
	}
}
