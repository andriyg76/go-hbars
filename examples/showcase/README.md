Showcase templates

This folder contains example templates and data that exercise most supported
syntax and helpers.

Files
- templates/: Handlebars templates and partials
- data.json: Example data for rendering

Generate Go code (optional)
1) Run the compiler against the templates:
   go run ./cmd/hbc -in examples/showcase/templates -out examples/showcase/templates_gen.go -pkg showcase

   Note: Core helpers are included by default. To use custom helpers or override defaults:
   go run ./cmd/hbc -in examples/showcase/templates -out examples/showcase/templates_gen.go -pkg showcase \
     -import github.com/andriyg76/go-hbars/helpers/handlebars \
     -helpers Upper,Lower,Capitalize,Replace,StripProtocol,StripQuerystring,Size,Keys,Values,Join,Split,FormatDate,Truncate,And,Gt,Default,Lookup,StripTags,Has,Length,Add,FormatNumber,First,Last,Eq,ToFixed,Now

2) Use the generated RenderMain/RenderMainString with data.json.

