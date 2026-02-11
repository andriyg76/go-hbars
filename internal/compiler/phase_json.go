package compiler

import (
	"encoding/json"

	"github.com/andriyg76/go-hbars/internal/ast"
)

// ASTNodeDTO is a JSON-serializable representation of ast.Node.
type ASTNodeDTO struct {
	Kind   string       `json:"kind"`
	Value  string       `json:"value,omitempty"`
	Expr   string       `json:"expr,omitempty"`
	Raw    bool         `json:"raw,omitempty"`
	Name   string       `json:"name,omitempty"`
	Args   string       `json:"args,omitempty"`
	Params []string     `json:"params,omitempty"`
	Body   []ASTNodeDTO `json:"body,omitempty"`
	Else   []ASTNodeDTO `json:"else,omitempty"`
}

func astNodesToDTO(nodes []ast.Node) []ASTNodeDTO {
	out := make([]ASTNodeDTO, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, astNodeToDTO(n))
	}
	return out
}

func astNodeToDTO(n ast.Node) ASTNodeDTO {
	switch node := n.(type) {
	case *ast.Text:
		return ASTNodeDTO{Kind: "Text", Value: node.Value}
	case *ast.Mustache:
		return ASTNodeDTO{Kind: "Mustache", Expr: node.Expr, Raw: node.Raw}
	case *ast.Partial:
		return ASTNodeDTO{Kind: "Partial", Expr: node.Expr}
	case *ast.Block:
		return ASTNodeDTO{
			Kind:   "Block",
			Name:   node.Name,
			Args:   node.Args,
			Params: node.Params,
			Body:   astNodesToDTO(node.Body),
			Else:   astNodesToDTO(node.Else),
		}
	default:
		return ASTNodeDTO{Kind: "unknown"}
	}
}

