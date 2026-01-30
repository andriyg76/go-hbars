package compiler

import (
	"strings"
	"testing"
)

func TestCompileTemplates_GeneratesFunctions(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hello {{name}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "func RenderMain(w io.Writer, data MainContext) error") {
		t.Fatalf("missing RenderMain writer signature")
	}
	if !strings.Contains(src, "func RenderMainString(data MainContext) (string, error)") {
		t.Fatalf("missing RenderMainString wrapper")
	}
	if !strings.Contains(src, "runtime.WriteEscaped") {
		t.Fatalf("missing runtime.WriteEscaped call")
	}
}

func TestCompileTemplates_GeneratorVersion(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hi",
	}, Options{PackageName: "templates", GeneratorVersion: "v0.1.0"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "// Generator version: v0.1.0") {
		t.Fatalf("expected Generator version comment in generated code, got:\n%s", src)
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
	if !strings.Contains(string(code), "Upper(") {
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
	if !strings.Contains(src, "Lower(") {
		t.Fatalf("expected Lower helper call in generated code")
	}
	if !strings.Contains(src, "Upper(") {
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
	// Typed context: each uses range over typed collection, no runtime.Iterate
	if !strings.Contains(src, "range") {
		t.Fatalf("expected range (each) in generated code")
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
	// Typed context: each iterates over typed collection
	if !strings.Contains(src, "range") {
		t.Fatalf("expected range for each block")
	}
	// Block params item/idx appear as scope or loop vars
	if !strings.Contains(src, "item") {
		t.Fatalf("expected item block param in generated code")
	}
	if !strings.Contains(src, "idx") {
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
	if !strings.Contains(src, "MissingPartialOutput") {
		t.Fatalf("expected MissingPartialOutput for dynamic partial when not found")
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
	// Typed context: section uses typed scope, no runtime WithScope
	if !strings.Contains(string(code), "IsTruthy") {
		t.Fatalf("expected IsTruthy (universal section) in generated code")
	}
}

func TestCompileTemplates_LayoutBlocks(t *testing.T) {
	// Layout block/partial: {{#block "name"}}default{{/block}} and {{#partial "name"}}content{{/partial}}
	code, err := CompileTemplates(map[string]string{
		"layout": `<div>{{#block "content"}}default content{{/block}}</div>`,
		"page":   `{{#partial "content"}}<p>page body</p>{{/partial}}{{> layout}}`,
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "*runtime.Blocks") {
		t.Fatalf("expected *runtime.Blocks in generated code when using block/partial")
	}
	if !strings.Contains(src, "RenderPageWithBlocks") {
		t.Fatalf("expected RenderPageWithBlocks in generated code")
	}
	if !strings.Contains(src, "blocks.Get(") {
		t.Fatalf("expected blocks.Get in generated code for {{#block}}")
	}
	if !strings.Contains(src, "blocks.Set(") {
		t.Fatalf("expected blocks.Set in generated code for {{#partial}}")
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
	// Typed context: section uses typed scope (getter + IsTruthy), no runtime WithScope
	if !strings.Contains(src, "IsTruthy") {
		t.Fatalf("expected IsTruthy in generated code for universal section")
	}
	if !strings.Contains(src, "render") {
		t.Fatalf("expected render call in generated code")
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
	// Typed context: helper args are resolved as typed getters/literals, no EvalArg
	if !strings.Contains(src, "Upper(") {
		t.Fatalf("expected Upper helper call in generated code")
	}
	if !strings.Contains(src, "\"Ada\"") {
		t.Fatalf("expected inline literal in generated code")
	}
}

func TestCompileTemplates_ContextInterfaces(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "{{title}}\n{{#with user}}{{name}}{{/with}}\n{{#each items as |item|}}{{item.id}}{{/each}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "// MainContext is the context interface inferred from template \"main\".") {
		t.Fatalf("expected MainContext interface comment in generated code")
	}
	if !strings.Contains(src, "type MainContext interface {") {
		t.Fatalf("expected type MainContext interface in generated code")
	}
	if !strings.Contains(src, "Title() any") {
		t.Fatalf("expected Title() any in MainContext")
	}
	if !strings.Contains(src, "User() MainUserContext") {
		t.Fatalf("expected User() MainUserContext in MainContext")
	}
	if !strings.Contains(src, "Items() []MainItemsItemContext") {
		t.Fatalf("expected Items() []MainItemsItemContext in MainContext")
	}
	if !strings.Contains(src, "type MainUserContext interface {") {
		t.Fatalf("expected nested MainUserContext interface")
	}
	if !strings.Contains(src, "type MainItemsItemContext interface {") {
		t.Fatalf("expected MainItemsItemContext for each element")
	}
	if !strings.Contains(src, "Raw() any") {
		t.Fatalf("expected Raw() any in root context interface")
	}
	if !strings.Contains(src, "type MainContextData struct") {
		t.Fatalf("expected MainContextData map-backed type")
	}
	if !strings.Contains(src, "func MainContextFromMap(m map[string]any) MainContext") {
		t.Fatalf("expected MainContextFromMap constructor")
	}
}

func TestCompileTemplates_PartialsUseFromMap(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main":    "{{title}}{{> header}}",
		"header":  "<h1>{{title}}</h1>",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "func contextMap(ctx any) map[string]any") {
		t.Fatalf("expected contextMap helper in generated code")
	}
	if !strings.Contains(src, "m := contextMap(ctx)") {
		t.Fatalf("expected partials to use contextMap(ctx)")
	}
	if !strings.Contains(src, "MainContextFromMap(m)") {
		t.Fatalf("expected partials to use FromMap(m) instead of type assertion")
	}
	if strings.Contains(src, "ctx.(MainContext)") {
		t.Fatalf("partials must not use type assertion for context")
	}
}

