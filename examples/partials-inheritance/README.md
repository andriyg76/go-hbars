# Partials inheritance (canonical partial context)

Example demonstrating how the compiler handles shared partials and context embedding.

## Key concepts

- **Same-scope partial inclusion** (`{{> layout}}`): When multiple root templates include the same partial without an explicit context, the compiler emits a single canonical context interface (e.g., `LayoutContext`) and all callers **embed** this type.
- **Different-context partial inclusion** (`{{> menu _shared.menu}}`): When a partial is included with an explicit context, the partial has its own independent interface; the caller just provides access to the context path.

## Templates

- **default.hbs** – page that includes `{{> layout}}` (same scope)
- **firstpage.hbs** – page that includes `{{> layout}}` (same scope)
- **news.hbs** – page that includes `{{> layout}}` (same scope)
- **layout.hbs** – includes menu with explicit context: `{{> menu _shared.menu}}`
- **menu.hbs** – partial (nav with title), used only by layout

## Generated context structure

```go
// LayoutContext is the canonical context for layout.hbs
type LayoutContext interface {
    Shared() LayoutSharedContext
    Raw() any
}

// DefaultContext embeds LayoutContext (same-scope inclusion)
type DefaultContext interface {
    LayoutContext  // embedded
    Title() any
    Raw() any
}

// MenuContext is independent (different-context inclusion)
type MenuContext interface {
    Title() any
    Raw() any
}
```

## Data

- `data/index.json` – data for the index page (`_page.template`: default, `_page.output`: index.html)
- `data/news.json` – data for the news page (`_page.template`: news, `_page.output`: news.html)

## Build two pages: index.html and news.html

1. Generate Go code from templates (from repo root):

   ```bash
   go run ./cmd/hbc -in examples/partials-inheritance/templates -out examples/partials-inheritance/gen/templates_gen.go -pkg templates
   ```

   To also emit all compiler intermediate outputs (phase AST, type analysis, IR) in this directory:

   ```bash
   cd examples/partials-inheritance
   go run github.com/andriyg76/go-hbars/cmd/hbc -in templates -out gen/templates_gen.go -pkg templates \
     -max-phase 5 -phase1-output phase1_ast.json -phase2a-output phase2a.json \
     -phase2b-output phase2b.json -phase3-output phase3_ir.json
   ```

   This produces: `gen/templates_gen.go`, `phase1_ast.json`, `phase2a.json`, `phase2b.json`, `phase3_ir.json`.

2. Build pages (from this directory):

   ```bash
   cd examples/partials-inheritance
   go run .
   ```

   This writes **index.html** (from template `default`) and **news.html** (from template `news`) using data from `data/index.json` and `data/news.json`.
