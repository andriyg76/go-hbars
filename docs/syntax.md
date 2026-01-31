# Handlebars Syntax Guide

This document describes the Handlebars template syntax supported by go-hbars.

**Custom extensions** (includeZero for `{{#if}}`/`{{#unless}}`) are covered in [Custom Extensions](extensions.md).

## Values and Expressions

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

## Helpers

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

## Partials

**Basic partial (current context only):**
```handlebars
{{> header}}
```
Renders the `header` partial with the **current context**. The current context is passed only when there are no explicit context argument and no hash.

**Partial with explicit context:**
```handlebars
{{> userCard user}}
{{> footer . note="thanks"}}
```
Renders the partial with the given context (e.g. `user` or `.`). Hash arguments are merged with that context (hash overrides same keys).

**Partial with only hash (no explicit context):**
```handlebars
{{> footer note="thanks"}}
```
Renders `footer` with a context built from the **hash** plus **only the keys that the partial template uses** from the current scope. So the partial receives the hash and any other root-level paths it references (e.g. `{{title}}` in the partial), taken from the current context.

**Dynamic partial names:**
```handlebars
{{> (lookup . "cardPartial") user}}
```
Uses a helper to determine the partial name at runtime.

## Block Helpers

**Conditional rendering (`if`):**
```handlebars
{{#if user.active}}
  Welcome back, {{user.name}}!
{{else}}
  Please activate your account.
{{/if}}
```
Renders the block if the expression is truthy, otherwise renders the `{{else}}` block if present.

**Custom extension — includeZero:**  
To treat numeric zero as truthy (e.g. “0 items” vs “no data”), use `includeZero=true`:  
`{{#if count includeZero=true}}...{{/if}}`. See [Custom Extensions — includeZero](extensions.md#includezero) for details.

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

**Universal section:**  
Any `{{#name}}...{{/name}}` that is not a built-in (`if`/`unless`/`with`/`each`) and not a registered helper is treated as a section: resolve `name` from context; if truthy, render the block with that value as context; else render `{{else}}` if present. Same semantics as `{{#with name}}...{{/with}}`. See [Custom Extensions — Universal section](extensions.md#universal-section).

## Paths and Data Variables

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

## Truthiness

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

## Whitespace Control

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

## Raw Blocks

Raw blocks prevent parsing of inner content:
```handlebars
{{{{raw}}}}
  {{this will be output literally}}
  {{#if something}} not parsed {{/if}}
{{{{/raw}}}}
```
Useful for outputting Handlebars syntax or other template-like content that should not be processed.

## Template Examples

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

## See also

- **[Custom Extensions](extensions.md)** — includeZero for `{{#if}}`/`{{#unless}}`
