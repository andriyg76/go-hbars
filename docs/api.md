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

```go
// HashArg extracts hash arguments from helper arguments
hash, ok := runtime.HashArg(args)

// GetBlockOptions extracts block options from helper arguments (for block helpers)
opts, ok := runtime.GetBlockOptions(args)
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

Simple helpers (non-block) must match this signature:

```go
func MyHelper(args []any) (any, error)
```

Arguments are **resolved by the compiler** before being passed; you receive the evaluated values. No context is passed—the compiler bakes in the needed lookups.

### Accessing Arguments

```go
func MyHelper(args []any) (any, error) {
    // Positional arguments (already evaluated)
    if len(args) == 0 {
        return nil, fmt.Errorf("missing argument")
    }
    firstArg := args[0]
    
    // Hash arguments (key=value pairs)
    hash, ok := runtime.HashArg(args)
    if ok {
        value := hash["key"]
    }
    
    return result, nil
}
```

### Block Helpers

Block helpers are invoked by the compiler with a single argument: the full `args` slice, whose **last element** is the block options. Use signature `func(args []any) error` and extract options with `runtime.GetBlockOptions(args)`:

```go
func MyBlockHelper(args []any) error {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return fmt.Errorf("block helper did not receive BlockOptions")
    }
    // Render the main block (opts.Fn(w) requires w from caller scope)
    if opts.Fn != nil {
        if err := opts.Fn(w); err != nil {
            return err
        }
    }
    // Render the inverse/else block
    if opts.Inverse != nil {
        if err := opts.Inverse(w); err != nil {
            return err
        }
    }
    return nil
}
```

`BlockOptions` provides:

```go
type BlockOptions struct {
    Fn      func(io.Writer) error  // main block body
    Inverse func(io.Writer) error   // else block body
}
```

The runtime also defines `BlockHelper` as `func(args []any, options BlockOptions) error` for use when you call a block helper manually with two arguments. When invoked from generated code, only `args` is passed (with options as the last element).

## Partials

Partials are automatically registered in the generated code:

```go
// partials map (internal): template name -> func(data any, w io.Writer) error
partials["header"](data, w)
```

Templates use them via `{{> header}}` or `{{> (lookup ...) }}`.

## Data Types

### Context Data

The context data for a template satisfies the generated context interface (e.g. `MainContext`). In practice you can pass:

- Maps (`map[string]any`)
- Structs (with exported fields or JSON tags)
- The compiler also generates `XxxContextFromMap` constructors to build context from `map[string]any`.

### Hash Arguments

Hash arguments are passed as `runtime.Hash`:

```go
type Hash map[string]any
```

### Block Options

```go
type BlockOptions struct {
    Fn      func(io.Writer) error
    Inverse func(io.Writer) error
}
```

## Error Handling

All render functions return errors. Common error scenarios:

- Missing template or partial (compile-time error)
- Missing helper (compile-time error)
- Runtime errors in helpers
- Invalid data types
- Block helper did not receive BlockOptions

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
func FormatCurrency(args []any) (any, error) {
    if len(args) == 0 {
        return "", nil
    }
    
    amount := runtime.Stringify(args[0])
    hash, _ := runtime.HashArg(args)
    
    symbol := "$"
    if hash != nil {
        if s, ok := hash["symbol"].(string); ok {
            symbol = s
        }
    }
    
    return fmt.Sprintf("%s%s", symbol, amount), nil
}
```

### Block Helper

```go
func IfHelper(args []any) error {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return fmt.Errorf("if: no block options")
    }
    if len(args) < 1 {
        return fmt.Errorf("if requires a condition")
    }
    condition := args[0]
    if runtime.IsTruthy(condition) {
        if opts.Fn != nil {
            return opts.Fn(w) // w is the template output writer (in scope in generated code)
        }
    } else if opts.Inverse != nil {
        return opts.Inverse(w)
    }
    return nil
}
```

Note: the built-in `if`/`unless`/`each`/`with` are implemented by the compiler; the above illustrates the runtime API for custom block helpers. When the compiler invokes a block helper it calls `helper(args)`; the writer `w` is in scope in the generated render function. Custom helpers that are called from generated code and need to render the block must receive or capture the writer (e.g. via an adapter).

## See also

- [Compiled template file](compiled-templates.md) — What the compiler generates (context types, RenderXxx, FromMap).
- [Template Syntax](syntax.md) — Handlebars expressions and blocks.
- [Built-in Helpers](helpers.md) — Available helpers and how to register custom ones.
