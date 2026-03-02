package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
	"github.com/andriyg76/go-hbars/internal/parser"
	"github.com/andriyg76/hexerr"
)

// phase1ToMaps converts Phase1Result to string-keyed maps for existing context_infer and compile helpers.
func phase1ToMaps(p1 *Phase1Result) (parsed map[string][]ast.Node, names []string, funcNames map[string]string, helperExprs map[string]string) {
	parsed = make(map[string][]ast.Node, len(p1.Templates))
	for t, nodes := range p1.Templates {
		parsed[string(t)] = nodes
	}
	names = make([]string, 0, len(p1.Names))
	for _, t := range p1.Names {
		names = append(names, string(t))
	}
	funcNames = make(map[string]string, len(p1.FuncNames))
	for t, id := range p1.FuncNames {
		funcNames[string(t)] = string(id)
	}
	helperExprs = make(map[string]string, len(p1.HelperExprs))
	for h, expr := range p1.HelperExprs {
		helperExprs[string(h)] = expr
	}
	return parsed, names, funcNames, helperExprs
}

// RunPhase1 parses templates and returns Phase1Result.
func RunPhase1(templates map[string]string, opts Options) (*Phase1Result, error) {
	if opts.Log != nil {
		opts.Log.Trace("phase 1: parsing %d templates", len(templates))
	}
	if opts.PackageName == "" {
		return nil, hexerr.New("compiler: package name is required")
	}
	runtimeImport := opts.RuntimeImport
	if runtimeImport == "" {
		runtimeImport = "github.com/andriyg76/go-hbars/runtime"
	}
	helperExprs, helperImports, err := prepareHelpers(opts.Helpers, runtimeImport)
	if err != nil {
		return nil, err
	}

	names := make([]Template, 0, len(templates))
	parsed := make(map[Template][]ast.Node, len(templates))
	for name, tmpl := range templates {
		nodes, err := parser.Parse(tmpl)
		if err != nil {
			if file := opts.TemplateFiles[name]; file != "" {
				return nil, templateErrorWithFile(file, err)
			}
			return nil, hexerr.Wrapf(err, "compiler: template %q", name)
		}
		t := Template(name)
		parsed[t] = nodes
		names = append(names, t)
	}
	sort.Slice(names, func(i, j int) bool { return string(names[i]) < string(names[j]) })

	funcNames := make(map[Template]GoIdent, len(names))
	seenFunc := make(map[GoIdent]Template, len(names))
	for _, name := range names {
		ident := GoIdent(goIdent(string(name)))
		if prev, exists := seenFunc[ident]; exists {
			return nil, hexerr.New(fmt.Sprintf("compiler: templates %q and %q map to %q", prev, name, ident))
		}
		seenFunc[ident] = name
		funcNames[name] = ident
	}

	parsedStr, _, _, _ := phase1ToMaps(&Phase1Result{Templates: parsed, Names: names, FuncNames: funcNames, HelperExprs: make(map[Helper]string)})
	usedHelpersMap := collectUsedHelperNames(parsedStr, helperExprs)
	usedHelpers := make([]Helper, 0, len(usedHelpersMap))
	for h := range usedHelpersMap {
		usedHelpers = append(usedHelpers, Helper(h))
	}
	sort.Slice(usedHelpers, func(i, j int) bool { return string(usedHelpers[i]) < string(usedHelpers[j]) })
	helperImports = filterHelperImports(helperImports, opts.Helpers, usedHelpersMap)

	templateFiles := make(map[Template]string, len(opts.TemplateFiles))
	for k, v := range opts.TemplateFiles {
		templateFiles[Template(k)] = v
	}
	helperExprsTyped := make(map[Helper]string, len(helperExprs))
	for k, v := range helperExprs {
		helperExprsTyped[Helper(k)] = v
	}
	if opts.Log != nil {
		opts.Log.Debug("phase 1 done: %d templates, names: %v", len(names), names)
	}

	return &Phase1Result{
		Templates:     parsed,
		Names:         names,
		FuncNames:     funcNames,
		HelperExprs:   helperExprsTyped,
		HelperImports: helperImports,
		UsedHelpers:   usedHelpers,
		TemplateFiles: templateFiles,
	}, nil
}

