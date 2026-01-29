# How to integrate go-hbars bootstrap (QuickServer + QuickProcessor)

Step-by-step guide to add Handlebars templates with **bootstrap** code: `-bootstrap` generates `NewQuickServer()` and `NewQuickProcessor()` so you can run a semi-static HTTP server or generate static HTML from data files. Uses go-hbars from GitHub (no local `replace` in production).

**Alternative:** use the [init](init.md) command: `go run github.com/andriyg76/go-hbars/cmd/init@latest new myapp -bootstrap` or `init add -bootstrap` in an existing module.

## 1. Create a new project

```bash
mkdir myapp && cd myapp
go mod init myapp
```

## 2. Add templates and go:generate with -bootstrap

Create a directory for templates. Because Go does not allow import paths with a leading dot (e.g. `.processor/templates`), use a path like `processor/templates/` or `templates/`.

Put your `.hbs` files there (e.g. `main.hbs`, `header.hbs`, `footer.hbs`).

Add a file that triggers code generation **with `-bootstrap`**. For example `processor/templates/gen.go`:

```go
//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
```

- `-bootstrap` — generates `NewQuickServer()` and `NewQuickProcessor()` in addition to `RenderXxx` functions.
- The generated package will only import public packages (`pkg/renderer`, `pkg/sitegen`), so it can be used from your module without importing internal packages.

## 3. Generate template code

From the project root:

```bash
go generate ./...
go mod tidy
```

## 4. Add data files with `_page`

Each data file that should produce a page must include a `_page` section:

- `template` — name of the template (e.g. `main`, matching `main.hbs`).
- `output` — output path relative to the output directory (e.g. `index.html`, `blog/post.html`).

Example `data/index.json`:

```json
{
  "_page": {
    "template": "main",
    "output": "index.html"
  },
  "title": "Welcome",
  "content": "Hello, world!"
}
```

Create a `data/` directory and add one or more JSON (or YAML/TOML) files with `_page` and your template data.

## 5. Use QuickProcessor (static site generation)

In your `main.go` (or a CLI command), use the generated `NewQuickProcessor()` and optionally set config paths:

```go
package main

import (
	"log"

	templates "myapp/processor/templates"
)

func main() {
	proc, err := templates.NewQuickProcessor()
	if err != nil {
		log.Fatal(err)
	}
	proc.Config().DataPath = "data"
	proc.Config().OutputPath = "pages"

	if err := proc.Process(); err != nil {
		log.Fatal(err)
	}
}
```

Run with `go run .` (or build and run). This reads all files under `data/`, merges shared data from `shared/` (if present), and writes HTML under `pages/` (or your `OutputPath`).

## 6. Use QuickServer (development server)

To run a semi-static HTTP server that renders pages on demand:

```go
	srv, err := templates.NewQuickServer()
	if err != nil {
		log.Fatal(err)
	}
	srv.Config().DataPath = "data"
	srv.Config().Addr = ":8080"

	log.Fatal(srv.Start())
```

The server maps URL paths to data files (e.g. `/` → `data/index.json`, `/about` → `data/about.json`) and renders them with the configured template.

## 7. Shared data (optional)

Create a `shared/` directory. JSON/YAML/TOML files there are loaded and merged into every page under the `_shared` key. Use `{{_shared.site.name}}` and similar in templates.

## Summary

| Step | Action |
|------|--------|
| 1 | New module: `go mod init myapp` |
| 2 | Add `processor/templates/*.hbs` and `processor/templates/gen.go` with `//go:generate ... hbc@latest ... -bootstrap` |
| 3 | Run `go generate ./...` then `go mod tidy` |
| 4 | Add `data/*.json` (or YAML/TOML) with `_page.template` and `_page.output` |
| 5 | In main: `templates.NewQuickProcessor()` and `proc.Process()` for static build, or `templates.NewQuickServer()` and `srv.Start()` for HTTP |
| 6 | Set Config().DataPath, OutputPath/Addr to match your layout |

Bootstrap uses only public packages (`pkg/renderer`, `pkg/sitegen`), so your project can depend on go-hbars from GitHub without a local `replace`.

### Working with a local checkout

When developing go-hbars or testing unreleased changes:

1. Clone the repo locally (e.g. `~/src/go-hbars`).
2. In your app’s `go.mod` add: `replace github.com/andriyg76/go-hbars => /path/to/go-hbars`
3. Keep the same `//go:generate` line; `go generate ./...` will use the local hbc via the replace.
4. Or from the go-hbars repo: `go run ./cmd/hbc -in . -out ./templates_gen.go -pkg templates -bootstrap` (run from the template directory or pass correct `-in`).

Generated code includes a `// Generator version: ...` comment when the compiler supplies version info.

(What bootstrap generates and the developer-facing interface: see [Bootstrap-generated code](bootstrap-generated.md).)
