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

## Template Syntax Guide

### Values and Expressions

**Simple variables:**
```handlebars
{{user.name}}
{{title}}
```
Outputs the value of `user.name` or `title` from the current context, HTML-escaped.

**Raw output (no escaping):**
```handlebars
{{{htmlContent}}}
{{& htmlContent}}
```
Both forms output the value without HTML escaping. Use with caution.

**Current context:**
```handlebars
{{.}}
{{this}}
```
Both refer to the current context object.

**Comments:**
```handlebars
{{! This is a comment}}
{{!-- This is a block comment --}}
```
Comments are removed from output.

### Helpers

**Inline helpers:**
```handlebars
{{upper user.name}}
{{formatDate user.created format="2006-01-02"}}
{{default user.nickname value="(none)"}}
```
Helpers are functions that transform values. They can take positional arguments and hash arguments (key=value pairs).

**Hash arguments:**
```handlebars
{{truncate description length=50 suffix="..."}}
{{formatNumber price precision=2 separator=","}}
```
Hash arguments are passed as the last argument to helpers as a `runtime.Hash` map.

**Subexpressions:**
```handlebars
{{upper (lower title)}}
{{formatDate (lookup user "created") format="2006-01-02"}}
```
Nested helper calls where the result of one helper is passed as an argument to another.

### Partials

**Basic partial:**
```handlebars
{{> header}}
```
Renders the `header` partial template with the current context.

**Partial with different context:**
```handlebars
{{> userCard user}}
```
Renders `userCard` partial with `user` as the context instead of the current context.

**Partial with locals:**
```handlebars
{{> footer note="thanks"}}
```
Renders `footer` partial with `note` available as a local variable.

**Dynamic partial names:**
```handlebars
{{> (lookup . "cardPartial") user}}
```
Uses a helper to determine the partial name at runtime.

**Partial blocks (fallback content):**
```handlebars
{{#> header}}
  <h1>Default Header</h1>
{{/header}}
```
Partial blocks render the partial if it exists, otherwise render the fallback block content. Useful for providing default content when a partial is missing.

### Block Helpers

**Conditional rendering (`if`):**
```handlebars
{{#if user.active}}
  Welcome back, {{user.name}}!
{{else}}
  Please activate your account.
{{/if}}
```
Renders the block if the expression is truthy, otherwise renders the `{{else}}` block if present.

**Block parameters for `if`/`unless`:**
```handlebars
{{#if user.active as |active|}}
  Status: {{active}}
{{/if}}
```
Block parameters create a local variable with the condition value. Useful for avoiding repeated evaluation.

**else if shorthand:**
```handlebars
{{#if user.role == "admin"}}
  Admin panel
{{else if user.role == "moderator"}}
  Moderator panel
{{else}}
  User panel
{{/if}}
```
The `{{else if condition}}` syntax creates nested if blocks. You can also use `{{elseif condition}}` as an alternative.

**Inverted condition (`unless`):**
```handlebars
{{#unless user.active}}
  Please activate your account.
{{/unless}}
```
Renders the block if the expression is falsy. Also supports block parameters and else if.

**Context switching (`with`):**
```handlebars
{{#with user}}
  <h1>{{name}}</h1>
  <p>{{role}}</p>
{{else}}
  <p>No user data</p>
{{/with}}
```
Changes the context to the specified value inside the block. If the value is falsy, renders the `{{else}}` block.

**Iteration (`each`):**
```handlebars
{{#each users}}
  <li>{{name}}</li>
{{else}}
  <li>No users yet.</li>
{{/each}}
```
Iterates over arrays, slices, or maps. Inside the block, `{{this}}` or `{{.}}` refers to the current item. If the collection is empty, renders the `{{else}}` block.

**Block parameters:**
```handlebars
{{#each users as |person idx|}}
  {{idx}}: {{person.name}}
{{/each}}
```
Block parameters create local variables inside the block. The first parameter (`person`) is the item, the second (`idx`) is the index.

**Block parameters with maps:**
```handlebars
{{#each settings as |val key|}}
  {{key}} = {{val}}
{{/each}}
```
For maps, the first parameter is the value, the second is the key.

**Custom block helpers:**
```handlebars
{{#myHelper arg1 arg2 key=value}}
  Block content
{{else}}
  Inverse content
{{/myHelper}}
```
Any registered helper can be used as a block helper. The helper receives `BlockOptions` with `Fn` and `Inverse` callbacks to render the block content. Block helpers should check for `BlockOptions` in their arguments and call the appropriate callback.

### Paths and Data Variables

**Parent paths:**
```handlebars
{{#each users}}
  {{name}} - {{../title}}
{{/each}}
```
`../` accesses the parent context. Useful inside nested blocks.

**Data variables:**
```handlebars
{{#each items}}
  {{@index}} - {{name}} ({{@first}}/{{@last}})
{{/each}}
```
Special variables available in certain contexts:
- `@index` - Current index in `each` loops (0-based)
- `@key` - Current key in `each` loops over maps
- `@first` - `true` if this is the first item
- `@last` - `true` if this is the last item
- `@root` - The root context (top-level data)

**Root access:**
```handlebars
{{#with user}}
  {{name}} - {{@root.title}}
{{/with}}
```
Access the root context from anywhere using `@root`.

### Truthiness

Values are considered **falsy** when they are:
- `nil`
- `false`
- `0` (any numeric zero)
- `""` (empty string)
- Empty arrays, slices, or maps

Everything else is **truthy**, including:
- Non-zero numbers
- Non-empty strings
- Non-empty arrays/slices/maps
- Objects/structs

### Whitespace Control

**Trim left whitespace:**
```handlebars
Title: {{~title}}
```
The `~` before the expression trims whitespace to the left.

