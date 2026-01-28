# How to integrate go-hbars API into your program (templates + go:generate)

Step-by-step guide to add Handlebars templates to a Go project using go-hbars from GitHub and `go:generate` with the compiler (hbc). No local `replace` — dependency is resolved from GitHub.

## 1. Create a new project

```bash
mkdir myapp && cd myapp
go mod init myapp
```

## 2. Add the template package and go:generate

Create a directory for templates, for example `templates/`, and put your `.hbs` files there (e.g. `main.hbs`, `header.hbs`, `footer.hbs`).

Add a file that triggers code generation. For example `templates/gen.go`:

```go
//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates

package templates
```

- `-in .` — current directory (the `templates/` package dir) is the template root.
- `-out ./templates_gen.go` — generated Go file in the same package.
- `-pkg templates` — package name for the generated code.

Using `go run .../cmd/hbc@latest` runs the compiler from GitHub; you do not need to install `hbc` separately.

## 3. Generate template code

From the project root:

```bash
go generate ./...
go mod tidy
```

This downloads go-hbars (and hbc) if needed, generates `templates_gen.go`, and updates `go.mod`/`go.sum`.

## 4. Use the generated API in your program

In your `main.go` (or any package), import the templates package and call the generated render functions:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	templates "myapp/templates"
)

func main() {
	dataBytes, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read data: %v\n", err)
		os.Exit(1)
	}
	var data map[string]any
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		fmt.Fprintf(os.Stderr, "parse data: %v\n", err)
		os.Exit(1)
	}

	// Render to string (template name = file name without .hbs, e.g. main -> RenderMainString)
	out, err := templates.RenderMainString(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(out)
}
```

For each template `name.hbs`, the generated package exposes:

- `RenderName(w io.Writer, data any) error`
- `RenderNameString(data any) (string, error)`

## Generated file details (compiled templates)

The generated file (e.g. `templates_gen.go`) contains:

- Package header and imports (`io`, `strings`, the runtime package, and helper imports if used).
- Internal render functions named `renderXxx(ctx *runtime.Context, w io.Writer) error`.
- Public functions named `RenderXxx(w io.Writer, data any) error` and `RenderXxxString(data any) (string, error)`.
- A `partials` map used for `{{> partial}}` lookups inside templates.

### How template file names map to function names

Template names are derived from the file path relative to the template root (`-in`):

- `main.hbs` → `RenderMain` / `RenderMainString`
- `blog/post.hbs` → `RenderBlogPost` / `RenderBlogPostString`
- `user-card.hbs` → `RenderUserCard` / `RenderUserCardString`
- `user_card.hbs` → `RenderUserCard` / `RenderUserCardString`
- `user.profile.hbs` → `RenderUserProfile` / `RenderUserProfileString`

Rules:

- The `.hbs` extension is removed.
- Path separators and non-alphanumeric characters are treated as word boundaries.
- Words are title-cased and concatenated (`blog/post` → `BlogPost`).
- If two templates produce the same identifier, compilation fails with a clear error.

### Partial lookup names

Partials are keyed by their template name (path without `.hbs`), for example:

- `partials["header"]` for `header.hbs`.
- `partials["blog/post"]` for `blog/post.hbs`.

When you write `{{> blog/post}}` inside a template, the compiler resolves it to the matching entry in `partials`.

## 5. Run the program

```bash
go run .
```

## Summary

| Step | Action |
|------|--------|
| 1 | New module: `go mod init myapp` |
| 2 | Add `templates/*.hbs` and `templates/gen.go` with `//go:generate go run github.com/andriyg76/go-hbars/cmd/hbc@latest -in . -out ./templates_gen.go -pkg templates` |
| 3 | Run `go generate ./...` then `go mod tidy` |
| 4 | In main: import templates, load data, call `templates.RenderXxxString(data)` |
| 5 | Run with `go run .` |

No `replace` in `go.mod` is required; the dependency is taken from GitHub. To pin a version, use a specific tag instead of `@latest` in the `go:generate` line (e.g. `@v0.1.0` when available).
