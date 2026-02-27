# Template API Reference

This document describes the runtime API for working with compiled Handlebars templates.

## Basic Usage

After compiling templates with `hbc`, you get generated functions for each template. The compiler emits **typed context** types (e.g. `MainContext`) inferred from your template expressions:

```go
import "github.com/your/project/templates"

// Render to a writer (data must satisfy the template's context type, e.g. MainContext)
var b strings.Builder
if err := templates.RenderMain(&b, data); err != nil {
    // handle error
}
out := b.String()

// Or use the string wrapper. For map data, use MainContextFromMap(data).
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

## Generated Functions

For each template file (e.g. `main.hbs`), the compiler generates:

1. **Internal render function**: `renderMain(data MainContext, w io.Writer, root any) error` (used by partials; `root` is the caller’s root context for `@root`)
2. **Public render function**: `RenderMain(w io.Writer, data MainContext) error`
3. **String wrapper**: `RenderMainString(data MainContext) (string, error)`

The context type (e.g. `MainContext`) is an interface inferred from paths used in the template; you can pass a struct or `map[string]any` that provides the required fields.

## Runtime Package

The `runtime` package provides types and utilities used by generated code and by custom helpers.

### Output

```go
// WriteEscaped writes an escaped value into the writer
runtime.WriteEscaped(w, value)

// WriteRaw writes a raw value into the writer
runtime.WriteRaw(w, value)

// Stringify converts a value to its string representation
str := runtime.Stringify(value)
```

### Helper Arguments

Helpers receive a single argument of type `runtime.HelperArgs`:

```go
type HelperArgs struct {
    HashArgs   map[string]any  // named (hash) arguments
    Args       []any           // positional arguments
    BlockFn    func() error    // when IsBlock: renders the main block (writer captured in closure); nil otherwise
    InverseFn  func() error    // when IsBlock and else: renders the else block; may be nil
    IsBlock    bool            // true for block helper invocations
}
```

### Truthiness

```go
// IsTruthy checks if a value is truthy
if runtime.IsTruthy(value) {
    // ...
}
```

### Safe Strings

```go
// SafeString marks a value as pre-escaped HTML
safe := runtime.SafeString("<b>bold</b>")
```

### Context and partials

```go
// LookupPath returns the value at a dot-separated path from root (e.g. "title", "user.name").
// Root can be map[string]any or implement Raw() any returning a map.
// Used by generated code for @root.xxx inside partials when root comes from another template.
val := runtime.LookupPath(root, "title")
```

## Helper Functions

All helpers use the same signature (simple and block):

```go
func MyHelper(args runtime.HelperArgs) (any, error)
```

Arguments are **resolved by the compiler** before being passed; you receive `HelperArgs` with `Args` (positional) and `HashArgs` (named). For block invocations the return value is ignored; call `args.BlockFn()` and `args.InverseFn()` to render block content (the writer is captured in the closure) (they return `error`).

### Accessing Arguments

```go
func MyHelper(args runtime.HelperArgs) (any, error) {
    // Positional arguments (already evaluated)
    if len(args.Args) == 0 {
        return nil, fmt.Errorf("missing argument")
    }
    firstArg := args.Args[0]
    
    // Hash arguments (key=value pairs)
    if args.HashArgs != nil {
        value := args.HashArgs["key"]
    }
    
    return result, nil
}
```

### Block Helpers

When a helper is used as a block (`{{#name}}...{{/name}}`), `args.IsBlock` is true and `args.Writer` is the template output writer. Call `args.BlockFn()` or `args.InverseFn()` with no arguments; each is a closure that already captures the template writer. They return `error`. The block is only rendered when you call them.

```go
func MyBlockHelper(args runtime.HelperArgs) (any, error) {
    if !args.IsBlock {
        return nil, nil
    }
    if args.BlockFn != nil {
        if err := args.BlockFn(); err != nil {
            return nil, err
        }
    }
    if args.InverseFn != nil {
        if err := args.InverseFn(); err != nil {
            return nil, err
        }
    }
    return nil, nil
}
```

## Partials

Partials are automatically registered in the generated code:

```go
// partials map (internal): template name -> func(data any, w io.Writer) error
partials["header"](data, w)
```

Templates use them via `{{> header}}` or `{{> (lookup ...) }}`.

When **multiple root templates** include the same partial (e.g. `{{> menu}}`), the compiler emits a **single canonical context type** for that partial (e.g. `MenuContext`). All such pages use that type in their method signatures; no primary-embedding types are emitted. Root contexts that include a single layout embed that layout’s context and only add template-specific methods. See [Compiled template file → Canonical partial context](compiled-templates.md#canonical-partial-context).

## Data Types

### Context Data

The context data for a template satisfies the generated context interface (e.g. `MainContext`). When multiple templates share a partial, that partial has one **canonical** context type (e.g. `MenuContext`), so you can pass data that implements that interface from any page. In practice you can pass:

- Maps (`map[string]any`)
- Structs (with exported fields or JSON tags)
- The compiler also generates `XxxContextFromMap` constructors to build context from `map[string]any`.

### Hash Arguments

Hash arguments are available as `args.HashArgs` (type `map[string]any`) on `HelperArgs`.

## Error Handling

All render functions return errors. Common error scenarios:

- Missing template or partial (compile-time error)
- Missing helper (compile-time error)
- Runtime errors in helpers
- Invalid data types
- Helper runtime errors

Always check errors. When using a `map[string]any` (e.g. from JSON), use the generated `XxxContextFromMap` so data satisfies the context type:

```go
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
if err != nil {
    log.Fatal(err)
}
```

## Performance Considerations

- Templates are compiled to Go code, so execution is fast
- No runtime template parsing
- Context types are resolved at compile time; helpers receive pre-evaluated arguments

## Examples

### Simple Template Rendering

```go
data := map[string]any{
    "title": "Hello",
    "user": map[string]any{
        "name": "Alice",
    },
}
// If your template uses these paths, the generated MainContext will allow map or struct.
// Use MainContextFromMap(data) if the compiler generated it, or pass a struct.
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

### Custom Helper

```go
func FormatCurrency(args runtime.HelperArgs) (any, error) {
    if len(args.Args) == 0 {
        return "", nil
    }
    
    amount := runtime.Stringify(args.Args[0])
    symbol := "$"
    if args.HashArgs != nil {
        if s, ok := args.HashArgs["symbol"].(string); ok {
            symbol = s
        }
    }
    
    return fmt.Sprintf("%s%s", symbol, amount), nil
}
```

### Block Helper

```go
func IfHelper(args runtime.HelperArgs) (any, error) {
    if !args.IsBlock || len(args.Args) < 1 {
        return nil, fmt.Errorf("if requires a condition and block")
    }
    condition := args.Args[0]
    if ok, _ := runtime.IsTruthy(condition); ok {
        if args.BlockFn != nil {
            return nil, args.BlockFn()
        }
    } else if args.InverseFn != nil {
        return nil, args.InverseFn()
    }
    return nil, nil
}
```

Note: the built-in `if`/`unless`/`each`/`with` are implemented by the compiler; the above illustrates the runtime API for custom block helpers using `HelperArgs.BlockFn()` and `InverseFn()` (writer is captured in the closure).

## See also

- [Compiled template file](compiled-templates.md) — What the compiler generates (context types, RenderXxx, FromMap).
- [Template Syntax](syntax.md) — Handlebars expressions and blocks.
- [Built-in Helpers](helpers.md) — Available helpers and how to register custom ones.
