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

## Template syntax (supported)

### Values and helpers

- `{{var}}` (HTML-escaped)
- `{{{var}}}` and `{{& var}}` (raw)
- Inline helpers: `{{helper arg1 arg2}}`
- Comments: `{{! comment}}` and `{{!-- block --}}`

### Partials

- `{{> partialName}}`
- `{{> partialName contextExpr}}` (render with a different context)

### Block helpers

- `{{#if expr}}...{{else}}...{{/if}}`
- `{{#unless expr}}...{{/unless}}`
- `{{#with expr}}...{{else}}...{{/with}}`
- `{{#each expr}}...{{else}}...{{/each}}`

`expr` is a single argument (path or literal). The current context is updated
inside `each` and `with`, so `{{this}}` / `{{.}}` refer to the item/object.

### Truthiness and iteration

`if`, `unless`, and `with` treat values as false when they are `nil`, `false`,
`0`, `""`, or empty arrays/slices/maps. Everything else is truthy.

`each` iterates over slices, arrays, and maps with string keys (keys are sorted
for deterministic output). Empty or non-iterable values render the `{{else}}`
branch when present.

## Template examples

Simple values and helpers:

```
Hello {{user.name}}!
{{upper user.role}}
```

Conditional rendering:

```
{{#if user.active}}
  Welcome back, {{user.name}}!
{{else}}
  Please activate your account.
{{/if}}
```

Iteration with an else fallback:

```
{{#each users}}
  <li>{{name}}</li>
{{else}}
  <li>No users yet.</li>
{{/each}}
```

Nested context with `with` and partials:

```
{{#with user}}
  {{> userCard this}}
{{/with}}
```

## Suggested custom helpers

You can implement common helpers as regular Go functions and map them with
`-helper name=Ident`. Examples:

- `eq`, `ne`, `and`, `or`, `not`
- `default` (fallback when value is empty)
- `join`, `split`
- `json` (serialize a value)
- `formatDate`, `formatNumber`, `pluralize`

## Not implemented yet

- Hash arguments: `{{helper arg key=value}}`
- Subexpressions: `{{helper (other arg)}}`
- Block parameters and `@index` / `@key`
- Whitespace control (`~`)