// RunPhase2a computes partial/context type analysis from Phase1Result.
func RunPhase2a(phase1 *Phase1Result, opts Options) (*Phase2aResult, error) {
	if opts.Log != nil {
		opts.Log.Trace("phase 2a: partial/context type analysis")
	}
	parsed, names, funcNamesStr, helperExprs := phase1ToMaps(phase1)

	partialParamTypes, hashOnlyPartials := CollectPartialParamTypes(parsed, names, funcNamesStr, helperExprs)
	typeSet, _ := buildPartialTypeSet(parsed, names, funcNamesStr, helperExprs)
	for partialName, set := range typeSet {
		if len(set) > 1 && funcNamesStr[partialName] == "" {
			funcNamesStr[partialName] = goIdent(partialName)
		}
	}
	canonicalType, primaryCaller := SharedPartialInfo(typeSet, names, funcNamesStr)

	partialParamTypesTyped := make(map[PartialName]ContextTypeName, len(partialParamTypes))
	for k, v := range partialParamTypes {
		partialParamTypesTyped[PartialName(k)] = ContextTypeName(v)
	}
	typeSetTyped := make(map[PartialName]map[ContextTypeName]bool, len(typeSet))
	for k, set := range typeSet {
		m := make(map[ContextTypeName]bool, len(set))
		for ct := range set {
			m[ContextTypeName(ct)] = true
		}
		typeSetTyped[PartialName(k)] = m
	}
	canonicalTypeTyped := make(map[PartialName]ContextTypeName, len(canonicalType))
	for k, v := range canonicalType {
		canonicalTypeTyped[PartialName(k)] = ContextTypeName(v)
	}
	primaryCallerTyped := make(map[PartialName]Template, len(primaryCaller))
	for k, v := range primaryCaller {
		primaryCallerTyped[PartialName(k)] = Template(v)
	}
	funcNamesTyped := make(map[PartialName]GoIdent, len(funcNamesStr))
	for k, v := range funcNamesStr {
		funcNamesTyped[PartialName(k)] = GoIdent(v)
	}
	if opts.Log != nil {
		opts.Log.Debug("phase 2a done: canonical types: %d partials", len(canonicalTypeTyped))
	}

	return &Phase2aResult{
		PartialParamTypes: partialParamTypesTyped,
		HashOnlyPartials:  hashOnlyPartials,
		TypeSet:           typeSetTyped,
		CanonicalType:     canonicalTypeTyped,
		PrimaryCaller:     primaryCallerTyped,
		FuncNames:         funcNamesTyped,
	}, nil
}

// RunPhase2b builds type trees for each template.
func RunPhase2b(phase1 *Phase1Result, phase2a *Phase2aResult, opts Options) (*Phase2bResult, error) {
	if opts.Log != nil {
		opts.Log.Trace("phase 2b: building type trees")
	}
	parsed, _, funcNamesStr, helperExprs := phase1ToMaps(phase1)

	typeTrees := make(map[Template]*typeNode)
	for _, name := range phase1.Names {
		if opts.Log != nil {
			opts.Log.Trace("phase 2b: type tree for template %s", name)
		}
		col := newPathCollector(helperExprs)
		col.setParsed(parsed)
		nodes := parsed[string(name)]
		if err := col.collectNodes(nodes); err != nil {
			if file := phase1.TemplateFiles[name]; file != "" {
				return nil, templateErrorWithFile(file, err)
			}
			return nil, hexerr.Wrapf(err, "compiler: template %q context inference", name)
		}
		typeTrees[name] = buildTypeTree(col.paths, col.eachFields)
	}

	contextTypeToTemplate := make(map[ContextTypeName]Template)
	for _, name := range phase1.Names {
		goName := funcNamesStr[string(name)]
		ownCtx := goName + "Context"
		paramType := phase2a.PartialParamTypes[PartialName(name)]
		if paramType != "" && string(paramType) != ownCtx {
			continue
		}
		tree := typeTrees[name]
		for _, ctxType := range AllContextTypeNames(goName, tree) {
			contextTypeToTemplate[ContextTypeName(ctxType)] = name
		}
	}

	var emitOrder []Template
	for _, name := range phase1.Names {
		goName := funcNamesStr[string(name)]
		ownCtx := goName + "Context"
		paramType := phase2a.PartialParamTypes[PartialName(name)]
		if paramType != "" && string(paramType) != ownCtx {
			continue
		}
		emitOrder = append(emitOrder, name)
	}
	var first, rest []Template
	for _, name := range emitOrder {
		if phase2a.CanonicalType[PartialName(name)] != "" {
			first = append(first, name)
		} else {
			rest = append(rest, name)
		}
	}
	emitOrder = append(first, rest...)

	needFmt := templatesUseBlockHelpers(parsed, helperExprs) || opts.GenerateBootstrap
	useLayoutBlocks := templatesUsesLayoutBlocks(parsed)
	if opts.Log != nil {
		opts.Log.Debug("phase 2b done: %d type trees, emit order: %v", len(typeTrees), emitOrder)
	}

	return &Phase2bResult{
		Phase2a:         phase2a,
		TypeTrees:       typeTrees,
		ContextToTmpl:   contextTypeToTemplate,
		EmitOrder:       emitOrder,
		NeedFmt:         needFmt,
		UseLayoutBlocks: useLayoutBlocks,
	}, nil
}

