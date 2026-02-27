package compiler

import (
	"encoding/json"
	"fmt"

	"github.com/andriyg76/go-hbars/internal/ast"
)

// ExprDTO is a JSON-serializable representation of ast.Expr.
type ExprDTO struct {
	Kind   string    `json:"kind"`
	Value  string    `json:"value,omitempty"`
	Callee string    `json:"callee,omitempty"`
	Params []ExprDTO `json:"params,omitempty"`
	Hash   []HashDTO `json:"hash,omitempty"`
}

// HashDTO is a JSON-serializable key=value pair.
type HashDTO struct {
	Key   string  `json:"key"`
	Value ExprDTO `json:"value"`
}

func exprToDTO(e ast.Expr) ExprDTO {
	if e == nil {
		return ExprDTO{Kind: "null"}
	}
	switch v := e.(type) {
	case *ast.PathRef:
		return ExprDTO{Kind: "path", Value: v.Path}
	case *ast.StringLit:
		return ExprDTO{Kind: "string", Value: v.Value}
	case *ast.NumberLit:
		return ExprDTO{Kind: "number", Value: v.Value}
	case *ast.BoolLit:
		val := "false"
		if v.Value {
			val = "true"
		}
		return ExprDTO{Kind: "bool", Value: val}
	case *ast.NullLit:
		return ExprDTO{Kind: "null"}
	case *ast.HelperCall:
		params := make([]ExprDTO, len(v.Params))
		for i, p := range v.Params {
			params[i] = exprToDTO(p)
		}
		hash := make([]HashDTO, len(v.Hash))
		for i, h := range v.Hash {
			hash[i] = HashDTO{Key: h.Key, Value: exprToDTO(h.Value)}
		}
		return ExprDTO{Kind: "call", Callee: v.Callee, Params: params, Hash: hash}
	default:
		return ExprDTO{Kind: "unknown"}
	}
}

func dtoToExpr(d ExprDTO) ast.Expr {
	switch d.Kind {
	case "path":
		return &ast.PathRef{Path: d.Value}
	case "string":
		return &ast.StringLit{Value: d.Value}
	case "number":
		return &ast.NumberLit{Value: d.Value}
	case "bool":
		return &ast.BoolLit{Value: d.Value == "true"}
	case "null":
		return &ast.NullLit{}
	case "call":
		params := make([]ast.Expr, len(d.Params))
		for i, p := range d.Params {
			params[i] = dtoToExpr(p)
		}
		hash := make([]ast.HashEntry, len(d.Hash))
		for i, h := range d.Hash {
			hash[i] = ast.HashEntry{Key: h.Key, Value: dtoToExpr(h.Value)}
		}
		return &ast.HelperCall{Callee: d.Callee, Params: params, Hash: hash}
	default:
		return &ast.NullLit{}
	}
}

func hashEntriesToDTO(entries []ast.HashEntry) []HashDTO {
	if len(entries) == 0 {
		return nil
	}
	out := make([]HashDTO, len(entries))
	for i, h := range entries {
		out[i] = HashDTO{Key: h.Key, Value: exprToDTO(h.Value)}
	}
	return out
}

func dtoToHashEntries(dtos []HashDTO) []ast.HashEntry {
	if len(dtos) == 0 {
		return nil
	}
	out := make([]ast.HashEntry, len(dtos))
	for i, h := range dtos {
		out[i] = ast.HashEntry{Key: h.Key, Value: dtoToExpr(h.Value)}
	}
	return out
}

func exprsToDTO(exprs []ast.Expr) []ExprDTO {
	if len(exprs) == 0 {
		return nil
	}
	out := make([]ExprDTO, len(exprs))
	for i, e := range exprs {
		out[i] = exprToDTO(e)
	}
	return out
}

