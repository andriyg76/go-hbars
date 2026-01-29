# go-hbars

Handlebars template compiler for Go.

## Status

Early MVP. Current focus is on a minimal core with HTML escaping, helpers, and partials.

## Quick Start

Install the compiler:

```bash
go install ./cmd/hbc
```

Generate Go code from templates:

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates
```

Render a template:

```go
import "github.com/your/project/templates"

var b strings.Builder
if err := templates.RenderMain(&b, data); err != nil {
    // handle error
}
out := b.String()
```

Or use the string wrapper:

```go
out, err := templates.RenderMainString(data)
```

## Documentation

- **[init: create or add to a project](docs/init.md)** - Scaffold a new go-hbars project or add templates to an existing module
- **[Template Syntax](docs/syntax.md)** - Complete Handlebars syntax reference
- **[Custom Extensions](docs/extensions.md)** - includeZero
- **[Built-in Helpers](docs/helpers.md)** - Available helpers and how to use them
- **[Processor & Server](docs/processor-server.md)** - CLI tools for static site generation
- **[Embedded API](docs/embedded.md)** - Embedding processor and server in your applications
- **[Template API](docs/api.md)** - Runtime API for compiled templates

## Features

- ✅ Compile-time template compilation (no runtime parsing)
- ✅ Full Handlebars syntax support
- ✅ Comprehensive built-in helpers library
- ✅ Custom helpers support
- ✅ Partials and dynamic partials
- ✅ Block helpers with parameters
- ✅ Static site generation
- ✅ Semi-static web server
- ✅ Bootstrap code generation for quick setup
- ✅ `go:generate` integration

## Installation

```bash
go install github.com/andriyg76/go-hbars/cmd/hbc@latest
```

## Basic Usage

### Compile Templates

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -bootstrap
```

### Use Compiled Templates

```go
import "github.com/your/project/templates"

// Render to string
html, err := templates.RenderMainString(data)

// Or render to writer
err := templates.RenderMain(writer, data)
```

### Quick Server (with bootstrap)

```go
import "github.com/your/project/templates"

srv, err := templates.NewQuickServer()
if err != nil {
    log.Fatal(err)
}
log.Fatal(srv.Start())
```

### Quick Processor (with bootstrap)

```go
import "github.com/your/project/templates"

proc, err := templates.NewQuickProcessor()
if err != nil {
    log.Fatal(err)
}
if err := proc.Process(); err != nil {
    log.Fatal(err)
}
```

## Implementation Status

All core Handlebars syntax features are now implemented:
- ✅ Custom block helpers
- ✅ Block params for `if`/`unless`
- ✅ `else if` shorthand

## Compatibility

See `examples/compat` for a small template set that exercises hash arguments, subexpressions, data variables, parent paths, block params, dynamic partials, whitespace control, and raw blocks.

## License

See [LICENSE](LICENSE) file.