// buildIRContextInfo builds IRContextIface and IRContextData lists from type trees and canonical type info.
// These populate the Phase3 IR so the JSON output describes the full context structure for debugging.
func buildIRContextInfo(phase1 *Phase1Result, phase2b *Phase2bResult) ([]IRContextIface, []IRContextData) {
	funcNamesStr := make(map[string]string, len(phase2b.Phase2a.FuncNames))
	for k, v := range phase2b.Phase2a.FuncNames {
		funcNamesStr[string(k)] = string(v)
	}
	canonicalType := make(map[string]string, len(phase2b.Phase2a.CanonicalType))
	for k, v := range phase2b.Phase2a.CanonicalType {
		canonicalType[string(k)] = string(v)
	}
	typeTrees := make(map[string]*typeNode, len(phase2b.TypeTrees))
	for t, tree := range phase2b.TypeTrees {
		typeTrees[string(t)] = tree
	}

	var ifaces []IRContextIface
	var datas []IRContextData
	seenGoName := make(map[string]bool)

	// Canonical-only partials (shared partials not in emitOrder) are emitted first.
	emitOrderSet := make(map[string]bool, len(phase2b.EmitOrder))
	for _, name := range phase2b.EmitOrder {
		emitOrderSet[string(name)] = true
	}
	var canonicalOnly []string
	for partialName := range canonicalType {
		if !emitOrderSet[partialName] {
			canonicalOnly = append(canonicalOnly, partialName)
		}
	}
	sort.Strings(canonicalOnly)
	for _, partialName := range canonicalOnly {
		canon := canonicalType[partialName]
		if canon == "" {
			continue
		}
		goName := strings.TrimSuffix(canon, "Context")
		var subtree *typeNode
		for _, name := range phase2b.EmitOrder {
			tree := typeTrees[string(name)]
			if tree != nil && tree.fields != nil && tree.fields[partialName] != nil {
				subtree = tree.fields[partialName]
				break
			}
		}
		if subtree == nil {
			continue
		}
		collectContextIR(partialName, goName, subtree, canonicalType, typeTrees, seenGoName, &ifaces, &datas)
	}

	// Templates in emit order.
	for _, name := range phase2b.EmitOrder {
		nameStr := string(name)
		goName := funcNamesStr[nameStr]
		if goName == "" {
			continue
		}
		tree := typeTrees[nameStr]
		collectContextIR(nameStr, goName, tree, canonicalType, typeTrees, seenGoName, &ifaces, &datas)
	}

	return ifaces, datas
}

