package compiler

import (
	"strings"
	"testing"

	"github.com/andriyg76/go-hbars/helpers"
)

func TestCompileTemplates_GeneratesFunctions(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hello {{name}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "func RenderMain(w io.Writer, data any) error") {
		t.Fatalf("missing RenderMain writer signature")
	}
	if !strings.Contains(src, "func RenderMainString(data any) (string, error)") {
		t.Fatalf("missing RenderMainString wrapper")
	}
	if !strings.Contains(src, "runtime.WriteEscaped") {
		t.Fatalf("missing runtime.WriteEscaped call")
	}
}

func TestCompileTemplates_HelperDirectCall(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{upper name}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	if !strings.Contains(string(code), "Upper(ctx") {
		t.Fatalf("expected Upper helper call in generated code")
	}
}

func TestCompileTemplates_HelperHashArgs(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{upper name foo=\"bar\" count=2}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "runtime.Hash") {
		t.Fatalf("expected runtime.Hash in generated code")
	}
	if !strings.Contains(src, "\"foo\"") || !strings.Contains(src, "\"count\"") {
		t.Fatalf("expected hash keys in generated code")
	}
}

func TestCompileTemplates_Subexpressions(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{upper (lower name)}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
			"lower": {Ident: "Lower"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "Lower(ctx") {
		t.Fatalf("expected Lower helper call in generated code")
	}
	if !strings.Contains(src, "Upper(ctx") {
		t.Fatalf("expected Upper helper call in generated code")
	}
}

func TestCompileTemplates_MissingHelper(t *testing.T) {
	_, err := CompileTemplates(map[string]string{
		"main": "{{upper name}}",
	}, Options{PackageName: "templates"})
	if err == nil || !strings.Contains(err.Error(), "helper \"upper\"") {
		t.Fatalf("expected missing helper error, got %v", err)
	}
}

func TestCompileTemplates_MissingPartial(t *testing.T) {
	_, err := CompileTemplates(map[string]string{
		"main": "{{> header}}",
	}, Options{PackageName: "templates"})
	if err == nil || !strings.Contains(err.Error(), "partial \"header\"") {
		t.Fatalf("expected missing partial error, got %v", err)
	}
}

func TestCompileTemplates_BlockHelpers(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{#if ok}}Yes{{else}}{{#with user}}{{name}}{{/with}}{{/if}}{{#each items}}{{name}}{{/each}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "runtime.IsTruthy") {
		t.Fatalf("expected runtime.IsTruthy in generated code")
	}
	if !strings.Contains(src, "runtime.Iterate") {
		t.Fatalf("expected runtime.Iterate in generated code")
	}
}

