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

Render a template:

```go
out, err := templates.RenderMain(data)
```

Register helpers:

```go
templates.RegisterHelper("upper", func(ctx *runtime.Context, args []any) (any, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToUpper(runtime.Stringify(args[0])), nil
})
```

## Template support (current)

- `{{var}}` (HTML-escaped)
- `{{{var}}}` and `{{& var}}` (raw)
- Inline helpers: `{{helper arg1 arg2}}`
- Partials: `{{> partialName}}`

Blocks (`{{#if}}`, `{{#each}}`) and hash arguments are not implemented yet.