// collectContextIR builds IRContextIface/IRContextData entries for one template and its nested types.
func collectContextIR(
	templateName, goName string,
	tree *typeNode,
	canonicalType map[string]string,
	typeTrees map[string]*typeNode,
	seenGoName map[string]bool,
	ifaces *[]IRContextIface,
	datas *[]IRContextData,
) {
	if seenGoName[goName] {
		return
	}
	seenGoName[goName] = true

	layoutPartial := rootLayoutPartial(tree, canonicalType)
	if layoutPartial == templateName {
		layoutPartial = ""
	}
	var skipFieldsAtRoot map[string]bool
	if layoutPartial != "" {
		if layoutTree := typeTrees[layoutPartial]; layoutTree != nil && layoutTree.fields != nil {
			skipFieldsAtRoot = make(map[string]bool)
			skipFieldsAtRoot[layoutPartial] = true
			for f := range layoutTree.fields {
				if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
					continue
				}
				skipFieldsAtRoot[f] = true
			}
		}
	}

	// Root context interface.
	rootIface := IRContextIface{
		Template: Template(templateName),
		GoName:   GoIdent(goName),
		Comment:  fmt.Sprintf("context interface inferred from template %q", templateName),
	}
	if layoutPartial != "" {
		rootIface.Methods = append(rootIface.Methods, IRIfaceMethod{
			FieldPath:   "",
			ReturnType:  ContextTypeName(canonicalType[layoutPartial]),
			IsCanonical: true,
		})
	}
	ownFields := irFieldsFromTree(goName, "", tree, canonicalType, skipFieldsAtRoot)
	rootIface.Methods = append(rootIface.Methods, ownFields...)
	rootIface.Methods = append(rootIface.Methods, IRIfaceMethod{FieldPath: "Raw", ReturnType: "any"})
	*ifaces = append(*ifaces, rootIface)

	// Root context data struct: includes fields from embedded layout + own fields.
	rootData := IRContextData{
		Template: Template(templateName),
		GoName:   GoIdent(goName),
		Comment:  fmt.Sprintf("map-backed implementation of %sContext", goName),
	}
	if layoutPartial != "" {
		layoutGoName := strings.TrimSuffix(canonicalType[layoutPartial], "Context")
		layoutTree := typeTrees[layoutPartial]
		rootData.Fields = append(rootData.Fields, irDataFieldsFromTree(layoutGoName, layoutTree, canonicalType, nil)...)
	}
	rootData.Fields = append(rootData.Fields, irDataFieldsFromTree(goName, tree, canonicalType, skipFieldsAtRoot)...)
	*datas = append(*datas, rootData)

	// Nested context types from tree fields (non-canonical, non-skipped paths with sub-fields).
	if tree != nil && tree.fields != nil {
		var fieldNames []string
		for f := range tree.fields {
			if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
				continue
			}
			fieldNames = append(fieldNames, f)
		}
		sort.Strings(fieldNames)
		for _, field := range fieldNames {
			if canonicalType[field] != "" {
				continue // canonical partial: its own collectContextIR handles it
			}
			child := tree.fields[field]
			if child.isSlice && child.sliceElem != nil {
				collectNestedContextIR(templateName, goName, field, child, canonicalType, seenGoName, ifaces, datas)
			} else if len(child.fields) > 0 {
				collectNestedContextIR(templateName, goName, field, child, canonicalType, seenGoName, ifaces, datas)
			}
		}
	}
}