func TestCompileTemplates_IncludeZero(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{#if count includeZero=true}}zero{{else}}nope{{/if}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "runtime.IncludeZeroTruthy") {
		t.Fatalf("expected runtime.IncludeZeroTruthy in generated code when includeZero=true, got:\n%s", src)
	}
	// Without includeZero, must use IsTruthy
	code2, err := CompileTemplates(map[string]string{
		"main": "{{#if count}}zero{{else}}nope{{/if}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	if strings.Contains(string(code2), "IncludeZeroTruthy") {
		t.Fatalf("expected IsTruthy (not IncludeZeroTruthy) when includeZero is not set")
	}
}

func TestCompileTemplates_BlockParams(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{#each items as |item idx|}}{{item}}:{{idx}}{{/each}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "ctx.WithScope(item.Value") {
		t.Fatalf("expected WithScope for each block params")
	}
	if !strings.Contains(src, "\"item\"") {
		t.Fatalf("expected item block param in generated code")
	}
	if !strings.Contains(src, "\"idx\"") {
		t.Fatalf("expected idx block param in generated code")
	}
}

func TestCompileTemplates_DynamicPartial(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main":   "{{> (lookup . \"partial\")}}",
		"header": "<h1>{{title}}</h1>",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"lookup": {Ident: "Lookup"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "partials[") {
		t.Fatalf("expected dynamic partial lookup")
	}
	if !strings.Contains(src, "MissingPartial") {
		t.Fatalf("expected MissingPartial error handling")
	}
}
func TestCompileTemplates_UnknownBlock(t *testing.T) {
	// Unknown block (no registered helper) compiles as universal section, not an error
	code, err := CompileTemplates(map[string]string{
		"main": "{{#noop}}ignored{{/noop}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("expected success (universal section), got %v", err)
	}
	if !strings.Contains(string(code), "WithScope") {
		t.Fatalf("expected WithScope (universal section) in generated code")
	}
}

func TestCompileTemplates_UniversalSection(t *testing.T) {
	// {{#date}}...{{/date}} and {{#foo}}...{{/foo}} with no helper => compiled as section (with-like)
	code, err := CompileTemplates(map[string]string{
		"main": `{{#date}}x{{/date}}{{#foo}}y{{else}}n{{/foo}}`,
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	// Section semantics: resolve name, if truthy WithScope(val), else inverse
	if !strings.Contains(src, "WithScope") {
		t.Fatalf("expected WithScope in generated code for universal section")
	}
	if !strings.Contains(src, "IsTruthy") {
		t.Fatalf("expected IsTruthy in generated code for universal section")
	}
}

func TestCompileTemplates_UniversalSection_HelperWins(t *testing.T) {
	// When a name is registered as block helper, it is used — not section
	reg := helpers.Registry()
	compilerHelpers := make(map[string]HelperRef, len(reg))
	for name, ref := range reg {
		compilerHelpers[name] = HelperRef{ImportPath: ref.ImportPath, Ident: ref.Ident}
	}
	code, err := CompileTemplates(map[string]string{
		"main": `{{#block "x"}}default{{/block}}`,
	}, Options{PackageName: "templates", Helpers: compilerHelpers})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "runtime.Block") {
		t.Fatalf("expected runtime.Block when block is registered helper, got section fallback")
	}
}

func TestCompileTemplates_BlockPartialHelpers(t *testing.T) {
	// Layout block helpers: {{#block "name"}}default{{/block}} and {{#partial "name"}}body{{/partial}}
	reg := helpers.Registry()
	compilerHelpers := make(map[string]HelperRef, len(reg))
	for name, ref := range reg {
		compilerHelpers[name] = HelperRef{ImportPath: ref.ImportPath, Ident: ref.Ident}
	}
	code, err := CompileTemplates(map[string]string{
		"main": `{{#partial "header"}}<title>X</title>{{/partial}}{{#block "header"}}default{{/block}}`,
	}, Options{PackageName: "templates", Helpers: compilerHelpers})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "runtime.Block") {
		t.Fatalf("expected runtime.Block in generated code for {{#block}}, got:\n%s", src)
	}
	if !strings.Contains(src, "runtime.Partial") {
		t.Fatalf("expected runtime.Partial in generated code for {{#partial}}, got:\n%s", src)
	}
}

func TestCompileTemplates_DuplicateIdentifiers(t *testing.T) {
	_, err := CompileTemplates(map[string]string{
		"a-b": "one",
		"a_b": "two",
	}, Options{PackageName: "templates"})
	if err == nil || !strings.Contains(err.Error(), "map to") {
		t.Fatalf("expected duplicate identifier error, got %v", err)
	}
}

func TestCompileTemplates_InlineLiteralArgs(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{upper \"Ada\" 3 true null}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if strings.Contains(src, "EvalArg") {
		t.Fatalf("expected literals to be inlined without EvalArg")
	}
	if !strings.Contains(src, "int64(3)") {
		t.Fatalf("expected int64 literal for numeric arg")
	}
}

func TestCompileTemplates_PrebuildLiteralHash(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{upper name foo=\"bar\" count=2}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
		},
	})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "staticHash") {
		t.Fatalf("expected static hash map for literal values")
	}
	if strings.Contains(src, ":= runtime.Hash") {
		t.Fatalf("expected no per-call hash allocation for literal hash")
	}
}

func TestCompileTemplates_DuplicateHashKeys(t *testing.T) {
	_, err := CompileTemplates(map[string]string{
		"main": "{{upper foo=1 foo=2}}",
	}, Options{
		PackageName: "templates",
		Helpers: map[string]HelperRef{
			"upper": {Ident: "Upper"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate hash key") {
		t.Fatalf("expected duplicate hash key error, got %v", err)
	}
}

func TestCompileTemplates_ConstantFoldIf(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{#if true}}Yes{{else}}No{{/if}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if strings.Contains(src, "IsTruthy") {
		t.Fatalf("expected constant-folded if to avoid IsTruthy")
	}
	if strings.Contains(src, "\"No\"") {
		t.Fatalf("expected else branch to be removed for constant true")
	}
}

func TestCompileTemplates_ConstantFoldWith(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{#with \"Ada\"}}{{this}}{{/with}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if strings.Contains(src, "IsTruthy") {
		t.Fatalf("expected constant-folded with to avoid IsTruthy")
	}
	if !strings.Contains(src, "WithScope(\"Ada\"") {
		t.Fatalf("expected with block to use literal scope value")
	}
}

func TestCompileTemplates_PreparsedPaths(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hello {{user.name}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "ResolvePathValueParsed") {
		t.Fatalf("expected pre-parsed path resolution")
	}
	if !strings.Contains(src, "staticPath") {
		t.Fatalf("expected static parsed path declaration")
	}
}
