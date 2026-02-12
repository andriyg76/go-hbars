# Partial AST: structured params (feature branch)

## Done

- **AST**: `ast.Partial` now has `NameOrExpr` (partial name or subexpression) and `Params []PartialParam` (ordered list). `PartialParam` is either `Path` (context reference) or `Hash` (key=value).
- **Parser**: `parsePartialContent()` tokenizes partial content and produces `nameOrExpr` + ordered params; each param is one path ref or one key=value (no merging consecutive hashes).
- **Compiler**: `partialToParts(n *ast.Partial)` converts AST to `([]expr, []hashArg)`; all Partial handling uses it instead of `parseParts(n.Expr)`.
- **Phase JSON**: `ASTNodeDTO` has `NameOrExpr` and `PartialParams []PartialParamDTO` for Partial; serialization/deserialization updated.
- **Tests**: `ast` and `parser` tests updated for new Partial shape.

## Next / TODO

- **Regenerate intermediate outputs**: Run hbc with `-phase1-output phase1_ast.json` etc. in `examples/partials-inheritance/` so `phase1_ast.json`, `phase2a.json`, `phase2b.json`, `phase3_ir.json` reflect the new Partial structure (e.g. `"nameOrExpr": "menu"`, `"params": [{"path": "_shared.menu"]}`).
- **emitPartial multiple context params**: Currently at most one context path is supported (`len(parts) > 2` → error). Support merging multiple path params + hash into one context map (e.g. `{{> menu this _shared.menu menu=final_value }}`).
- **Full test pass**: Run `go test ./...` and fix any remaining failures (e.g. e2e, compiler tests that still assume old Partial or timeout).
