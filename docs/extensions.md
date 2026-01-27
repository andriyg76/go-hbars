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

## Layout blocks (partial / block)

**Availability:** block helpers `{{#partial}}` and `{{#block}}` (first-class extensions, not tied to sitegen).

Layout blocks let a **page** define slot content that a **layout** renders. A shared `ctx.Blocks` map is passed down through partials and scopes, so the page and layout see the same data in one pass.

**Syntax:**

- **`{{#partial "name"}}...{{/partial}}`** — renders the block body into a buffer and stores it in the shared blocks map under `"name"`. Writes nothing to the main output. Used by the page to define slot content.
- **`{{#block "name"}}...{{/block}}`** — if `"name"` is set in the blocks map, outputs that content; otherwise runs the block body (default content) to the current output. Used by the layout to render a slot.

**Direction A (page → layout, one pass):**

1. The page calls the layout partial, e.g. `{{> layout}}`.
2. Before that, the page can use `{{#partial "header"}}...{{/partial}}` to fill slots.
3. The layout uses `{{#block "header"}}default content{{/block}}` to render each slot; if the page filled it, that content is shown; otherwise the default is shown.
4. The same context (including `ctx.Blocks` and `ctx.Output`) is passed into the layout partial, so one render pass is enough.

**Behavior:**
- `{{#partial "name"}}` creates `ctx.Blocks` if nil, then runs the block body with a buffer and stores the result in `ctx.Blocks["name"]`. It does not write to `ctx.Output`.
- `{{#block "name"}}default{{/block}}` writes `ctx.Blocks["name"]` to `ctx.Output` when that key is non-empty; otherwise it runs the default block (the body) to `ctx.Output`.
- `ctx.Blocks` and `ctx.Output` are preserved through `WithScope` (e.g. inside `{{#with}}`, `{{#each}}`, and when invoking partials), so nested templates share the same blocks map.

**Example (page → layout):**

Page template (e.g. `main`):

```handlebars
{{#partial "header"}}<title>My Page</title>{{/partial}}
{{> layout}}
```

Layout partial:

```handlebars
<html>
<head>{{#block "header"}}<title>Default</title>{{/block}}</head>
<body>content</body>
</html>
```

When `main` is rendered, the output includes `<title>My Page</title>` in the head, because the page filled the `"header"` slot before calling the layout.
