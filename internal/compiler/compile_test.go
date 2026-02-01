package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestCompileTemplates_CompilerVersion(t *testing.T) {
	code, err := CompileTemplates(map[string]string{
		"main": "Hi",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "// Compiler version: ") {
		t.Fatalf("expected Compiler version comment in generated code, got:\n%s", src)
	}
	if !strings.Contains(src, "// Compiler version: "+Version) {
		t.Fatalf("expected Compiler version %q in generated code, got:\n%s", Version, src)
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
	if !strings.Contains(src, "// Compiler version: ") {
		t.Fatalf("expected Compiler version comment in generated code, got:\n%s", src)
	}
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

func TestCompileTemplates_ErrorIncludesFileNameAndLine(t *testing.T) {
	// Parse error with TemplateFiles: should get "path:line: message"
	_, err := CompileTemplates(
		map[string]string{"main": "{{!--"},
		Options{PackageName: "templates", TemplateFiles: map[string]string{"main": "templates/main.hbs"}},
	)
	if err == nil {
		t.Fatalf("expected parse error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "templates/main.hbs") {
		t.Fatalf("expected error to include file path, got: %s", msg)
	}
	if !strings.Contains(msg, ":1:") {
		t.Fatalf("expected error to include line (e.g. :1:), got: %s", msg)
	}
	// Codegen error (missing partial) with TemplateFiles: should get "path: message"
	_, err = CompileTemplates(
		map[string]string{"main": "{{> header}}"},
		Options{PackageName: "templates", TemplateFiles: map[string]string{"main": "main.hbs"}},
	)
	if err == nil {
		t.Fatalf("expected missing partial error")
	}
	if !strings.Contains(err.Error(), "main.hbs") {
		t.Fatalf("expected error to include file path, got: %s", err.Error())
	}
}

func TestCompileTemplates_GeneratedCodeIncludesSourceRef(t *testing.T) {
	// When TemplateFiles is set, generated code should include "// from path:1" for contexts and render functions.
	code, err := CompileTemplates(
		map[string]string{"main": "Hello {{name}}", "header": "<h1>{{title}}</h1>"},
		Options{
			PackageName:   "templates",
			TemplateFiles: map[string]string{"main": "templates/main.hbs", "header": "templates/header.hbs"},
		},
	)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "// from templates/main.hbs:1") {
		t.Fatalf("expected generated code to contain // from templates/main.hbs:1, got (excerpt): %s", src[:min(800, len(src))])
	}
	if !strings.Contains(src, "// from templates/header.hbs:1") {
		t.Fatalf("expected generated code to contain // from templates/header.hbs:1")
	}
	// Should appear before context types and before render functions
	if idx := strings.Index(src, "// from templates/main.hbs:1"); idx < 0 || idx > strings.Index(src, "type MainContext") {
		t.Fatalf("// from ... should appear before type MainContext")
	}
	if idx := strings.Index(src, "// from templates/main.hbs:1"); idx < 0 || idx > strings.Index(src, "func renderMain(") {
		t.Fatalf("// from ... should appear before func renderMain (or a second occurrence before it)")
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

func TestCompileTemplates_EachOverCollection(t *testing.T) {
	// {{#each orders}} with block params: compiler emits iteration (range or []any/map fallback)
	code, err := CompileTemplates(map[string]string{
		"main": "{{#each orders as |order idx|}}{{order.id}}{{/each}}",
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates error: %v", err)
	}
	src := string(code)
	if !strings.Contains(src, "range") {
		t.Fatalf("expected range (or slice/map iteration) for each block")
	}
	if !strings.Contains(src, "Orders()") && !strings.Contains(src, "orders") {
		t.Fatalf("expected Orders() or orders access for each collection")
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

func TestCompileTemplates_TemplateNameWithSlash(t *testing.T) {
	// Template names like "layout/header" (from layout/header.hbs) must produce valid Go identifiers, not "Layout/header()".
	code, err := CompileTemplates(map[string]string{
		"main":          `{{title}}{{> layout/header}}`,
		"layout/header": `<h1>{{title}}</h1>`,
	}, Options{PackageName: "templates"})
	if err != nil {
		t.Fatalf("CompileTemplates: %v", err)
	}
	src := string(code)
	if strings.Contains(src, "Layout/header()") || strings.Contains(src, "/header()") {
		t.Fatalf("generated code must not contain '/' in method names (invalid Go); got slash in identifier")
	}
	if !strings.Contains(src, "LayoutHeader()") {
		t.Fatalf("expected LayoutHeader() (PascalCase from layout/header) in generated code, got: %s", src[min(1500, len(src)):])
	}
	mustBuildGeneratedCode(t, code, "templates")
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

// TestCompileTemplates_PartialContextRules checks partial context rules:
// - {{> name}} (no args, no hash) => current context passed as-is (renderXxx(data, ...)).
// - {{> name note="x"}} (only hash) => context = hash + keys used in partial from current scope (MergePartialContext(nil, hash) then add scope keys).
func TestCompileTemplates_PartialContextRules(t *testing.T) {
	// Only hash + partial uses a key not in hash => that key comes from current scope.
	t.Run("only_hash_plus_partial_keys", func(t *testing.T) {
		code, err := CompileTemplates(map[string]string{
			"main":   `{{title}}{{> footer note="thanks"}}`,
			"footer": `{{note}} {{title}}`,
		}, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)
		if !strings.Contains(src, "runtime.MergePartialContext(nil,") {
			t.Errorf("only-hash partial should use MergePartialContext(nil, hash); got snippet: %s", grepOne(src, "MergePartialContext"))
		}
		// Footer uses "title"; not in hash => should add from scope.
		if !strings.Contains(src, `["title"]`) && !strings.Contains(src, "[\"title\"]") {
			t.Errorf("partial uses {{title}} but hash has only note; expected partialCtx[\"title\"] = ... from scope")
		}
	})
	// No args, no hash => current context (renderXxx(data, ...)), not partials map.
	t.Run("no_args_current_context", func(t *testing.T) {
		code, err := CompileTemplates(map[string]string{
			"main":   `{{> header}}`,
			"header": `<h1>{{title}}</h1>`,
		}, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)
		if !strings.Contains(src, "renderHeader(data,") {
			t.Errorf("{{> header}} with no args should call renderHeader(data, ...); got: %s", grepOne(src, "renderHeader"))
		}
	})
}

// TestCompileTemplates_CanonicalPartialContext checks that when multiple root templates
// include the same partial with the same scope (e.g. {{> menu}}), the compiler emits
// one canonical context interface per partial (e.g. MenuContext) and all callers use
// the canonical type (e.g. DefaultContext.Menu() MenuContext, FirstpageContext.Menu() MenuContext).
// No primary-embedding types (DefaultMenuContext, etc.) are emitted.
func TestCompileTemplates_CanonicalPartialContext(t *testing.T) {
	t.Run("two_roots_share_partial", func(t *testing.T) {
		tmpls := map[string]string{
			"default":  `Page: {{title}}{{> menu}}`,
			"firstpage": `First: {{title}}{{> menu}}`,
			"menu":      `<nav>{{title}}</nav>`,
		}
		code, err := CompileTemplates(tmpls, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)

		if !strings.Contains(src, "type MenuContext interface") {
			t.Errorf("expected canonical type MenuContext interface; got:\n%s", grepOne(src, "Context interface"))
		}
		if strings.Contains(src, "type DefaultMenuContext interface") {
			t.Errorf("must not emit primary-embedding type DefaultMenuContext; use canonical MenuContext only")
		}
		if !strings.Contains(src, "DefaultContext interface") {
			t.Errorf("expected DefaultContext; got:\n%s", grepOne(src, "DefaultContext"))
		}
		if !strings.Contains(src, "MenuContext") {
			t.Errorf("expected canonical MenuContext (embedded or Menu() MenuContext); got:\n%s", grepOne(src, "Menu"))
		}
		if !strings.Contains(src, "FirstpageContext interface") {
			t.Errorf("expected FirstpageContext; got:\n%s", grepOne(src, "FirstpageContext"))
		}
		if strings.Contains(src, "type FirstpageMenuContext interface") {
			t.Errorf("must not emit template-specific FirstpageMenuContext; use canonical MenuContext only")
		}
	})

	// Three roots (default, firstpage, about) all include menu; all use canonical MenuContext.
	t.Run("three_roots_share_partial", func(t *testing.T) {
		tmpls := map[string]string{
			"default":  `Default: {{title}}{{> menu}}`,
			"firstpage": `First: {{title}}{{> menu}}`,
			"about":    `About: {{title}}{{> menu}}`,
			"menu":     `<nav>{{title}}</nav>`,
		}
		code, err := CompileTemplates(tmpls, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)

		if !strings.Contains(src, "type MenuContext interface") {
			t.Errorf("expected canonical MenuContext; got:\n%s", grepOne(src, "MenuContext"))
		}
		if strings.Contains(src, "type AboutMenuContext interface") {
			t.Errorf("must not emit primary-embedding AboutMenuContext; use MenuContext only")
		}
		if strings.Contains(src, "type FirstpageMenuContext interface") {
			t.Errorf("FirstpageMenuContext must not exist; use MenuContext")
		}
		if strings.Contains(src, "type DefaultMenuContext interface") {
			t.Errorf("DefaultMenuContext must not exist; use MenuContext")
		}
		if !strings.Contains(src, "MenuContext") {
			t.Errorf("expected canonical MenuContext; got:\n%s", grepOne(src, "Menu"))
		}
	})

	// Page "inherits" default: page includes default with same scope, and both include menu.
	// default is a shared partial (only caller is page); menu is shared (default and page both include it).
	// All use canonical types; PageContext embeds DefaultContext and has Menu() MenuContext.
	t.Run("page_inherits_default_both_include_menu", func(t *testing.T) {
		tmpls := map[string]string{
			"default":  `Layout: {{title}}{{> menu}}`,
			"page":     `Page: {{title}}{{> menu}}{{> default}}`,
			"menu":     `<nav>{{title}}</nav>`,
		}
		code, err := CompileTemplates(tmpls, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)

		if !strings.Contains(src, "type MenuContext interface") {
			t.Errorf("expected canonical MenuContext; got:\n%s", grepOne(src, "MenuContext"))
		}
		if !strings.Contains(src, "type PageContext interface") {
			t.Errorf("expected PageContext (page is root); got:\n%s", grepOne(src, "PageContext"))
		}
		if strings.Contains(src, "type PageMenuContext interface") {
			t.Errorf("PageMenuContext must not exist; use Menu() MenuContext")
		}
		if !strings.Contains(src, "MenuContext") {
			t.Errorf("expected canonical MenuContext; got:\n%s", grepOne(src, "Menu"))
		}
	})

	// Nested partials: menu is shared and includes menuItem; canonical types at both levels.
	t.Run("nested_partials_shared", func(t *testing.T) {
		tmpls := map[string]string{
			"default":  `{{title}}{{> menu}}`,
			"firstpage": `{{title}}{{> menu}}`,
			"menu":      `<nav>{{title}}{{#each items}}{{> menuItem}}{{/each}}</nav>`,
			"menuItem":  `<span>{{label}}</span>`,
		}
		code, err := CompileTemplates(tmpls, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		src := string(code)

		if !strings.Contains(src, "type MenuContext interface") {
			t.Errorf("expected canonical MenuContext; got:\n%s", grepOne(src, "MenuContext"))
		}
		// Canonical item context under menu (slice items): MenuItemsItemContext is used (type or FromMap).
		if !strings.Contains(src, "MenuItemsItemContext") {
			t.Errorf("expected canonical MenuItemsItemContext for nested partial under menu; got:\n%s", grepOne(src, "MenuItemsItem"))
		}
		if strings.Contains(src, "type FirstpageMenuContext interface") {
			t.Errorf("FirstpageMenuContext must not exist; use MenuContext")
		}
		if strings.Contains(src, "type FirstpageMenuItemsItemContext interface") {
			t.Errorf("nested partial context must be canonical MenuItemsItemContext, not FirstpageMenuItemsItemContext")
		}
	})
}

func grepOne(src, sub string) string {
	if i := strings.Index(src, sub); i >= 0 {
		end := i + len(sub) + 80
		if end > len(src) {
			end = len(src)
		}
		return src[i:end]
	}
	return "(not found)"
}

// repoRootForCompileTest returns the repository root. Tests run with cwd = package directory (internal/compiler).
func repoRootForCompileTest(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	return dir
}

// mustBuildGeneratedCode writes the generated code to a temp module and runs go build.
// If the build fails, the test fails with the build output.
func mustBuildGeneratedCode(t *testing.T, code []byte, packageName string) {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := strings.ReplaceAll(repoRootForCompileTest(t), "\\", "/")
	writeFile := func(path, body string) {
		full := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	writeFile("go.mod", `module test-compile

go 1.24

require github.com/andriyg76/go-hbars v0.0.0

replace github.com/andriyg76/go-hbars => `+repoPath+`
`)
	// Copy go.sum from repo so transitive deps of the replaced module resolve.
	if sum, err := os.ReadFile(filepath.Join(repoRootForCompileTest(t), "go.sum")); err == nil {
		writeFile("go.sum", string(sum))
	}
	writeFile(packageName+"/templates_gen.go", string(code))
	writeFile("main.go", `package main

import _ "test-compile/`+packageName+`"

func main() {}
`)

	if out, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	cmd := exec.Command("go", "build", "-mod=mod", ".")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated code did not compile:\n%s", out)
	}
}

// TestCompileTemplates_GeneratedCodeCompiles verifies that representative generated outputs actually compile.
func TestCompileTemplates_GeneratedCodeCompiles(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		code, err := CompileTemplates(map[string]string{
			"main": "Hello {{name}}",
		}, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		mustBuildGeneratedCode(t, code, "templates")
	})
	t.Run("with_partials", func(t *testing.T) {
		code, err := CompileTemplates(map[string]string{
			"main":   `{{title}}{{> header}}`,
			"header": `<h1>{{title}}</h1>`,
		}, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		mustBuildGeneratedCode(t, code, "templates")
	})
	t.Run("canonical_partial_context", func(t *testing.T) {
		code, err := CompileTemplates(map[string]string{
			"default":  `Page: {{title}}{{> menu}}`,
			"firstpage": `First: {{title}}{{> menu}}`,
			"menu":     `<nav>{{title}}</nav>`,
		}, Options{PackageName: "templates"})
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		mustBuildGeneratedCode(t, code, "templates")
	})
}

