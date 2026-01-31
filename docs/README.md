# go-hbars Documentation

## Getting started

- **[init: create or add to a project](init.md)** — Scaffold a new project or add templates to an existing module (`init new`, `init add`).
- **[How to integrate API](howto-integrate-api.md)** — Add templates and `go:generate` to a Go project (no bootstrap).
- **[How to integrate bootstrap](howto-integrate-bootstrap.md)** — Add QuickServer + QuickProcessor with data files and `_page`.

## Reference

- **[Template Syntax](syntax.md)** — Handlebars syntax supported by go-hbars.
- **[Custom Extensions](extensions.md)** — includeZero, universal section.
- **[Built-in Helpers](helpers.md)** — String, comparison, date, collection, math, object, URL helpers.
- **[Template API](api.md)** — Runtime API for compiled templates (context types, helpers, partials).
- **[Compiled template file](compiled-templates.md)** — What `hbc` generates (names, functions, context types).
- **[Bootstrap-generated code](bootstrap-generated.md)** — What `-bootstrap` adds (NewQuickServer, NewQuickProcessor).

## Static site and server

- **[Processor & Server](processor-server.md)** — CLI tools (`cmd/build`, `cmd/server`), data format, shared data.
- **[Embedded API](embedded.md)** — Embed processor and server in your app (sitegen, custom renderer).

## Other

- **[Testing](testing.md)** — Unit and E2E tests.
