package compiler

import "github.com/andriyg76/go-hbars/internal/ast"

// Derived string types for type-safe maps and slices in phase results and IR.

// Template is a template name (key in templates map, e.g. "main", "layout/header").
type Template string

// PartialName is a partial/template name when used as a partial (may equal Template for root templates).
type PartialName string

// Helper is a helper name (e.g. "upper", "lookup").
type Helper string

// GoIdent is the Go identifier for a template (e.g. "Main", "LayoutHeader").
type GoIdent string

// ContextTypeName is a Go context interface name (e.g. "MainContext").
type ContextTypeName string

// MaxPhase limits how far compilation runs (1 through 4). Phase 2 is split into 2a and 2b internally.
type MaxPhase int

const (
	Phase1  MaxPhase = 1 // Parsing only
	Phase2a MaxPhase = 2 // Partial/context types (2a)
	Phase2b MaxPhase = 3 // Type trees (2b)
	Phase3  MaxPhase = 4 // IR build
	Phase4  MaxPhase = 5 // Go codegen
)

// Phase1Result is the output of phase 1 (parsing).
type Phase1Result struct {
	Templates     map[Template][]ast.Node
	Names         []Template
	FuncNames     map[Template]GoIdent
	HelperExprs   map[Helper]string
	HelperImports []importSpec
	UsedHelpers   []Helper
	TemplateFiles map[Template]string
}

// Phase2aResult is the output of phase 2a (partial/context type analysis).
type Phase2aResult struct {
	PartialParamTypes map[PartialName]ContextTypeName
	HashOnlyPartials  map[string]bool // partial names only ever called with hash (e.g. {{> menu menu=...}})
	TypeSet           map[PartialName]map[ContextTypeName]bool
	CanonicalType     map[PartialName]ContextTypeName
	PrimaryCaller     map[PartialName]Template
	// FuncNames may be extended for shared partials; pass through from Phase1.
	FuncNames map[PartialName]GoIdent
}

// Phase2bResult is the output of phase 2b (type trees per template).
type Phase2bResult struct {
	Phase2a        *Phase2aResult
	TypeTrees      map[Template]*typeNode
	ContextToTmpl  map[ContextTypeName]Template
	EmitOrder      []Template
	NeedFmt        bool
	UseLayoutBlocks bool
}

// Phase3Result is the structured IR (codegen plan) produced by phase 3.
type Phase3Result struct {
	PackageName     string
	RuntimeImport   string
	EmitOrder       []Template
	ContextIfaces   []IRContextIface
	ContextData     []IRContextData
	Partials        []IRPartial
	RenderFuncs     []IRRenderFunc
	UseLayoutBlocks bool
	GenerateBootstrap bool
	HelperImports   []importSpec
}

// IRContextIface describes one context interface to emit.
type IRContextIface struct {
	Template   Template
	GoName     GoIdent
	Methods    []IRIfaceMethod
	Comment    string
}

// IRIfaceMethod is a method on a context interface (field path -> return type).
type IRIfaceMethod struct {
	FieldPath    string
	ReturnType   ContextTypeName
	IsCanonical bool
}

// IRContextData describes one context data struct to emit.
type IRContextData struct {
	Template Template
	GoName   GoIdent
	Fields   []IRDataField
	Comment  string
}

// IRDataField is a field in a context data struct.
type IRDataField struct {
	FieldPath string
	Type      ContextTypeName
}

// IRPartial describes one partial entry in the partials map.
type IRPartial struct {
	Name       Template
	GoName     GoIdent
	CtxType    ContextTypeName
	FromMapFn  string
}

// IRRenderFunc describes one render function (renderX, RenderX, RenderXString, etc.).
type IRRenderFunc struct {
	Template       Template
	GoName         GoIdent
	RootContext    ContextTypeName
	UseLayoutBlocks bool
}
