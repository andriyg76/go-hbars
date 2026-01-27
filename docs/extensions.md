# Custom Extensions

go-hbars adds a small set of **custom extensions** that are not part of standard Handlebars.js. They are opt-in via hash options and keep default behavior unchanged.

## includeZero

**Availability:** `{{#if}}` and `{{#unless}}`

By default, numeric zero (`0`, `0.0`) is falsy: `{{#if count}}...{{/if}}` skips the block when `count` is `0`. The **includeZero** option makes numeric zero count as truthy for that block, so the block is rendered when the value is zero.

**Syntax:**
```handlebars
{{#if count includeZero=true}}
  Zero or positive
{{else}}
  Missing or negative
{{/if}}
```

**Behavior:**
- `includeZero=true` — numeric zero is treated as truthy; the block runs when the value is `0` (or `0.0`).
- Omitted or `includeZero=false` — standard behavior: zero is falsy.
- Only the first (condition) argument is affected; other hash options are ignored for this logic.
- Supports `json.Number` when decoding JSON with `UseNumber()`.

**Example:**
```handlebars
{{! count is 0 }}
{{#if count includeZero=true}}show{{else}}hide{{/if}}
{{! outputs: show }}

{{#if count}}show{{else}}hide{{/if}}
{{! outputs: hide }}
```

**Rationale:** Useful when zero is a valid, “present” value (e.g. “0 items” vs “no data”) and you want the main block to run for zero without changing your data shape.
