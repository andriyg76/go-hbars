package runtime

import (
	"fmt"
	"strings"
	"sync"
)

// Helper is a user-defined function invoked from a template.
type Helper func(ctx *Context, args []any) (any, error)

// Template is a compiled template function.
type Template func(ctx *Context, b *strings.Builder) error

// Engine holds registered helpers and partials.
type Engine struct {
	mu       sync.RWMutex
	helpers  map[string]Helper
	partials map[string]Template
}

// NewEngine creates a new runtime engine.
func NewEngine() *Engine {
	return &Engine{
		helpers:  make(map[string]Helper),
		partials: make(map[string]Template),
	}
}

// RegisterHelper registers or replaces a helper.
func (e *Engine) RegisterHelper(name string, h Helper) {
	if e == nil || name == "" || h == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.helpers[name] = h
}

// RegisterPartial registers or replaces a partial template.
func (e *Engine) RegisterPartial(name string, t Template) {
	if e == nil || name == "" || t == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.partials[name] = t
}

// HasHelper reports whether a helper is registered.
func (e *Engine) HasHelper(name string) bool {
	if e == nil || name == "" {
		return false
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.helpers[name]
	return ok
}

// CallHelper invokes a helper by name.
func (e *Engine) CallHelper(name string, ctx *Context, args []any) (any, error) {
	if e == nil {
		return nil, fmt.Errorf("runtime: helper %q: engine is nil", name)
	}
	e.mu.RLock()
	h := e.helpers[name]
	e.mu.RUnlock()
	if h == nil {
		return nil, fmt.Errorf("runtime: helper %q is not registered", name)
	}
	return h(ctx, args)
}

// RenderPartial renders a registered partial into the builder.
func (e *Engine) RenderPartial(name string, ctx *Context, data any, b *strings.Builder) error {
	if e == nil {
		return fmt.Errorf("runtime: partial %q: engine is nil", name)
	}
	e.mu.RLock()
	t := e.partials[name]
	e.mu.RUnlock()
	if t == nil {
		return fmt.Errorf("runtime: partial %q is not registered", name)
	}
	if data != nil {
		ctx = ctx.WithData(data)
	}
	return t(ctx, b)
}