// collectNestedContextIR emits IRContextIface/Data for nested paths in the type tree.
func collectNestedContextIR(
	templateName, goIdent, pathPrefix string,
	n *typeNode,
	canonicalType map[string]string,
	seenGoName map[string]bool,
	ifaces *[]IRContextIface,
	datas *[]IRContextData,
) {
	if n == nil {
		return
	}
	if n.isSlice && n.sliceElem != nil {
		ifaceName := contextItemInterfaceName(goIdent, pathPrefix)
		ifaceGoName := strings.TrimSuffix(ifaceName, "Context")
		if seenGoName[ifaceGoName] {
			return
		}
		seenGoName[ifaceGoName] = true
		iface := IRContextIface{
			Template: Template(templateName),
			GoName:   GoIdent(ifaceGoName),
			Comment:  fmt.Sprintf("context for one element of %q", pathPrefix),
		}
		iface.Methods = append(iface.Methods, irFieldsFromTree(goIdent, pathPrefix+".", n.sliceElem, canonicalType, nil)...)
		iface.Methods = append(iface.Methods, IRIfaceMethod{FieldPath: "Raw", ReturnType: "any"})
		*ifaces = append(*ifaces, iface)
		data := IRContextData{
			Template: Template(templateName),
			GoName:   GoIdent(ifaceGoName),
			Comment:  fmt.Sprintf("map-backed implementation of %s", ifaceName),
			Fields:   irDataFieldsFromTree(goIdent, n.sliceElem, canonicalType, nil),
		}
		*datas = append(*datas, data)
		collectNestedContextIR(templateName, goIdent, pathPrefix+".", n.sliceElem, canonicalType, seenGoName, ifaces, datas)
		return
	}
	if n.fields == nil {
		return
	}
	ifaceName := contextInterfaceName(goIdent, pathPrefix)
	ifaceGoName := strings.TrimSuffix(ifaceName, "Context")
	if seenGoName[ifaceGoName] {
		return
	}
	seenGoName[ifaceGoName] = true
	iface := IRContextIface{
		Template: Template(templateName),
		GoName:   GoIdent(ifaceGoName),
		Comment:  fmt.Sprintf("context for path %q", pathPrefix),
	}
	iface.Methods = append(iface.Methods, irFieldsFromTree(goIdent, pathPrefix+".", n, canonicalType, nil)...)
	iface.Methods = append(iface.Methods, IRIfaceMethod{FieldPath: "Raw", ReturnType: "any"})
	*ifaces = append(*ifaces, iface)
	data := IRContextData{
		Template: Template(templateName),
		GoName:   GoIdent(ifaceGoName),
		Comment:  fmt.Sprintf("map-backed implementation of %s", ifaceName),
		Fields:   irDataFieldsFromTree(goIdent, n, canonicalType, nil),
	}
	*datas = append(*datas, data)
	// Recurse into sub-fields.
	var fieldNames []string
	for f := range n.fields {
		if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		fieldNames = append(fieldNames, f)
	}
	sort.Strings(fieldNames)
	for _, field := range fieldNames {
		if canonicalType[field] != "" {
			continue
		}
		child := n.fields[field]
		subPath := pathPrefix + "." + field
		if pathPrefix == "" {
			subPath = field
		}
		if (child.isSlice && child.sliceElem != nil) || len(child.fields) > 0 {
			collectNestedContextIR(templateName, goIdent, subPath, child, canonicalType, seenGoName, ifaces, datas)
		}
	}
}

// irFieldsFromTree builds IRIfaceMethod list for a type tree node (one level of interface methods).
// skipFields: set of fields to skip (used to skip layout-covered fields at root).
func irFieldsFromTree(goName, pathPrefix string, n *typeNode, canonicalType map[string]string, skipFields map[string]bool) []IRIfaceMethod {
	if n == nil || n.fields == nil {
		return nil
	}
	var fieldNames []string
	for f := range n.fields {
		if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		if skipFields != nil && skipFields[f] {
			continue
		}
		fieldNames = append(fieldNames, f)
	}
	sort.Strings(fieldNames)
	var out []IRIfaceMethod
	for _, field := range fieldNames {
		child := n.fields[field]
		var retType string
		if child.isSlice && child.sliceElem != nil {
			retType = "[]" + contextItemInterfaceName(goName, pathPrefix+field)
		} else if len(child.fields) > 0 {
			if canon := canonicalType[field]; canon != "" {
				retType = canon
			} else {
				retType = contextInterfaceName(goName, pathPrefix+field)
			}
		} else {
			retType = "any"
		}
		out = append(out, IRIfaceMethod{FieldPath: field, ReturnType: ContextTypeName(retType)})
	}
	return out
}

