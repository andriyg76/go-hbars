package compiler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
	"github.com/andriyg76/go-hbars/internal/parser"
	"github.com/andriyg76/hexerr"
)

// TemplateName identifies a template (e.g. "main", "header", "menu").
// Used as map key for Templates and in partial-relation maps.
type TemplateName string

// GoIdent is the Go identifier for a template (e.g. "Main", "Menu").
// Used for generated function names and as the base for context type names.
type GoIdent string

// ContextTypeName is the Go context interface name (e.g. "MainContext", "MenuContext").
// Used when describing which context type a partial receives or as canonical type for shared partials.
type ContextTypeName string

// Template holds all information for a single template (parsed AST, inferred type tree, and metadata).
// Bundle structures data by template so everything about one template lives in one place.
type Template struct {
	Parsed     []ast.Node // AST for this template
	FuncName   GoIdent    // Go identifier (e.g. "Main")
	TypeTree   *TypeNode  // inferred context type tree for this template
	SourcePath string     // source file path for comments/errors (optional)
}

// Bundle is the intermediate representation after parsing and context inference.
// Per-template data lives in Templates keyed by TemplateName; cross-template relations (partials)
// use typed maps so key/value roles are explicit.
// Use BuildBundle to create a Bundle; use EmitGo to generate Go source from it.
type Bundle struct {
	// Templates: all information for each template, keyed by template name.
	Templates map[TemplateName]*Template
	// Names: sorted template names for stable iteration order.
	Names []TemplateName

	// Cross-template: partial inclusion and shared-partial resolution (all keyed by partial/template name).
	PartialParamTypes map[TemplateName]ContextTypeName     // partial name -> context type (empty = own context)
	TypeSet          map[TemplateName]map[ContextTypeName]bool // partial name -> set of caller context types (same-scope)
	CanonicalType    map[TemplateName]ContextTypeName     // shared partial name -> canonical context type name
	PrimaryCaller    map[TemplateName]TemplateName        // shared partial name -> primary caller template name
}

// BuildOptions configures building the IR (phase 1). Only what is needed for parsing and context inference.
type BuildOptions struct {
	// Helpers maps helper name to implementation; used to distinguish helpers from data paths during inference.
	Helpers map[string]HelperRef
	// TemplateFiles maps template name (key in templates map) to source file path for error messages.
	TemplateFiles map[string]string
}

// BuildBundle parses templates and runs context inference. It returns a Bundle (IR) without generating Go.
// Use this to test or inspect template structure (e.g. type trees, partials) without code generation.
func BuildBundle(templates map[string]string, opts BuildOptions) (*Bundle, error) {
	// Build helper name set for pathCollector (we only need names, not Go expressions).
	helperExprs := make(map[string]string)
	if opts.Helpers != nil {
		for name := range opts.Helpers {
			helperExprs[name] = name
		}
	}

	names := make([]string, 0, len(templates))
	parsed := make(map[string][]ast.Node, len(templates))
	for name, tmpl := range templates {
		nodes, err := parser.Parse(tmpl)
		if err != nil {
			if file := opts.TemplateFiles[name]; file != "" {
				return nil, templateErrorWithFile(file, err)
			}
			return nil, hexerr.Wrapf(err, "compiler: template %q", name)
		}
		parsed[name] = nodes
		names = append(names, name)
	}
	sort.Strings(names)

	funcNames := make(map[string]string, len(names))
	seenFunc := make(map[string]string, len(names))
	for _, name := range names {
		ident := goIdent(name)
		if prev, exists := seenFunc[ident]; exists {
			return nil, hexerr.New(fmt.Sprintf("compiler: templates %q and %q map to %q", prev, name, ident))
		}
		seenFunc[ident] = name
		funcNames[name] = ident
	}

	partialParamTypes := CollectPartialParamTypes(parsed, names, funcNames, helperExprs)
	typeSet := buildPartialTypeSet(parsed, names, funcNames, helperExprs)
	for partialName, set := range typeSet {
		if len(set) > 1 && funcNames[partialName] == "" {
			funcNames[partialName] = goIdent(partialName)
		}
	}
	canonicalType, primaryCaller := SharedPartialInfo(typeSet, names, funcNames)

	typeTrees := make(map[string]*TypeNode)
	for _, name := range names {
		col := newPathCollector(helperExprs)
		col.setParsed(parsed)
		if err := col.collectNodes(parsed[name]); err != nil {
			if file := opts.TemplateFiles[name]; file != "" {
				return nil, templateErrorWithFile(file, err)
			}
			return nil, hexerr.Wrapf(err, "compiler: template %q context inference", name)
		}
		typeTrees[name] = buildTypeTree(col.paths, col.eachFields)
	}

	templateMap := make(map[TemplateName]*Template, len(names))
	for _, name := range names {
		sourcePath := ""
		if opts.TemplateFiles != nil {
			sourcePath = opts.TemplateFiles[name]
		}
		templateMap[TemplateName(name)] = &Template{
			Parsed:     parsed[name],
			FuncName:   GoIdent(funcNames[name]),
			TypeTree:   typeTrees[name],
			SourcePath: sourcePath,
		}
	}
	// Convert partial-relation maps to typed maps.
	ppt := make(map[TemplateName]ContextTypeName, len(partialParamTypes))
	for k, v := range partialParamTypes {
		ppt[TemplateName(k)] = ContextTypeName(v)
	}
	ts := make(map[TemplateName]map[ContextTypeName]bool, len(typeSet))
	for k, set := range typeSet {
		m := make(map[ContextTypeName]bool, len(set))
		for ctx := range set {
			m[ContextTypeName(ctx)] = true
		}
		ts[TemplateName(k)] = m
	}
	ct := make(map[TemplateName]ContextTypeName, len(canonicalType))
	for k, v := range canonicalType {
		ct[TemplateName(k)] = ContextTypeName(v)
	}
	pc := make(map[TemplateName]TemplateName, len(primaryCaller))
	for k, v := range primaryCaller {
		pc[TemplateName(k)] = TemplateName(v)
	}
	namesTyped := make([]TemplateName, len(names))
	for i, n := range names {
		namesTyped[i] = TemplateName(n)
	}
	return &Bundle{
		Templates:         templateMap,
		Names:             namesTyped,
		PartialParamTypes: ppt,
		TypeSet:           ts,
		CanonicalType:     ct,
		PrimaryCaller:     pc,
	}, nil
}

// HasField reports whether the type tree for template name has a field at the given path.
// Path is dot-separated (e.g. "user", "user.name", "items" for a slice).
func (b *Bundle) HasField(templateName, path string) bool {
	t := b.Templates[TemplateName(templateName)]
	if t == nil || t.TypeTree == nil {
		return false
	}
	tree := t.TypeTree
	parts := strings.Split(path, ".")
	cur := tree
	for _, part := range parts {
		if part == "" || cur == nil || cur.Fields == nil {
			return false
		}
		child, ok := cur.Fields[part]
		if !ok {
			return false
		}
		cur = child
	}
	return true
}
