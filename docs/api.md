# Template API Reference

This document describes the runtime API for working with compiled Handlebars templates.

## Basic Usage

After compiling templates with `hbc`, you get generated functions for each template:

```go
import "github.com/your/project/templates"

// Render to a writer
var b strings.Builder
if err := templates.RenderMain(&b, data); err != nil {
    // handle error
}
out := b.String()

// Or use the string wrapper
out, err := templates.RenderMainString(data)
```

## Generated Functions

For each template file (e.g., `main.hbs`), the compiler generates:

1. **Internal render function**: `renderMain(ctx *runtime.Context, w io.Writer) error`
2. **Public render function**: `RenderMain(w io.Writer, data any) error`
3. **String wrapper**: `RenderMainString(data any) (string, error)`

## Runtime Package

The `runtime` package provides the core functionality for template execution.

### Context

```go
// NewContext creates a new rendering context
ctx := runtime.NewContext(data)

// WithData creates a child context with new data
childCtx := ctx.WithData(newData)

// WithScope creates a child context with new data and optional locals/data vars
childCtx := ctx.WithScope(data, locals, dataVars)
```

### Path Resolution

```go
// ResolvePath looks up a dotted path in the current context
value, ok := runtime.ResolvePath(ctx, "user.name")

// ResolvePathParsed resolves a pre-parsed path expression
value, ok := runtime.ResolvePathParsed(ctx, parsedPath)
```

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
// EvalArg evaluates an argument expression
value := runtime.EvalArg(ctx, runtime.ArgPath, "user.name")
value := runtime.EvalArg(ctx, runtime.ArgString, "literal")
value := runtime.EvalArg(ctx, runtime.ArgNumber, "42")

// HashArg extracts hash arguments from helper arguments
hash, ok := runtime.HashArg(args)

// GetBlockOptions extracts block options from helper arguments
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

## Helper Functions

Helper functions must match this signature:

```go
func MyHelper(ctx *runtime.Context, args []any) (any, error)
```

### Accessing Arguments

```go
func MyHelper(ctx *runtime.Context, args []any) (any, error) {
    // Positional arguments
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

```go
func MyBlockHelper(ctx *runtime.Context, args []any) (any, error) {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        // Not used as a block
        return "default", nil
    }
    
    // Render the main block
    if opts.Fn != nil {
        var b strings.Builder
        if err := opts.Fn(ctx, &b); err != nil {
            return nil, err
        }
        return b.String(), nil
    }
    
    // Render the inverse/else block
    if opts.Inverse != nil {
        var b strings.Builder
        if err := opts.Inverse(ctx, &b); err != nil {
            return nil, err
        }
        return b.String(), nil
    }
    
    return "", nil
}
```

## Partials

Partials are automatically registered in the generated code:

```go
// Access partials map (internal)
partials["header"](ctx, w)

// Partials are used automatically in templates via {{> header}}
```

## Data Types

### Context Data

The context data can be any Go type:
- Maps (`map[string]any`)
- Structs (with exported fields or JSON tags)
- Slices/arrays
- Primitives (string, int, float, bool, etc.)

### Hash Arguments

Hash arguments are passed as `runtime.Hash`:

```go
type Hash map[string]any
```

### Block Options

```go
type BlockOptions struct {
    Fn      func(*Context, io.Writer) error
    Inverse func(*Context, io.Writer) error
}
```

## Error Handling

All render functions return errors. Common error scenarios:

- Missing template or partial (compile-time error)
- Missing helper (compile-time error)
- Runtime errors in helpers
- Invalid data types
- Path resolution failures

Always check errors:

```go
out, err := templates.RenderMainString(data)
if err != nil {
    log.Fatal(err)
}
```

## Performance Considerations

- Templates are compiled to Go code, so execution is fast
- No runtime template parsing
- Context creation is lightweight
- Path resolution uses runtime string lookup (`ResolvePath` / `ResolvePathValue`)

## Examples

### Simple Template Rendering

```go
data := map[string]any{
    "title": "Hello",
    "user": map[string]any{
        "name": "Alice",
    },
}

out, err := templates.RenderMainString(data)
```

### Custom Helper

```go
func FormatCurrency(ctx *runtime.Context, args []any) (any, error) {
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
func IfHelper(ctx *runtime.Context, args []any) (any, error) {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return nil, fmt.Errorf("if must be used as block")
    }
    
    condition := args[0]
    if runtime.IsTruthy(condition) {
        if opts.Fn != nil {
            var b strings.Builder
            if err := opts.Fn(ctx, &b); err != nil {
                return nil, err
            }
            return b.String(), nil
        }
    } else {
        if opts.Inverse != nil {
            var b strings.Builder
            if err := opts.Inverse(ctx, &b); err != nil {
                return nil, err
            }
            return b.String(), nil
        }
    }
    
    return "", nil
}
```