// irDataFieldsFromTree builds IRDataField list from a type tree (fields needed for a ContextData struct).
// skipFields: set of root-level fields to skip.
func irDataFieldsFromTree(goName string, n *typeNode, canonicalType map[string]string, skipFields map[string]bool) []IRDataField {
	if n == nil || n.fields == nil {
		return nil
	}
	var fieldNames []string
	for f := range n.fields {
		if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		if skipFields != nil && skipFields[f] {
			continue
		}
		fieldNames = append(fieldNames, f)
	}
	sort.Strings(fieldNames)
	var out []IRDataField
	for _, field := range fieldNames {
		child := n.fields[field]
		var fieldType string
		if child.isSlice && child.sliceElem != nil {
			fieldType = "[]" + contextItemInterfaceName(goName, field)
		} else if len(child.fields) > 0 {
			if canon := canonicalType[field]; canon != "" {
				fieldType = canon
			} else {
				fieldType = contextInterfaceName(goName, field)
			}
		} else {
			fieldType = "any"
		}
		out = append(out, IRDataField{FieldPath: field, Type: ContextTypeName(fieldType)})
	}
	return out
}

// RunPhase3 builds the structured IR from Phase2bResult and Phase1Result.
func RunPhase3(phase1 *Phase1Result, phase2b *Phase2bResult, opts Options) (*Phase3Result, error) {
	if opts.Log != nil {
		opts.Log.Trace("phase 3: building IR")
	}
	runtimeImport := opts.RuntimeImport
	if runtimeImport == "" {
		runtimeImport = "github.com/andriyg76/go-hbars/runtime"
	}

	contextIfaces, contextData := buildIRContextInfo(phase1, phase2b)

	ir := &Phase3Result{
		PackageName:       opts.PackageName,
		RuntimeImport:     runtimeImport,
		EmitOrder:         phase2b.EmitOrder,
		UseLayoutBlocks:   phase2b.UseLayoutBlocks,
		GenerateBootstrap: opts.GenerateBootstrap,
		HelperImports:     phase1.HelperImports,
		ContextIfaces:     contextIfaces,
		ContextData:       contextData,
		Partials:          nil,
		RenderFuncs:       nil,
	}

	// Partials: one entry per root template (name, go name, context type, FromMap name).
	ir.Partials = make([]IRPartial, 0, len(phase1.Names))
	for _, name := range phase1.Names {
		goName := phase2b.Phase2a.FuncNames[PartialName(name)]
		if goName == "" {
			continue
		}
		ctxType := phase2b.Phase2a.PartialParamTypes[PartialName(name)]
		if ctxType == "" {
			ctxType = ContextTypeName(string(goName) + "Context")
		}
		fromMap := strings.TrimSuffix(string(ctxType), "Context") + "ContextFromMap"
		ir.Partials = append(ir.Partials, IRPartial{
			Name:      name,
			GoName:    goName,
			CtxType:   ctxType,
			FromMapFn: fromMap,
		})
	}

	// RenderFuncs: one per root template (renderX, RenderX, RenderXString; optional WithBlocks).
	ir.RenderFuncs = make([]IRRenderFunc, 0, len(phase1.Names))
	for _, name := range phase1.Names {
		goName := phase2b.Phase2a.FuncNames[PartialName(name)]
		if goName == "" {
			continue
		}
		rootCtx := phase2b.Phase2a.PartialParamTypes[PartialName(name)]
		if rootCtx == "" {
			rootCtx = ContextTypeName(string(goName) + "Context")
		}
		ir.RenderFuncs = append(ir.RenderFuncs, IRRenderFunc{
			Template:        name,
			GoName:          goName,
			RootContext:     rootCtx,
			UseLayoutBlocks: phase2b.UseLayoutBlocks,
		})
	}

	if opts.Log != nil {
		opts.Log.Debug("phase 3 done: emit order %d, partials %d, render funcs %d", len(ir.EmitOrder), len(ir.Partials), len(ir.RenderFuncs))
	}
	return ir, nil
}

