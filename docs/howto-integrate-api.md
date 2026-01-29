# How to integrate go-hbars API into your program (templates + go:generate)

Step-by-step guide to add Handlebars templates to a Go project using go-hbars from GitHub and `go:generate` with the compiler (hbc). No local `replace` — dependency is resolved from GitHub.

**Alternative:** use the [init](init.md) command to create a new project or add templates to an existing module: `go run github.com/andriyg76/go-hbars/cmd/init@latest new myapp` or `init add`.

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

(How template file names map to Go function names and what the generated file contains: see [Compiled template file](compiled-templates.md).)

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

## Working with a local checkout

When developing go-hbars itself or testing changes before a release:

1. Clone the repo locally, e.g. `~/src/go-hbars`.
2. In your application’s `go.mod` add a `replace` so the module points at the local copy:

   ```go
   replace github.com/andriyg76/go-hbars => /home/you/src/go-hbars
   ```

3. Keep the same `//go:generate` line (with or without `@latest`). When you run `go generate ./...`, `go run` will use the replaced module and thus your local hbc.
4. From the go-hbars repo you can also run the compiler directly: `go run ./cmd/hbc -in /path/to/templates -out /path/to/templates_gen.go -pkg templates` (from the repo root).

Generated files include a `// Generator version: ...` comment when the compiler was built with version info (e.g. from the hbc CLI).