**Trim right whitespace:**
```handlebars
{{name~}}, {{role}}
```
The `~` after the expression trims whitespace to the right.

**Combined:**
```handlebars
{{~name~}}
```
Trims whitespace on both sides.

### Raw Blocks

Raw blocks prevent parsing of inner content:
```handlebars
{{{{raw}}}}
  {{this will be output literally}}
  {{#if something}} not parsed {{/if}}
{{{{/raw}}}}
```
Useful for outputting Handlebars syntax or other template-like content that should not be processed.

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

Dynamic partials:

```
{{> (lookup . "cardPartial") user}}
```

## Built-in Helpers Library

go-hbars includes a comprehensive helpers library matching Handlebars.js core and handlebars-helpers 7.4. Import the helpers package and use the registry:

```go
import "github.com/andriyg76/go-hbars/helpers"

// In your go:generate command
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -helper upper=github.com/andriyg76/go-hbars/helpers/handlebars:Upper \
//  -helper lower=github.com/andriyg76/go-hbars/helpers/handlebars:Lower \
//  -helper formatDate=github.com/andriyg76/go-hbars/helpers/handlebars:FormatDate
```

Or use the registry helper to get all helpers at once:

```go
import (
	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

registry := helpers.Registry()
compilerHelpers := make(map[string]compiler.HelperRef)
for name, ref := range registry {
	compilerHelpers[name] = compiler.HelperRef{
		ImportPath: ref.ImportPath,
		Ident:      ref.Ident,
	}
}
opts := compiler.Options{
	PackageName: "templates",
	Helpers:     compilerHelpers,
}
```

### Available Helpers

**String helpers:**
- `upper`, `lower` - Convert case
- `capitalize`, `capitalizeAll` - Capitalize words
- `truncate` - Truncate strings with optional suffix
- `reverse` - Reverse a string
- `replace` - Replace substrings
- `stripTags`, `stripQuotes` - Remove HTML tags or quotes
- `join`, `split` - Join/split arrays with separator

**Comparison helpers:**
- `eq`, `ne` - Equality checks
- `lt`, `lte`, `gt`, `gte` - Numeric comparisons
- `and`, `or`, `not` - Logical operations

**Date helpers:**
- `formatDate` - Format dates with custom format (Go time format)
- `now` - Current time
- `ago` - Human-readable time ago

**Collection helpers:**
- `lookup` - Look up values by key
- `default` - Fallback for empty values
- `length` - Get length of strings/arrays/objects
- `first`, `last` - Get first/last array element
- `inArray` - Check if value is in array

**Math helpers:**
- `add`, `subtract`, `multiply`, `divide`, `modulo` - Arithmetic
- `floor`, `ceil`, `round`, `abs` - Rounding and absolute value
- `min`, `max` - Min/max of two numbers

**Number helpers:**
- `formatNumber` - Format with precision and separator
- `toInt`, `toFloat`, `toNumber` - Type conversions
- `toFixed` - Fixed decimal places
- `toString` - Convert to string

**Object helpers:**
- `has` - Check if object has property
- `keys`, `values` - Get object keys/values
- `size` - Get object/array size
- `isEmpty`, `isNotEmpty` - Empty checks

**URL helpers:**
- `encodeURI`, `decodeURI` - URI encoding/decoding
- `stripProtocol`, `stripQuerystring` - URL manipulation

## Custom Helpers

You can implement custom helpers as regular Go functions and map them with
`-helper name=Ident`. Helper functions must match this signature:

```go
func MyHelper(ctx *runtime.Context, args []any) (any, error)
```

Hash arguments are passed as the last element in `args`. Use `runtime.HashArg(args)` to retrieve them:

```go
func FormatCurrency(ctx *runtime.Context, args []any) (any, error) {
	amount := args[0]
	hash, _ := runtime.HashArg(args)
	symbol := "$"
	if hash != nil {
		if s, ok := hash["symbol"].(string); ok {
			symbol = s
		}
	}
	return fmt.Sprintf("%s%.2f", symbol, amount), nil
}
```

### Block Helpers

Any helper can be used as a block helper. When used as a block, the helper receives `runtime.BlockOptions` as the last argument. Use `runtime.GetBlockOptions(args)` to retrieve it:

```go
func MyBlockHelper(ctx *runtime.Context, args []any) (any, error) {
	opts, ok := runtime.GetBlockOptions(args)
	if !ok {
		// Not used as a block, handle as regular helper
		return "default", nil
	}
	
	// Render the block content
	var b strings.Builder
	if err := opts.Fn(ctx, &b); err != nil {
		return nil, err
	}
	return b.String(), nil
}
```

Block helpers can conditionally render the main block (`opts.Fn`) or the inverse/else block (`opts.Inverse`):

```go
func IfHelper(ctx *runtime.Context, args []any) (any, error) {
	opts, ok := runtime.GetBlockOptions(args)
	if !ok {
		return nil, fmt.Errorf("if helper must be used as a block")
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

## Compatibility fixtures

See `examples/compat` for a small template set that exercises hash arguments,
subexpressions, data variables, parent paths, block params, dynamic partials,
whitespace control, and raw blocks.

## Compile-time optimization plan (todo)

- Constant-fold `if`/`unless`/`with` when the condition is a literal.
- Inline literal arguments directly (avoid `EvalArg`).
- Prebuild hash maps when all values are literals.
- Detect duplicate hash keys at compile time.
- Pre-resolve static partial names from string literals.
- Pre-parse path segments for faster runtime lookup.

## Implementation status

All core Handlebars syntax features are now implemented:
- ✅ Custom block helpers
- ✅ Block params for `if`/`unless`
- ✅ `else if` shorthand
- ✅ Partial blocks (`{{#> partial}}...{{/partial}}`)
