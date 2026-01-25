package runtime

import (
	"reflect"
	"strconv"
	"strings"
)

// Context holds the current data scope for rendering.
type Context struct {
	Data   any
	parent *Context
	engine *Engine
}

// NewContext creates a new rendering context.
func NewContext(data any, engine *Engine) *Context {
	return &Context{Data: data, engine: engine}
}

// WithData creates a child context with new data.
func (c *Context) WithData(data any) *Context {
	if c == nil {
		return &Context{Data: data}
	}
	return &Context{Data: data, parent: c, engine: c.engine}
}

// Engine returns the associated engine.
func (c *Context) Engine() *Engine {
	if c == nil {
		return nil
	}
	return c.engine
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
	if strings.HasPrefix(path, "@") {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var ok bool
	val := ctx.Data
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
