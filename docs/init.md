# go-hbars init: create or add to a project

The `init` command creates a new go-hbars project or adds templates and bootstrap to an existing Go module. Run it with:

```bash
go run github.com/andriyg76/go-hbars/cmd/init@latest <subcommand> [options]
```

Or install and run:

```bash
go install github.com/andriyg76/go-hbars/cmd/init@latest
init new myapp -bootstrap
```

## Subcommands

### init new [path]

Creates a new project. **path** defaults to the current directory (`.`).

| Flag | Description |
|------|-------------|
| `-bootstrap` | Include `processor/templates/`, `data/`, `shared/`, and a main that uses `NewQuickServer()` / `NewQuickProcessor()`. Without this, creates a simple API-style project with `templates/` and `RenderXxxString` usage. |
| `-module` | Module path for `go mod init` (default: directory name). |

**Examples:**

```bash
# New API-only project in ./myapp
init new myapp

# New bootstrap project (server + static generator)
init new myapp -bootstrap

# Same, path and flag in any order
init new -bootstrap myapp

# Current directory, explicit module name
init new . -module example.com/myapp
```

After creation, `init` runs `go generate ./...` and `go mod tidy` in the project directory so template code is generated and dependencies are resolved.

### init add

Adds templates (and optionally bootstrap) to the **current** directory. The current directory must be a Go module (have a `go.mod`).

| Flag | Description |
|------|-------------|
| `-bootstrap` | Add `processor/templates/`, `data/`, `shared/`, and an example main. If `main.go` already exists, creates `main_hbars_example.go` for you to merge. |

**Examples:**

```bash
cd /path/to/your/module
init add              # templates/ + gen.go + sample .hbs
init add -bootstrap   # processor/templates, data/, shared/, example main
```

Existing files (e.g. `gen.go`, `main.go`) are not overwritten; new files are created alongside. After adding files, `init` runs `go generate ./...` and `go mod tidy` in the module directory.

## Local checkout

When using a local clone of go-hbars, use a `replace` in your project’s `go.mod` and run init from the repo:

```bash
cd /path/to/go-hbars
go run ./cmd/init new /path/to/myapp -bootstrap
```

See [Working with a local checkout](howto-integrate-api.md#working-with-a-local-checkout) in the integration guides.
