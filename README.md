# go-hbars

Handlebars template compiler for Go.

## Status

Early MVP. Current focus is on a minimal core with HTML escaping, helpers, and partials.

## Usage

Install the compiler:

```
go install ./cmd/hbc
```

Generate Go code from templates (example):

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates
```

Render a template to a writer:

```go
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

Register helpers by mapping names to functions at compile time (no runtime registry):

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=Upper

func Upper(ctx *runtime.Context, args []any) (any, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToUpper(runtime.Stringify(args[0])), nil
}
```

To import helpers from another package:

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=github.com/you/helpers:Upper
```

Multiple helpers can be passed by repeating the flag:

```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -helper upper=Upper -helper lower=github.com/you/helpers:Lower
```

## Partials

Partials are compiled from the same input set. The partial name is the template
file base name (without extension). A missing partial or helper is a compile-time
error.

## Template support (current)

- `{{var}}` (HTML-escaped)
- `{{{var}}}` and `{{& var}}` (raw)
- Inline helpers: `{{helper arg1 arg2}}`
- Partials: `{{> partialName}}`

Blocks (`{{#if}}`, `{{#each}}`) and hash arguments are not implemented yet.
