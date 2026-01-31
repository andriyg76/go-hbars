# Testing

## Unit tests

Run all tests:

```bash
go test ./...
```

Use `-short` to skip long-running tests (including E2E).

## E2E tests

End-to-end tests live in `internal/compiler/e2e/`. They:

1. Compile Handlebars templates with the project compiler
2. Write generated Go code and a small driver into a temporary module
3. Run `go run` and assert on output

These tests are skipped when you pass `-short`:

```bash
go test ./... -short          # skips E2E
go test ./internal/compiler/e2e/... -v -count=1   # runs only E2E
```

### E2E test list

| Test | Description |
|------|-------------|
| `TestE2E_Compat_IteratorGenerated` | Compiles compat templates; asserts generated iterator code (e.g. `Users()`, `range`) |
| `TestE2E_CompatTemplates` | Compiles compat, runs generated code with `data.json`, compares to `expected.txt` |
| `TestE2E_IncludeZero` | `{{#if count includeZero=true}}` with `count=0` renders "zero" |
| `TestE2E_Showcase_NilContext` | Showcase templates with nil/empty context; no panic; dynamic partial error in output |
| `TestE2E_UniversalSection` | Block helper `date` and conditional; asserts output |
| `TestE2E_UserProject_Bootstrap_ServerAndProcessor` | User-style project with `-bootstrap`, `go generate`, `NewQuickProcessor()`; checks generated HTML |
| `TestE2E_UserProject_GoGenerate_CompatShowcase` | User-style project with go:generate (no bootstrap); compat + showcase data, RenderCompatString / RenderShowcaseString with `XxxContextFromMap` |

### Context and `FromMap`

Generated templates expect a context type (e.g. `MainContext`, `CompatContext`). When your data is `map[string]any` (e.g. from JSON), use the generated `XxxContextFromMap` so it satisfies the context interface:

```go
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

E2E tests that pass JSON data use this pattern.
