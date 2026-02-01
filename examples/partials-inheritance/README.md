# Partials inheritance (canonical partial context)

Example where multiple root templates include the same partial with the same scope (e.g. `{{> menu}}`). The compiler emits one canonical context interface per partial (e.g. `MenuContext`); all callers use that type (no primary-embedding types). Root contexts that include a single layout embed that layout’s context and only add template-specific methods.

## Templates

- **default** – page that includes `{{> layout}}`
- **firstpage** – page that includes `{{> layout}}`
- **news** – page that includes `{{> layout}}`
- **layout** – includes menu with explicit context: `{{> menu _shared.menu}}`
- **menu** – partial (nav with title), used only by layout

## Data

- `data/index.json` – data for the index page (`_page.template`: default, `_page.output`: index.html)
- `data/news.json` – data for the news page (`_page.template`: news, `_page.output`: news.html)

## Build two pages: index.html and news.html

1. Generate Go code from templates (from repo root):

   ```bash
   go run ./cmd/hbc -in examples/partials-inheritance/templates -out examples/partials-inheritance/gen/templates_gen.go -pkg templates
   ```

2. Build pages (from this directory):

   ```bash
   cd examples/partials-inheritance
   go run .
   ```

   This writes **index.html** (from template `default`) and **news.html** (from template `news`) using data from `data/index.json` and `data/news.json`.
