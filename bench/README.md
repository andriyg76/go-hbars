Benchmarks

This package is intended for render benchmarks using generated templates.

Generate templates
1) Run hbc to generate bench/templates_gen.go:
   go run ./cmd/hbc -in bench/templates -out bench/templates_gen.go -pkg bench

   Core helpers are included by default; no helper flags are needed.
   For custom helpers or overrides, use `-import` + `-helpers` (see README).

Run benchmarks
go test -tags benchgen -bench . -benchmem ./bench

Latest snapshot (buildBenchData(100, 12, 8))
BenchmarkRenderMain-22: 8736325 ns/op, 2913343 B/op, 84501 allocs/op
BenchmarkRenderSummary-22: 978.5 ns/op, 248 B/op, 15 allocs/op
BenchmarkRenderHelperHeavy-22: 26000 ns/op, 5738 B/op, 234 allocs/op
BenchmarkRenderMainString-22: 12853380 ns/op, 4455987 B/op, 84538 allocs/op
BenchmarkRenderMain_RecreateData-22: 15590277 ns/op, 3961720 B/op, 116724 allocs/op
BenchmarkRenderSummary_RecreateData-22: 4565587 ns/op, 1049090 B/op, 32236 allocs/op
BenchmarkRenderHelperHeavy_RecreateData-22: 4536662 ns/op, 1054642 B/op, 32455 allocs/op

