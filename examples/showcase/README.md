Showcase templates

This folder contains example templates and data that exercise most supported
syntax and helpers.

Files
- templates/: Handlebars templates and partials
- data.json: Example data for rendering

Generate Go code (optional)
1) Run the compiler against the templates:
   go run ./cmd/hbc -in examples/showcase/templates -out examples/showcase/templates_gen.go -pkg showcase

   Core helpers are included by default, so no helper flags are needed.
   For custom helpers or overrides, use `-import` + `-helpers` (see README).

2) Use the generated RenderMain/RenderMainString with data.json.

When using map-backed context (`MainContextFromMap(data)` from JSON), `{{#each}}` works with both JSON arrays (e.g. `orders`, `users`) and objects (e.g. `settings`): the compiler generates code that tries slice iteration first, then map iteration.

