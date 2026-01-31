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

---

## Universal section

**Availability:** any block `{{#name}}...{{/name}}` that is **not** a built-in (`if`, `unless`, `with`, `each`) and **not** a registered helper.

For `{{#anything}}...{{/anything}}`, if `anything` is not a known block helper, go-hbars treats it as a **section**: resolve `anything` from the context; if truthy, render the block with that value as context; otherwise render the `{{else}}` branch if present. Semantically this is the same as `{{#with anything}}...{{else}}...{{/with}}`.

**Syntax:**
```handlebars
{{#date}}
  <span>{{date}}</span>
{{else}}
  no date
{{/date}}
```
With no expression after the name, the block name is used as the path. You can also write `{{#section some.path}}...{{/section}}`; then `some.path` is the expression.

**Behavior:**
- **Registered helper wins** — if `date` (or whatever name) is in the helpers registry as a block helper, it is invoked as a custom block helper.
- **Otherwise section** — the block is compiled as `{{#with <expr>}}...{{else}}...{{/with}}`, where `<expr>` is the block's first argument or, if empty, the block name. So `{{#date}}` uses the context key `date`; `{{#foo bar}}` uses the expression `bar`.
- Compatible with **Handlebars.java** and similar engines where `{{#date}}` means "if date is present, render block with date as context".

**Example:**
```handlebars
{{! data: { date: "2024-01-15" } }}
{{#date}}{{date}}{{/date}}
{{! outputs: 2024-01-15 }}

{{! data: { date: "" } }}
{{#date}}shown{{else}}none{{/date}}
{{! outputs: none }}
```

