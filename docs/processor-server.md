# Static Site Processor and Web Server

go-hbars includes CLI tools for generating static sites and running a semi-static web server.

## Quick Start

1. **Create your project structure:**
```
project/
├── .processor/
│   ├── templates/
│   │   ├── main.hbs
│   │   ├── header.hbs
│   │   └── footer.hbs
│   └── templates_gen.go  # with go:generate directive
├── data/
│   └── index.json
└── shared/
    └── site.json
```

2. **Add go:generate directive to compile templates:**

Create `.processor/templates_gen.go` (run `go generate` from the project root; the directive runs with `.processor` as the current directory):
```go
//go:generate hbc -in templates -out templates_gen.go -pkg templates -bootstrap

package templates
```

The `-bootstrap` flag generates helper functions for quick server/processor setup.

3. **Generate templates:**
```bash
go generate ./.processor/...
# or
go generate ./...
```

4. **Create data files:**

`data/index.json`:
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

`shared/site.json`:
```json
{
  "name": "My Site",
  "url": "https://example.com"
}
```

## Static Site Generation

Generate static HTML files for hosting:

```bash
go run ./cmd/build --data-path data --output-path pages
```

### CLI Options

**Build command (`cmd/build`):**
- `--root` - Base directory for relative paths (default: current directory)
- `--data-path` - Data files directory (default: `data`)
- `--shared-path` - Shared data directory (default: `shared`)
- `--templates-path` - Templates directory (default: `.processor/templates`)
- `--output-path` - Output directory (default: `pages`)

## Semi-Static Web Server

Run a development server that generates pages on the fly:

```bash
go run ./cmd/server --data-path data --addr :8080
```

### CLI Options

**Server command (`cmd/server`):**
- `--root` - Base directory for relative paths (default: current directory)
- `--data-path` - Data files directory (default: `data`)
- `--shared-path` - Shared data directory (default: `shared`)
- `--templates-path` - Templates directory (default: `.processor/templates`)
- `--static-dir` - Static files directory (optional)
- `--addr` - Address to listen on (default: `:8080`)

## Data File Format

Each data file should include a `_page` section:

**JSON:**
```json
{
  "_page": {
    "template": "blog/post",
    "output": "blog/hello.html"
  },
  "title": "Hello",
  "author": "Ada"
}
```

**YAML:**
```yaml
_page:
  template: blog/post
  output: blog/hello.html
title: Hello
author: Ada
```

**TOML:**
```toml
[_page]
template = "blog/post"
output = "blog/hello.html"

title = "Hello"
author = "Ada"
```

The `_page` section specifies:
- `template`: Name of the template to use (relative to templates directory, without `.hbs` extension)
- `output`: Optional output path (relative to output directory). If omitted, uses the input file name with `.html` extension.

## Shared Data

Shared data files are loaded from the `shared/` directory and merged into all pages under the `_shared` key:

**Structure:**
```
shared/
  site.json
  navigation/
    menu.yaml
```

**Access in templates:**
```handlebars
<title>{{_shared.site.name}}</title>
<nav>
  {{#each _shared.navigation.menu.items}}
    <a href="{{href}}">{{label}}</a>
  {{/each}}
</nav>
```

## Using go:generate

The recommended workflow uses `go:generate` to automatically compile templates:

**In your template package file (e.g. `.processor/templates_gen.go` or `processor/templates/gen.go`):**
```go
//go:generate hbc -in . -out ./templates_gen.go -pkg templates -bootstrap

package templates
```

Run `go:generate` from the directory that contains the template files (e.g. `processor/templates/` or `.processor/templates/`); use `-in .` and `-out ./templates_gen.go` so the generated file lives next to the templates.

**Benefits:**
- Templates are automatically recompiled when you run `go generate ./...`
- Works seamlessly with `go build` and CI/CD pipelines
- No manual compilation step needed
- Templates are type-checked at compile time

**Workflow:**
1. Edit templates in `.processor/templates/`
2. Run `go generate ./...` to recompile
3. Build and run your application

**CI/CD Integration:**
```yaml
# Example GitHub Actions workflow
- name: Generate templates
  run: go generate ./...

- name: Build
  run: go build ./...
```