// ASTNodeDTO is a JSON-serializable representation of ast.Node.
type ASTNodeDTO struct {
	Kind        string       `json:"kind"`
	Value       string       `json:"value,omitempty"`
	Expr        *ExprDTO     `json:"expr,omitempty"`
	Raw         bool         `json:"raw,omitempty"`
	Name        string       `json:"name,omitempty"`
	NameExpr    *ExprDTO     `json:"nameExpr,omitempty"`
	CtxExpr     *ExprDTO     `json:"ctxExpr,omitempty"`
	Unless      bool         `json:"unless,omitempty"`
	Test        *ExprDTO     `json:"test,omitempty"`
	Collection  *ExprDTO     `json:"collection,omitempty"`
	WithValue   *ExprDTO     `json:"withValue,omitempty"`
	Params      []ExprDTO    `json:"params,omitempty"`
	Hash        []HashDTO    `json:"hash,omitempty"`
	BlockParams []string     `json:"blockParams,omitempty"`
	Body        []ASTNodeDTO `json:"body,omitempty"`
	Else        []ASTNodeDTO `json:"else,omitempty"`
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
		e := exprToDTO(node.Expr)
		return ASTNodeDTO{Kind: "Mustache", Expr: &e, Raw: node.Raw}
	case *ast.Partial:
		ne := exprToDTO(node.Name)
		d := ASTNodeDTO{Kind: "Partial", NameExpr: &ne, Hash: hashEntriesToDTO(node.Hash)}
		if node.Ctx != nil {
			ce := exprToDTO(node.Ctx)
			d.CtxExpr = &ce
		}
		return d
	case *ast.IfBlock:
		te := exprToDTO(node.Test)
		return ASTNodeDTO{
			Kind:   "IfBlock",
			Unless: node.Unless,
			Test:   &te,
			Hash:   hashEntriesToDTO(node.Hash),
			Body:   astNodesToDTO(node.Body),
			Else:   astNodesToDTO(node.Else),
		}
	case *ast.WithBlock:
		we := exprToDTO(node.Value)
		return ASTNodeDTO{
			Kind:        "WithBlock",
			WithValue:   &we,
			BlockParams: node.BlockParams,
			Body:        astNodesToDTO(node.Body),
			Else:        astNodesToDTO(node.Else),
		}
	case *ast.EachBlock:
		ce := exprToDTO(node.Collection)
		return ASTNodeDTO{
			Kind:        "EachBlock",
			Collection:  &ce,
			BlockParams: node.BlockParams,
			Body:        astNodesToDTO(node.Body),
			Else:        astNodesToDTO(node.Else),
		}
	case *ast.LayoutBlock:
		return ASTNodeDTO{Kind: "LayoutBlock", Name: node.Name, Body: astNodesToDTO(node.Body)}
	case *ast.LayoutPartial:
		return ASTNodeDTO{Kind: "LayoutPartial", Name: node.Name, Body: astNodesToDTO(node.Body)}
	case *ast.Block:
		return ASTNodeDTO{
			Kind:        "Block",
			Name:        node.Name,
			Params:      exprsToDTO(node.Params),
			Hash:        hashEntriesToDTO(node.Hash),
			BlockParams: node.BlockParams,
			Body:        astNodesToDTO(node.Body),
			Else:        astNodesToDTO(node.Else),
		}
	default:
		return ASTNodeDTO{Kind: fmt.Sprintf("unknown(%T)", n)}
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
		var e ast.Expr
		if d.Expr != nil {
			e = dtoToExpr(*d.Expr)
		}
		return &ast.Mustache{Expr: e, Raw: d.Raw}, nil
	case "Partial":
		var name ast.Expr
		if d.NameExpr != nil {
			name = dtoToExpr(*d.NameExpr)
		}
		var ctx ast.Expr
		if d.CtxExpr != nil {
			ctx = dtoToExpr(*d.CtxExpr)
		}
		return &ast.Partial{Name: name, Ctx: ctx, Hash: dtoToHashEntries(d.Hash)}, nil
	case "IfBlock":
		var test ast.Expr
		if d.Test != nil {
			test = dtoToExpr(*d.Test)
		}
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		els, err := astDTOToNodes(d.Else)
		if err != nil {
			return nil, err
		}
		return &ast.IfBlock{Unless: d.Unless, Test: test, Hash: dtoToHashEntries(d.Hash), Body: body, Else: els}, nil
	case "WithBlock":
		var val ast.Expr
		if d.WithValue != nil {
			val = dtoToExpr(*d.WithValue)
		}
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		els, err := astDTOToNodes(d.Else)
		if err != nil {
			return nil, err
		}
		return &ast.WithBlock{Value: val, BlockParams: d.BlockParams, Body: body, Else: els}, nil
	case "EachBlock":
		var col ast.Expr
		if d.Collection != nil {
			col = dtoToExpr(*d.Collection)
		}
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		els, err := astDTOToNodes(d.Else)
		if err != nil {
			return nil, err
		}
		return &ast.EachBlock{Collection: col, BlockParams: d.BlockParams, Body: body, Else: els}, nil
	case "LayoutBlock":
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		return &ast.LayoutBlock{Name: d.Name, Body: body}, nil
	case "LayoutPartial":
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		return &ast.LayoutPartial{Name: d.Name, Body: body}, nil
	case "Block":
		params := make([]ast.Expr, len(d.Params))
		for i, p := range d.Params {
			params[i] = dtoToExpr(p)
		}
		body, err := astDTOToNodes(d.Body)
		if err != nil {
			return nil, err
		}
		els, err := astDTOToNodes(d.Else)
		if err != nil {
			return nil, err
		}
		return &ast.Block{Name: d.Name, Params: params, Hash: dtoToHashEntries(d.Hash), BlockParams: d.BlockParams, Body: body, Else: els}, nil
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
	TypeSet         map[string][]string    `json:"typeSet"`
	TypeTrees       map[string]TypeNodeDTO `json:"typeTrees"`
	EmitOrder       []string               `json:"emitOrder"`
	NeedFmt         bool                   `json:"needFmt"`
	UseLayoutBlocks bool                   `json:"useLayoutBlocks"`
}

// ImportSpecJSON is the JSON shape for an import (path + optional name).
type ImportSpecJSON struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

// Phase3ResultDTO is the JSON shape for Phase3Result (IR).
type Phase3ResultDTO struct {
	PackageName       string           `json:"packageName"`
	RuntimeImport     string           `json:"runtimeImport"`
	EmitOrder         []string         `json:"emitOrder"`
	ContextIfaces     []IRContextIface `json:"contextIfaces"`
	ContextData       []IRContextData  `json:"contextData"`
	Partials          []IRPartial      `json:"partials"`
	RenderFuncs       []IRRenderFunc   `json:"renderFuncs"`
	UseLayoutBlocks   bool             `json:"useLayoutBlocks"`
	GenerateBootstrap bool             `json:"generateBootstrap"`
	HelperImports     []ImportSpecJSON `json:"helperImports"`
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