// RunPhase4 emits Go source from phase results (uses phase1 and phase2b to drive existing codegen).
func RunPhase4(phase1 *Phase1Result, phase2b *Phase2bResult, opts Options) ([]byte, error) {
	if opts.Log != nil {
		opts.Log.Trace("phase 4: emitting Go")
	}
	parsed, names, funcNamesStr, helperExprs := phase1ToMaps(phase1)
	partialParamTypes := make(map[string]string)
	for k, v := range phase2b.Phase2a.PartialParamTypes {
		partialParamTypes[string(k)] = string(v)
	}
	canonicalType := make(map[string]string)
	for k, v := range phase2b.Phase2a.CanonicalType {
		canonicalType[string(k)] = string(v)
	}
	primaryCaller := make(map[string]string)
	for k, v := range phase2b.Phase2a.PrimaryCaller {
		primaryCaller[string(k)] = string(v)
	}
	typeTrees := make(map[string]*typeNode)
	for t, tree := range phase2b.TypeTrees {
		typeTrees[string(t)] = tree
	}
	contextTypeToTemplate := make(map[string]string)
	for k, v := range phase2b.ContextToTmpl {
		contextTypeToTemplate[string(k)] = string(v)
	}
	runtimeImport := opts.RuntimeImport
	if runtimeImport == "" {
		runtimeImport = "github.com/andriyg76/go-hbars/runtime"
	}
	hashOnlyPartials := phase2b.Phase2a.HashOnlyPartials
	if hashOnlyPartials == nil {
		hashOnlyPartials = make(map[string]bool)
	}
	return compileCodegen(codegenParams{
		parsed:                 parsed,
		names:                  names,
		funcNames:              funcNamesStr,
		partialParamTypes:      partialParamTypes,
		hashOnlyPartials:       hashOnlyPartials,
		canonicalType:          canonicalType,
		primaryCaller:          primaryCaller,
		typeTrees:              typeTrees,
		contextTypeToTemplate:  contextTypeToTemplate,
		needFmt:                phase2b.NeedFmt,
		useLayoutBlocks:        phase2b.UseLayoutBlocks,
		helperExprs:            helperExprs,
		helperImports:          phase1.HelperImports,
		runtimeImport:          runtimeImport,
	}, opts)
}

// writePhaseOutput writes the phase result to path as JSON, or no-op if path is empty.
func writePhaseOutput(path string, v any) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return hexerr.Wrapf(err, "marshal phase output")
	}
	return hexerr.Wrapf(os.WriteFile(path, data, 0644), "write %s", path)
}

// RunPhased runs phases 1 through opts.MaxPhase, writes requested phase outputs, and returns
// the generated Go source when MaxPhase is Phase4. If MaxPhase is 0, RunPhased is not used.
func RunPhased(templates map[string]string, opts Options) (goSource []byte, err error) {
	max := opts.MaxPhase
	if max <= 0 {
		max = Phase4
	}
	var p1 *Phase1Result
	var p2a *Phase2aResult
	var p2b *Phase2bResult
	var p3 *Phase3Result

	if p1, err = RunPhase1(templates, opts); err != nil {
		return nil, err
	}
	if opts.Phase1Output != "" {
		if err = writePhaseOutput(opts.Phase1Output, p1); err != nil {
			return nil, err
		}
	}
	if max <= Phase1 {
		return nil, nil
	}

	if p2a, err = RunPhase2a(p1, opts); err != nil {
		return nil, err
	}
	if opts.Phase2aOutput != "" {
		if err = writePhaseOutput(opts.Phase2aOutput, p2a); err != nil {
			return nil, err
		}
	}
	if max <= Phase2a {
		return nil, nil
	}

	if p2b, err = RunPhase2b(p1, p2a, opts); err != nil {
		return nil, err
	}
	if opts.Phase2bOutput != "" {
		if err = writePhaseOutput(opts.Phase2bOutput, p2b); err != nil {
			return nil, err
		}
	}
	if max <= Phase2b {
		return nil, nil
	}

	if p3, err = RunPhase3(p1, p2b, opts); err != nil {
		return nil, err
	}
	if opts.Phase3Output != "" {
		if err = writePhaseOutput(opts.Phase3Output, p3); err != nil {
			return nil, err
		}
	}
	if max <= Phase3 {
		return nil, nil
	}

	goSource, err = RunPhase4(p1, p2b, opts)
	if err != nil {
		return nil, err
	}
	if opts.Phase4Output != "" && len(goSource) > 0 {
		if err = os.WriteFile(opts.Phase4Output, goSource, 0644); err != nil {
			return nil, hexerr.Wrapf(err, "write %s", opts.Phase4Output)
		}
	}
	return goSource, nil
}
