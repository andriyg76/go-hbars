Benchmarks

This package is intended for render benchmarks using generated templates.

Generate templates
1) Run hbc to generate bench/templates_gen.go:
   go run ./cmd/hbc -in bench/templates -out bench/templates_gen.go -pkg bench

   Note: Core helpers are included by default. The above command will work as-is.
   To explicitly select only needed helpers:
   go run ./cmd/hbc -in bench/templates -out bench/templates_gen.go -pkg bench \
     -no-core-helpers \
     -import github.com/andriyg76/go-hbars/helpers/handlebars \
     -helpers Now,Capitalize,Lower,Upper,And,Gt,ToFixed,Length,FormatNumber,Add,Default,First,Last,FormatDate

Run benchmarks
go test -tags benchgen -bench . -benchmem ./bench