func astDTOToNodes(dtos []ASTNodeDTO) ([]ast.Node, error) {
	out := make([]ast.Node, 0, len(dtos))
	for _, d := range dtos {
		n, err := astDTOToNode(d)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func astDTOToNode(d ASTNodeDTO) (ast.Node, error) {
	switch d.Kind {
	case "Text":
		return &ast.Text{Value: d.Value}, nil
	case "Mustache":
		return &ast.Mustache{Expr: d.Expr, Raw: d.Raw}, nil
	case "Partial":
		return &ast.Partial{Expr: d.Expr}, nil
	case "Block":
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		els, err := astDTOToNodes(d.Else)
		if err != nil {
			return nil, err
		}
		return &ast.Block{Name: d.Name, Args: d.Args, Params: d.Params, Body: body, Else: els}, nil
	default:
		return &ast.Text{Value: ""}, nil
	}
}

// Phase1ResultDTO is the JSON shape for Phase1Result.
type Phase1ResultDTO struct {
	Templates     map[string][]ASTNodeDTO `json:"templates"`
	Names         []string               `json:"names"`
	FuncNames     map[string]string      `json:"funcNames"`
	HelperExprs   map[string]string      `json:"helperExprs,omitempty"`
	UsedHelpers   []string               `json:"usedHelpers,omitempty"`
	TemplateFiles map[string]string      `json:"templateFiles,omitempty"`
}

// ToDTO converts Phase1Result to Phase1ResultDTO for JSON.
func (p *Phase1Result) ToDTO() Phase1ResultDTO {
	templates := make(map[string][]ASTNodeDTO, len(p.Templates))
	for t, nodes := range p.Templates {
		templates[string(t)] = astNodesToDTO(nodes)
	}
	names := make([]string, 0, len(p.Names))
	for _, t := range p.Names {
		names = append(names, string(t))
	}
	funcNames := make(map[string]string, len(p.FuncNames))
	for t, id := range p.FuncNames {
		funcNames[string(t)] = string(id)
	}
	helperExprs := make(map[string]string, len(p.HelperExprs))
	for h, expr := range p.HelperExprs {
		helperExprs[string(h)] = expr
	}
	usedHelpers := make([]string, 0, len(p.UsedHelpers))
	for _, h := range p.UsedHelpers {
		usedHelpers = append(usedHelpers, string(h))
	}
	templateFiles := make(map[string]string, len(p.TemplateFiles))
	for t, path := range p.TemplateFiles {
		templateFiles[string(t)] = path
	}
	return Phase1ResultDTO{
		Templates:     templates,
		Names:         names,
		FuncNames:     funcNames,
		HelperExprs:   helperExprs,
		UsedHelpers:   usedHelpers,
		TemplateFiles: templateFiles,
	}
}

// Phase2aResultDTO is the JSON shape for Phase2aResult.
type Phase2aResultDTO struct {
	PartialParamTypes map[string]string   `json:"partialParamTypes"`
	TypeSet           map[string][]string `json:"typeSet"`
	CanonicalType     map[string]string   `json:"canonicalType"`
	PrimaryCaller     map[string]string   `json:"primaryCaller"`
	FuncNames         map[string]string   `json:"funcNames"`
}

// ToDTO converts Phase2aResult to Phase2aResultDTO for JSON.
func (p *Phase2aResult) ToDTO() Phase2aResultDTO {
	partialParamTypes := make(map[string]string, len(p.PartialParamTypes))
	for k, v := range p.PartialParamTypes {
		partialParamTypes[string(k)] = string(v)
	}
	typeSet := make(map[string][]string, len(p.TypeSet))
	for k, set := range p.TypeSet {
		vals := make([]string, 0, len(set))
		for ct := range set {
			vals = append(vals, string(ct))
		}
		typeSet[string(k)] = vals
	}
	canonicalType := make(map[string]string, len(p.CanonicalType))
	for k, v := range p.CanonicalType {
		canonicalType[string(k)] = string(v)
	}
	primaryCaller := make(map[string]string, len(p.PrimaryCaller))
	for k, v := range p.PrimaryCaller {
		primaryCaller[string(k)] = string(v)
	}
	funcNames := make(map[string]string, len(p.FuncNames))
	for k, v := range p.FuncNames {
		funcNames[string(k)] = string(v)
	}
	return Phase2aResultDTO{
		PartialParamTypes: partialParamTypes,
		TypeSet:           typeSet,
		CanonicalType:     canonicalType,
		PrimaryCaller:     primaryCaller,
		FuncNames:         funcNames,
	}
}

// TypeNodeDTO is the JSON shape for typeNode.
type TypeNodeDTO struct {
	Fields    map[string]TypeNodeDTO `json:"fields,omitempty"`
	SliceElem *TypeNodeDTO           `json:"sliceElem,omitempty"`
	IsSlice   bool                   `json:"isSlice,omitempty"`
}

func typeNodeToDTO(n *typeNode) TypeNodeDTO {
	if n == nil {
		return TypeNodeDTO{}
	}
	d := TypeNodeDTO{IsSlice: n.isSlice}
	if n.fields != nil {
		d.Fields = make(map[string]TypeNodeDTO, len(n.fields))
		for k, v := range n.fields {
			d.Fields[k] = typeNodeToDTO(v)
		}
	}
	if n.sliceElem != nil {
		se := typeNodeToDTO(n.sliceElem)
		d.SliceElem = &se
	}
	return d
}

// Phase2bResultDTO is the JSON shape for Phase2bResult (type trees + metadata).
type Phase2bResultDTO struct {
	TypeSet          map[string][]string    `json:"typeSet"`
	TypeTrees        map[string]TypeNodeDTO `json:"typeTrees"`
	EmitOrder        []string               `json:"emitOrder"`
	NeedFmt          bool                   `json:"needFmt"`
	UseLayoutBlocks  bool                   `json:"useLayoutBlocks"`
}

// ImportSpecJSON is the JSON shape for an import (path + optional name).
type ImportSpecJSON struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

// Phase3ResultDTO is the JSON shape for Phase3Result (IR).
type Phase3ResultDTO struct {
	PackageName        string              `json:"packageName"`
	RuntimeImport      string              `json:"runtimeImport"`
	EmitOrder          []string             `json:"emitOrder"`
	ContextIfaces      []IRContextIface     `json:"contextIfaces"`
	ContextData        []IRContextData      `json:"contextData"`
	Partials           []IRPartial          `json:"partials"`
	RenderFuncs        []IRRenderFunc       `json:"renderFuncs"`
	UseLayoutBlocks    bool                 `json:"useLayoutBlocks"`
	GenerateBootstrap  bool                 `json:"generateBootstrap"`
	HelperImports      []ImportSpecJSON     `json:"helperImports"`
}

// ToDTO converts Phase3Result to Phase3ResultDTO for JSON.
func (p *Phase3Result) ToDTO() Phase3ResultDTO {
	emitOrder := make([]string, 0, len(p.EmitOrder))
	for _, t := range p.EmitOrder {
		emitOrder = append(emitOrder, string(t))
	}
	imports := make([]ImportSpecJSON, 0, len(p.HelperImports))
	for _, imp := range p.HelperImports {
		imports = append(imports, ImportSpecJSON{Path: imp.path, Name: imp.name})
	}
	return Phase3ResultDTO{
		PackageName:       p.PackageName,
		RuntimeImport:     p.RuntimeImport,
		EmitOrder:         emitOrder,
		ContextIfaces:     p.ContextIfaces,
		ContextData:       p.ContextData,
		Partials:          p.Partials,
		RenderFuncs:       p.RenderFuncs,
		UseLayoutBlocks:   p.UseLayoutBlocks,
		GenerateBootstrap: p.GenerateBootstrap,
		HelperImports:     imports,
	}
}

// MarshalJSON for Phase3Result.
func (p *Phase3Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToDTO())
}

// ToDTO converts Phase2bResult to Phase2bResultDTO for JSON.
func (p *Phase2bResult) ToDTO() Phase2bResultDTO {
	typeTrees := make(map[string]TypeNodeDTO, len(p.TypeTrees))
	for t, tree := range p.TypeTrees {
		typeTrees[string(t)] = typeNodeToDTO(tree)
	}
	emitOrder := make([]string, 0, len(p.EmitOrder))
	for _, t := range p.EmitOrder {
		emitOrder = append(emitOrder, string(t))
	}
	typeSet := make(map[string][]string)
	if p.Phase2a != nil {
		for k, set := range p.Phase2a.TypeSet {
			vals := make([]string, 0, len(set))
			for ct := range set {
				vals = append(vals, string(ct))
			}
			typeSet[string(k)] = vals
		}
	}
	return Phase2bResultDTO{
		TypeSet:         typeSet,
		TypeTrees:       typeTrees,
		EmitOrder:       emitOrder,
		NeedFmt:         p.NeedFmt,
		UseLayoutBlocks: p.UseLayoutBlocks,
	}
}

// MarshalJSON for Phase1Result so phase output can encode it via DTO.
func (p *Phase1Result) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToDTO())
}

// MarshalJSON for Phase2aResult.
func (p *Phase2aResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToDTO())
}

// MarshalJSON for Phase2bResult.
func (p *Phase2bResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.ToDTO())
}
