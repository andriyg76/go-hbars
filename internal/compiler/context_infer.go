package compiler

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
)

// #region agent log
func debugLog(hy string, loc string, msg string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["hypothesisId"] = hy
	data["location"] = loc
	data["message"] = msg
	f, err := os.OpenFile("/home/andrij/src/go-hbars/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

// #endregion

// pathScope represents the current scope when walking the template AST.
type pathScope struct {
	dataPath       string            // resolved path of current context, e.g. "user"
	params         map[string]string // block param name -> resolved path, e.g. "u" -> "user"
	eachCollection string            // when inside {{#each col}}, this is "col"
	eachParam      string            // when inside each, the first block param, e.g. "person"
}

type pathCollector struct {
	helpers    map[string]bool
	paths      map[string]bool
	eachFields map[string]map[string]bool // collection path -> set of element field names
	scopeStack []pathScope
	parsed     map[string][]ast.Node // template name -> AST; when set, collectPartial merges partial paths
}

func newPathCollector(helperNames map[string]string) *pathCollector {
	helpers := make(map[string]bool)
	for name := range helperNames {
		helpers[name] = true
	}
	return &pathCollector{
		helpers:    helpers,
		paths:      make(map[string]bool),
		eachFields: make(map[string]map[string]bool),
		scopeStack: []pathScope{{}},
	}
}

// setParsed sets the parsed templates so that collectPartial can merge paths from included partials.
func (c *pathCollector) setParsed(parsed map[string][]ast.Node) {
	c.parsed = parsed
}

func (c *pathCollector) pushWith(dataPath string, params []string) {
	paramMap := make(map[string]string)
	if len(params) > 0 {
		paramMap[params[0]] = dataPath
	}
	c.scopeStack = append(c.scopeStack, pathScope{
		dataPath: dataPath,
		params:   paramMap,
	})
}

func (c *pathCollector) pushEach(collectionPath string, params []string) {
	frame := pathScope{
		dataPath:       c.scopeStack[len(c.scopeStack)-1].dataPath,
		eachCollection: collectionPath,
	}
	if len(params) > 0 {
		frame.eachParam = params[0]
	}
	c.scopeStack = append(c.scopeStack, frame)
}

func (c *pathCollector) pop() {
	if len(c.scopeStack) > 1 {
		c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]
	}
}

// currentScopeType returns the context interface name for the current scope (e.g. MainContext, MainUserContext).
func (c *pathCollector) currentScopeType(goName string) string {
	if len(c.scopeStack) == 0 {
		return goName + "Context"
	}
	top := c.scopeStack[len(c.scopeStack)-1]
	if top.eachCollection != "" {
		return contextItemInterfaceName(goName, top.eachCollection)
	}
	if top.dataPath != "" {
		return contextInterfaceName(goName, top.dataPath)
	}
	return goName + "Context"
}

// pathToScopeType returns the context interface name for a path (e.g. "user" -> MainUserContext).
func (c *pathCollector) pathToScopeType(goName, pathStr string) string {
	if pathStr == "" {
		return c.currentScopeType(goName)
	}
	full, elem := c.resolvePath(pathStr)
	if elem != "" {
		return contextItemInterfaceName(goName, full)
	}
	if full != "" {
		return contextInterfaceName(goName, full)
	}
	return goName + "Context"
}

// resolvePath returns the full data path and whether this is an element field inside {{#each}}.
func (c *pathCollector) resolvePath(pathStr string) (fullPath string, elementField string) {
	return c.resolvePathAt(pathStr, len(c.scopeStack)-1)
}

func (c *pathCollector) resolvePathAt(pathStr string, scopeIdx int) (fullPath string, elementField string) {
	pathStr = strings.TrimSpace(pathStr)
	if pathStr == "" || scopeIdx < 0 {
		return "", ""
	}
	top := c.scopeStack[scopeIdx]
	parts := strings.Split(pathStr, ".")
	if parts[0] == "@root" {
		if len(parts) == 1 {
			return "", ""
		}
		return strings.Join(parts[1:], "."), ""
	}
	if parts[0] == ".." || strings.HasPrefix(pathStr, "../") {
		if scopeIdx == 0 {
			if len(parts) == 1 {
				return "", ""
			}
			return strings.Join(parts[1:], "."), ""
		}
		parent := c.scopeStack[scopeIdx-1]
		if len(parts) == 1 {
			return parent.dataPath, ""
		}
		rest := strings.Join(parts[1:], ".")
		p, ef := c.resolvePathAt(rest, scopeIdx-1)
		if p != "" {
			if parent.dataPath != "" {
				return parent.dataPath + "." + p, ef
			}
			return p, ef
		}
		if parent.dataPath != "" {
			return parent.dataPath + "." + rest, ""
		}
		return rest, ""
	}
	if top.eachCollection != "" && top.eachParam != "" && parts[0] == top.eachParam {
		if len(parts) == 1 {
			return top.eachCollection, ""
		}
		return top.eachCollection, strings.Join(parts[1:], ".")
	}
	if top.params != nil {
		if base, ok := top.params[parts[0]]; ok {
			if len(parts) == 1 {
				return base, ""
			}
			return base + "." + strings.Join(parts[1:], "."), ""
		}
	}
	if top.dataPath != "" {
		return top.dataPath + "." + pathStr, ""
	}
	return pathStr, ""
}

func (c *pathCollector) addPath(fullPath string, elementField string) {
	if fullPath == "" && elementField == "" {
		return
	}
	if fullPath != "" && (fullPath[0] == '@' || fullPath[0] == '.' || fullPath == ".") {
		return
	}
	if elementField != "" && (elementField[0] == '@' || elementField[0] == '.' || elementField == ".") {
		return
	}
	if elementField != "" && fullPath != "" {
		if c.eachFields[fullPath] == nil {
			c.eachFields[fullPath] = make(map[string]bool)
		}
		c.eachFields[fullPath][elementField] = true
		return
	}
	if fullPath != "" {
		c.paths[fullPath] = true
	}
}

// pathsFromExpr returns all path-like strings used in an expression (for data lookup).
func pathsFromExpr(e expr) []string {
	var out []string
	switch e.kind {
	case exprPath:
		return []string{e.value}
	case exprCall:
		for _, a := range e.args {
			out = append(out, pathsFromExpr(a)...)
		}
		for _, h := range e.hash {
			out = append(out, pathsFromExpr(h.value)...)
		}
		return out
	default:
		return nil
	}
}

func (c *pathCollector) collectNodes(nodes []ast.Node) error {
	for _, node := range nodes {
		if err := c.collectNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (c *pathCollector) collectNode(node ast.Node) error {
	switch n := node.(type) {
	case *ast.Text:
		return nil
	case *ast.Mustache:
		return c.collectMustache(n)
	case *ast.Partial:
		return c.collectPartial(n)
	case *ast.Block:
		return c.collectBlock(n)
	default:
		return nil
	}
}

func (c *pathCollector) collectMustache(n *ast.Mustache) error {
	parts, _, err := parseParts(n.Expr)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		if parts[0].kind == exprPath && !c.helpers[parts[0].value] {
			pathStr := parts[0].value
			// Don't add @root paths to type tree so partial context interfaces don't require root-only methods
			if strings.HasPrefix(pathStr, "@root") {
				return nil
			}
			full, elem := c.resolvePath(pathStr)
			c.addPath(full, elem)
		}
		return nil
	}
	if parts[0].kind != exprPath || !c.helpers[parts[0].value] {
		return nil
	}
	for _, p := range parts[1:] {
		for _, pathStr := range pathsFromExpr(p) {
			full, elem := c.resolvePath(pathStr)
			c.addPath(full, elem)
		}
	}
	return nil
}

func (c *pathCollector) collectPartial(n *ast.Partial) error {
	parts, hash, err := parseParts(n.Expr)
	if err != nil {
		return nil
	}
	// Hash keys become partial context fields (e.g. {{> footer note="thanks"}} -> context has "note").
	for _, h := range hash {
		if h.key != "" {
			c.addPath(h.key, "")
		}
	}
	// Merge paths from the partial template so the including template's context has the required methods.
	if c.parsed != nil && len(parts) >= 1 {
		var partialName string
		switch parts[0].kind {
		case exprString:
			partialName = parts[0].value
		case exprPath:
			partialName = parts[0].value
		default:
			partialName = ""
		}
		if partialName != "" {
			sameScope := len(parts) == 1
			if sameScope {
				// Same-scope partial ({{> partial}}): add partial name as path so caller's type tree
				// has a node for it, and caller will EMBED the partial's context type.
				c.addPath(partialName, "")
				if partialNodes, ok := c.parsed[partialName]; ok {
					c.pushWith(partialName, nil)
					err := c.collectNodes(partialNodes)
					c.pop()
					if err != nil {
						return err
					}
				}
			}
			// Different-context partial ({{> partial ctx}}): do NOT merge partial's paths.
			// The partial has its own context interface; caller just needs the context path to exist.
		}
	}
	// For different-context partials, add the context path to the type tree.
	if len(parts) >= 2 {
		for _, pathStr := range pathsFromExpr(parts[1]) {
			full, elem := c.resolvePath(pathStr)
			c.addPath(full, elem)
		}
	}
	return nil
}

func (c *pathCollector) collectBlock(n *ast.Block) error {
	parts, _, err := parseParts(n.Args)
	if err != nil {
		return nil
	}
	switch n.Name {
	case "if", "unless":
		if len(parts) == 1 && parts[0].kind == exprPath {
			full, elem := c.resolvePath(parts[0].value)
			c.addPath(full, elem)
		}
		if err := c.collectNodes(n.Body); err != nil {
			return err
		}
		if err := c.collectNodes(n.Else); err != nil {
			return err
		}
		return nil
	case "with":
		var dataPath string
		if len(parts) == 1 && parts[0].kind == exprPath {
			full, _ := c.resolvePath(parts[0].value)
			dataPath = full
			c.addPath(full, "")
		}
		c.pushWith(dataPath, n.Params)
		err := c.collectNodes(n.Body)
		c.pop()
		if err != nil {
			return err
		}
		if err := c.collectNodes(n.Else); err != nil {
			return err
		}
		return nil
	case "each":
		var collectionPath string
		if len(parts) == 2 && parts[0].kind == exprPath && parts[0].value == "in" {
			if parts[1].kind == exprPath {
				full, _ := c.resolvePath(parts[1].value)
				collectionPath = full
				c.addPath(full, "")
			}
		} else if len(parts) == 1 && parts[0].kind == exprPath {
			full, _ := c.resolvePath(parts[0].value)
			collectionPath = full
			c.addPath(full, "")
		}
		c.pushEach(collectionPath, n.Params)
		err := c.collectNodes(n.Body)
		c.pop()
		if err != nil {
			return err
		}
		if err := c.collectNodes(n.Else); err != nil {
			return err
		}
		return nil
	default:
		// Universal section (e.g. {{#date}}): add the section path so context has the getter
		if len(parts) == 1 && parts[0].kind == exprPath {
			full, _ := c.resolvePath(parts[0].value)
			c.addPath(full, "")
		}
		if c.helpers[n.Name] {
			for _, p := range parts {
				for _, pathStr := range pathsFromExpr(p) {
					full, elem := c.resolvePath(pathStr)
					c.addPath(full, elem)
				}
			}
		}
		if err := c.collectNodes(n.Body); err != nil {
			return err
		}
		if err := c.collectNodes(n.Else); err != nil {
			return err
		}
		return nil
	}
}

// walkPartialsCollect walks the AST and calls add(partialName, paramType, sameScope) for each static partial call.
// sameScope: true when {{> name}} (no expr), false when {{> name expr}} — partial called with explicit context keeps its own interface.
func (c *pathCollector) walkPartialsCollect(nodes []ast.Node, goName string, add func(partialName, paramType string, sameScope bool)) error {
	for _, node := range nodes {
		if err := c.walkPartialsCollectNode(node, goName, add); err != nil {
			return err
		}
	}
	return nil
}

func (c *pathCollector) walkPartialsCollectNode(node ast.Node, goName string, add func(partialName, paramType string, sameScope bool)) error {
	switch n := node.(type) {
	case *ast.Partial:
		parts, _, err := parseParts(n.Expr)
		if err != nil || len(parts) == 0 {
			return nil
		}
		// Static partial name: exprString or exprPath
		var partialName string
		switch parts[0].kind {
		case exprString:
			partialName = parts[0].value
		case exprPath:
			partialName = parts[0].value
		default:
			return nil // dynamic partial name, skip
		}
		sameScope := len(parts) == 1
		var paramType string
		if sameScope {
			paramType = c.currentScopeType(goName)
		} else if len(parts) == 2 && parts[1].kind == exprPath {
			paramType = c.pathToScopeType(goName, parts[1].value)
		} else {
			paramType = "any"
		}
		add(partialName, paramType, sameScope)
		return nil
	case *ast.Block:
		parts, _, err := parseParts(n.Args)
		if err != nil {
			return nil
		}
		switch n.Name {
		case "with":
			var dataPath string
			if len(parts) == 1 && parts[0].kind == exprPath {
				full, _ := c.resolvePath(parts[0].value)
				dataPath = full
			}
			c.pushWith(dataPath, n.Params)
			err := c.walkPartialsCollect(n.Body, goName, add)
			c.pop()
			if err != nil {
				return err
			}
			return c.walkPartialsCollect(n.Else, goName, add)
		case "each":
			var collectionPath string
			if len(parts) == 2 && parts[0].kind == exprPath && parts[0].value == "in" && parts[1].kind == exprPath {
				full, _ := c.resolvePath(parts[1].value)
				collectionPath = full
			} else if len(parts) == 1 && parts[0].kind == exprPath {
				full, _ := c.resolvePath(parts[0].value)
				collectionPath = full
			}
			c.pushEach(collectionPath, n.Params)
			err := c.walkPartialsCollect(n.Body, goName, add)
			c.pop()
			if err != nil {
				return err
			}
			return c.walkPartialsCollect(n.Else, goName, add)
		default:
			// if/unless/helper: recurse without changing scope
			if err := c.walkPartialsCollect(n.Body, goName, add); err != nil {
				return err
			}
			return c.walkPartialsCollect(n.Else, goName, add)
		}
	default:
		return nil
	}
}

// buildPartialTypeSet returns for each partial the set of caller context types (from same-scope calls only).
// Used by CollectPartialParamTypes and SharedPartialInfo.
func buildPartialTypeSet(parsed map[string][]ast.Node, names []string, funcNames map[string]string, helperExprs map[string]string) map[string]map[string]bool {
	typeSet := make(map[string]map[string]bool) // partialName -> set of param types (from same-scope calls only)
	for _, name := range names {
		goName := funcNames[name]
		col := newPathCollector(helperExprs)
		col.setParsed(parsed)
		add := func(partialName, paramType string, sameScope bool) {
			if !sameScope {
				return // explicit context: partial keeps its own interface (e.g. OrderRowContext)
			}
			if typeSet[partialName] == nil {
				typeSet[partialName] = make(map[string]bool)
			}
			typeSet[partialName][paramType] = true
		}
		if err := col.walkPartialsCollect(parsed[name], goName, add); err != nil {
			continue
		}
	}
	// Propagate caller types transitively: if template T includes partial P and P is a template that includes Q,
	// then every caller of P (including T) should count as a caller of Q so Q becomes shared and gets a canonical type.
	for changed := true; changed; {
		changed = false
		for _, name := range names {
			if typeSet[name] == nil {
				continue
			}
			col := newPathCollector(helperExprs)
			col.setParsed(parsed)
			var included []string
			add := func(partialName, paramType string, sameScope bool) {
				if !sameScope {
					return
				}
				included = append(included, partialName)
			}
			if err := col.walkPartialsCollect(parsed[name], funcNames[name], add); err != nil {
				continue
			}
			for _, q := range included {
				if typeSet[q] == nil {
					typeSet[q] = make(map[string]bool)
				}
				for c := range typeSet[name] {
					if !typeSet[q][c] {
						typeSet[q][c] = true
						changed = true
					}
				}
			}
		}
	}
	// #region agent log
	if set := typeSet["menu"]; set != nil {
		callers := make([]string, 0, len(set))
		for c := range set {
			callers = append(callers, c)
		}
		sort.Strings(callers)
		debugLog("H1", "buildPartialTypeSet:after", "typeSet[menu] callers", map[string]interface{}{"count": len(set), "callers": callers})
	}
	if set := typeSet["default"]; set != nil {
		callers := make([]string, 0, len(set))
		for c := range set {
			callers = append(callers, c)
		}
		sort.Strings(callers)
		debugLog("H4", "buildPartialTypeSet:after", "typeSet[default] callers", map[string]interface{}{"count": len(set), "callers": callers})
	}
	if set := typeSet["news/informer"]; set != nil {
		callers := make([]string, 0, len(set))
		for c := range set {
			callers = append(callers, c)
		}
		sort.Strings(callers)
		debugLog("H5", "buildPartialTypeSet:after", "typeSet[news/informer] callers", map[string]interface{}{"count": len(set), "callers": callers})
	}
	// Log all typeSet keys to see partial names (e.g. news/informer vs news.informer).
	allKeys := make([]string, 0, len(typeSet))
	for k := range typeSet {
		allKeys = append(allKeys, k)
	}
	sort.Strings(allKeys)
	debugLog("H8", "buildPartialTypeSet:after", "typeSet all keys", map[string]interface{}{"keys": allKeys})
	// #endregion
	return typeSet
}

// CollectPartialParamTypes returns for each partial the context type to use for its render param.
// Only same-scope calls ({{> name}} with no expr) contribute: then we use the caller's type so partial and caller share one interface.
// When a partial is called with explicit context (e.g. {{> orderRow order}}), we do not set result[partialName], so the partial keeps its own context interface (e.g. OrderRowContext) and is called with the row data explicitly.
// Returns map[partialName]contextTypeName; empty string means use the partial's own context interface.
func CollectPartialParamTypes(parsed map[string][]ast.Node, names []string, funcNames map[string]string, helperExprs map[string]string) map[string]string {
	typeSet := buildPartialTypeSet(parsed, names, funcNames, helperExprs)
	result := make(map[string]string)
	for partialName, set := range typeSet {
		if len(set) == 1 {
			for t := range set {
				result[partialName] = t
				break
			}
		}
	}
	return result
}

// SharedPartialInfo returns for each shared partial (multiple same-scope callers) its canonical context type name
// and the primary caller template name (first alphabetically). Used so the compiler can emit one canonical interface
// per partial and have non-primary callers use it; primary keeps an embedding interface.
func SharedPartialInfo(typeSet map[string]map[string]bool, names []string, funcNames map[string]string) (canonicalType, primaryCaller map[string]string) {
	canonicalType = make(map[string]string)
	primaryCaller = make(map[string]string)
	for partialName, set := range typeSet {
		if len(set) <= 1 {
			// #region agent log
			if partialName == "menu" {
				debugLog("H2", "SharedPartialInfo:skip", "menu not shared len<=1", map[string]interface{}{"partialName": partialName, "setLen": len(set)})
			}
			// #endregion
			continue
		}
		goName := funcNames[partialName]
		if goName == "" {
			debugLog("H6", "SharedPartialInfo:skip", "partial has no goName", map[string]interface{}{"partialName": partialName, "setLen": len(set)})
			continue
		}
		canonicalType[partialName] = goName + "Context"
		if partialName == "news/informer" {
			debugLog("H7", "SharedPartialInfo:set", "canonicalType[news/informer]", map[string]interface{}{"canon": goName + "Context"})
		}
		var callers []string
		for _, name := range names {
			if funcNames[name]+"Context" == "" {
				continue
			}
			if set[funcNames[name]+"Context"] {
				callers = append(callers, name)
			}
		}
		sort.Strings(callers)
		if len(callers) > 0 {
			primaryCaller[partialName] = callers[0]
		}
	}
	// Log all canonicalType entries to verify news/informer and others.
	canonKeys := make([]string, 0, len(canonicalType))
	for k := range canonicalType {
		canonKeys = append(canonKeys, k)
	}
	sort.Strings(canonKeys)
	debugLog("H9", "SharedPartialInfo:after", "canonicalType keys", map[string]interface{}{"keys": canonKeys})
	return canonicalType, primaryCaller
}

// AllContextTypeNames returns all context interface names that would be emitted from this template's type tree.
// Used to map each context type to its owning template (e.g. MainOrdersOrderItemContext -> "main").
func AllContextTypeNames(goName string, tree *typeNode) []string {
	var out []string
	seen := make(map[string]bool)
	var visit func(pathPrefix string, n *typeNode)
	visit = func(pathPrefix string, n *typeNode) {
		if n == nil {
			return
		}
		if pathPrefix == "" {
			out = append(out, goName+"Context")
			seen[goName+"Context"] = true
		}
		if n.isSlice && n.sliceElem != nil {
			elemName := contextItemInterfaceName(goName, strings.TrimPrefix(pathPrefix, "."))
			if !seen[elemName] {
				seen[elemName] = true
				out = append(out, elemName)
			}
			visit(pathPrefix+".", n.sliceElem)
			return
		}
		if n.fields == nil {
			return
		}
		for field, child := range n.fields {
			if field == "" || (len(field) > 0 && (field[0] == '@' || field[0] == '.')) {
				continue
			}
			subPath := field
			if pathPrefix != "" {
				if strings.HasSuffix(pathPrefix, ".") {
					subPath = pathPrefix + field
				} else {
					subPath = pathPrefix + "." + field
				}
			}
			if child.isSlice {
				visit(subPath, child)
				continue
			}
			if len(child.fields) > 0 {
				ifaceName := contextInterfaceName(goName, subPath)
				if !seen[ifaceName] {
					seen[ifaceName] = true
					out = append(out, ifaceName)
				}
				visit(subPath+".", child)
			}
		}
	}
	visit("", tree)
	return out
}

// typeNode is a node in the inferred type tree (object fields or slice element).
type typeNode struct {
	fields    map[string]*typeNode
	sliceElem *typeNode
	isSlice   bool
}

func buildTypeTree(paths map[string]bool, eachFields map[string]map[string]bool) *typeNode {
	root := &typeNode{fields: make(map[string]*typeNode)}
	for p := range paths {
		if p == "" || p == "." || (len(p) > 0 && (p[0] == '@' || p[0] == '.')) {
			continue
		}
		parts := strings.Split(p, ".")
		if len(parts) == 0 {
			continue
		}
		cur := root
		for i, part := range parts {
			if part == "" || part == "." || (len(part) > 0 && (part[0] == '@' || part[0] == '.')) {
				continue
			}
			if cur.fields == nil {
				cur.fields = make(map[string]*typeNode)
			}
			if cur.fields[part] == nil {
				cur.fields[part] = &typeNode{}
			}
			cur = cur.fields[part]
			if i == len(parts)-1 {
				// leaf: keep as any (cur.fields empty, no slice)
			}
		}
	}
	for col, fields := range eachFields {
		if col == "" || col == "." || (len(col) > 0 && (col[0] == '@' || col[0] == '.')) {
			continue
		}
		parts := strings.Split(col, ".")
		cur := root
		for _, part := range parts {
			if part == "" || part == "." || (len(part) > 0 && (part[0] == '@' || part[0] == '.')) {
				continue
			}
			if cur.fields == nil {
				cur.fields = make(map[string]*typeNode)
			}
			if cur.fields[part] == nil {
				cur.fields[part] = &typeNode{}
			}
			cur = cur.fields[part]
		}
		cur.isSlice = true
		cur.sliceElem = &typeNode{fields: make(map[string]*typeNode)}
		for f := range fields {
			if f == "" || f == "." || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
				continue
			}
			cur.sliceElem.fields[f] = &typeNode{}
		}
	}
	return root
}

// goFieldName returns a Go-style method name for a field (e.g. "user_name" -> "UserName", "layout/header" -> "LayoutHeader").
// Splits on non-alphanumeric runes so path separators and underscores produce a single PascalCase identifier.
func goFieldName(field string) string {
	parts := strings.FieldsFunc(field, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9')
	})
	var sb strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		sb.WriteString(capitalize(part))
	}
	return sb.String()
}

func contextInterfaceName(templateIdent, fieldPath string) string {
	if fieldPath == "" {
		return templateIdent + "Context"
	}
	parts := strings.Split(fieldPath, ".")
	var sb strings.Builder
	sb.WriteString(templateIdent)
	for _, p := range parts {
		sb.WriteString(goFieldName(p))
	}
	sb.WriteString("Context")
	return sb.String()
}

// used for element-of-slice interface naming
func contextItemInterfaceName(templateIdent, collectionPath string) string {
	parts := strings.Split(collectionPath, ".")
	var sb strings.Builder
	sb.WriteString(templateIdent)
	for _, p := range parts {
		sb.WriteString(goFieldName(p))
	}
	sb.WriteString("ItemContext")
	return sb.String()
}

// contextDataStructName returns the generated struct name for a context interface (e.g. MainContext -> MainContextData).
func contextDataStructName(ifaceName string) string {
	return ifaceName + "Data"
}

// resolvePathToMethodChain returns the method chain (e.g. "User().Name()") to resolve path from the current scope.
// pathPrefix is the current scope path (e.g. "" at root, "user" in {{#with user}}).
// path is the expression path from the template (e.g. "title", "user.name").
// Returns ("", false) for paths with "..", "@", or when path is not under the typed tree (e.g. slice element).
func resolvePathToMethodChain(node *typeNode, pathPrefix, path, goIdent string) (methodChain string, ok bool) {
	path = strings.TrimSpace(path)
	if path == "" || path == "." || path == "this" {
		return "", true // current context: caller should use current var as-is
	}
	if pathPrefix != "" && path == pathPrefix {
		return "", true // same as current scope: use current var (e.g. {{date}} inside {{#date}})
	}
	if strings.HasPrefix(path, "..") || strings.HasPrefix(path, "@") || strings.HasPrefix(path, "./") {
		return "", false
	}
	fullPath := path
	if pathPrefix != "" {
		if !strings.HasPrefix(path, pathPrefix+".") && path != pathPrefix {
			// Relative path: "name" when pathPrefix is "user" -> fullPath "user.name"
			fullPath = pathPrefix + "." + path
		}
	}
	relativePath := fullPath
	if pathPrefix != "" && strings.HasPrefix(fullPath, pathPrefix+".") {
		relativePath = fullPath[len(pathPrefix)+1:]
	}
	parts := strings.Split(relativePath, ".")
	if len(parts) == 0 {
		return "", true
	}
	cur := node
	var chain []string
	for i, segment := range parts {
		if segment == "" || (len(segment) > 0 && (segment[0] == '@' || segment[0] == '.')) {
			return "", false
		}
		if cur == nil || cur.fields == nil {
			return "", false
		}
		child, ok := cur.fields[segment]
		if !ok {
			return "", false
		}
		methodName := goFieldName(segment)
		if methodName == "" {
			return "", false
		}
		// Slice is allowed only as the last segment (e.g. "users" in {{#each users}}) so we get data.Users().
		if child.isSlice && child.sliceElem != nil && i < len(parts)-1 {
			return "", false
		}
		chain = append(chain, methodName+"()")
		if i < len(parts)-1 {
			if len(child.fields) == 0 {
				return "", false
			}
			cur = child
		}
	}
	return strings.Join(chain, "."), true
}

// nodeAtPath returns the type node at the given path from root, or nil.
func nodeAtPath(root *typeNode, path string) *typeNode {
	path = strings.TrimSpace(path)
	if path == "" || strings.HasPrefix(path, "..") || strings.HasPrefix(path, "@") {
		return nil
	}
	parts := strings.Split(path, ".")
	cur := root
	for _, segment := range parts {
		if segment == "" || (len(segment) > 0 && (segment[0] == '@' || segment[0] == '.')) {
			return nil
		}
		if cur == nil || cur.fields == nil {
			return nil
		}
		child, ok := cur.fields[segment]
		if !ok {
			return nil
		}
		cur = child
	}
	return cur
}

// RootFieldKeys returns root-level field names from the type tree (keys used at root context).
func RootFieldKeys(tree *typeNode) []string {
	if tree == nil || tree.fields == nil {
		return nil
	}
	var keys []string
	for k := range tree.fields {
		if k == "" || (len(k) > 0 && (k[0] == '@' || k[0] == '.')) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// rootLayoutPartial returns the single root-level shared partial name when the root template
// has exactly one root-level field that is a shared partial (in canonicalType). Otherwise returns "".
// Used to emit root contexts that embed the layout context and only template-specific methods.
func rootLayoutPartial(tree *typeNode, canonicalType map[string]string) string {
	if tree == nil || tree.fields == nil || canonicalType == nil {
		return ""
	}
	var layout string
	for f := range tree.fields {
		if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		if canonicalType[f] != "" {
			if layout != "" {
				return "" // more than one shared partial at root
			}
			layout = f
		}
	}
	return layout
}

func emitContextInterfaces(w *codeWriter, templateName, goName string, tree *typeNode, canonicalType, primaryCaller map[string]string, typeTrees map[string]*typeNode) {
	rootName := goName + "Context"
	seen := make(map[string]bool)
	seen[rootName] = true
	layoutPartial := rootLayoutPartial(tree, canonicalType)
	if layoutPartial == templateName {
		layoutPartial = ""
	}
	var skipFieldsAtRoot map[string]bool
	if layoutPartial != "" && typeTrees != nil {
		// Use the layout template's root-level fields so we skip all layout content (e.g. news/informer
		// merged from default) even when the current tree stores the layout as a leaf.
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
	w.line("")
	w.line("// %s is the context interface inferred from template %q.", rootName, templateName)
	w.line("type %s interface {", rootName)
	w.indentInc()
	if layoutPartial != "" {
		layoutCanon := canonicalType[layoutPartial]
		w.line("%s", layoutCanon)
	}
	emitInterfaceMethods(w, templateName, goName, "", tree, seen, canonicalType, primaryCaller, layoutPartial, skipFieldsAtRoot, typeTrees)
	w.line("Raw() any")
	w.indentDec()
	w.line("}")
	emitNodeInterfaces(w, templateName, goName, "", tree, seen, canonicalType, primaryCaller, typeTrees)
}

func emitNodeInterfaces(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool, canonicalType, primaryCaller map[string]string, typeTrees map[string]*typeNode) {
	if n == nil {
		return
	}
	if n.isSlice && n.sliceElem != nil {
		elemName := contextItemInterfaceName(goIdent, pathPrefix)
		if seen[elemName] {
			return
		}
		seen[elemName] = true
		w.line("")
		w.line("// %s is the context for one element of %s.", elemName, pathPrefix)
		w.line("type %s interface {", elemName)
		w.indentInc()
		emitInterfaceMethods(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen, canonicalType, primaryCaller, "", nil, typeTrees)
		w.line("Raw() any")
		w.indentDec()
		w.line("}")
		emitNodeInterfaces(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen, canonicalType, primaryCaller, typeTrees)
		return
	}
	if n.fields == nil {
		return
	}
	for field, child := range n.fields {
		if field == "" || field == "." || (len(field) > 0 && (field[0] == '@' || field[0] == '.')) {
			continue
		}
		subPath := field
		if pathPrefix != "" {
			subPath = pathPrefix + "." + field
		}
		if child.isSlice {
			emitNodeInterfaces(w, templateName, goIdent, subPath, child, seen, canonicalType, primaryCaller, typeTrees)
			continue
		}
		if len(child.fields) > 0 {
			// Shared partial at this path: skip emitting nested interface (canonical type emitted by partial's tree).
			if canonicalType != nil && canonicalType[field] != "" {
				continue
			}
			// Under a shared partial path (e.g. default.shared): skip; canonical partial's tree emits the type.
			if pathPrefix != "" && canonicalType != nil {
				firstSeg := strings.TrimSuffix(pathPrefix, ".")
				if idx := strings.Index(firstSeg, "."); idx >= 0 {
					firstSeg = firstSeg[:idx]
				}
				if canonicalType[firstSeg] != "" {
					continue
				}
			}
			ifaceName := contextInterfaceName(goIdent, subPath)
			if seen[ifaceName] {
				continue
			}
			seen[ifaceName] = true
			w.line("")
			w.line("// %s is the context for path %q.", ifaceName, subPath)
			w.line("type %s interface {", ifaceName)
			w.indentInc()
			emitInterfaceMethods(w, templateName, goIdent, subPath+".", child, seen, canonicalType, primaryCaller, "", nil, typeTrees)
			w.line("Raw() any")
			w.indentDec()
			w.line("}")
			emitNodeInterfaces(w, templateName, goIdent, subPath, child, seen, canonicalType, primaryCaller, typeTrees)
		}
	}
}

func emitInterfaceMethods(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool, canonicalType, primaryCaller map[string]string, layoutPartial string, skipFieldsAtRoot map[string]bool, typeTrees map[string]*typeNode) {
	if n == nil || n.fields == nil {
		return
	}
	var names []string
	for f := range n.fields {
		if f == "" || f == "." || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		names = append(names, f)
	}
	sort.Strings(names)
	for _, field := range names {
		if pathPrefix == "" && layoutPartial != "" && field == layoutPartial {
			continue
		}
		if pathPrefix == "" && skipFieldsAtRoot != nil && skipFieldsAtRoot[field] {
			continue
		}
		child := n.fields[field]
		methodName := goFieldName(field)
		if methodName == "" {
			continue
		}
		if child.isSlice && child.sliceElem != nil {
			fullPath := pathPrefix + field
			elemIdent := goIdent
			elemPath := fullPath
			if pathPrefix != "" && canonicalType != nil {
				firstSeg := pathPrefix
				if idx := strings.Index(firstSeg, "."); idx >= 0 {
					firstSeg = firstSeg[:idx]
				} else {
					firstSeg = strings.TrimSuffix(firstSeg, ".")
				}
				if canonRoot := canonicalType[firstSeg]; canonRoot != "" {
					elemIdent = strings.TrimSuffix(canonRoot, "Context")
					if strings.HasPrefix(fullPath, firstSeg+".") {
						elemPath = fullPath[len(firstSeg)+1:]
					}
				}
			}
			elemName := contextItemInterfaceName(elemIdent, elemPath)
			w.line("%s() []%s", methodName, elemName)
			continue
		}
		if len(child.fields) > 0 {
			canon := ""
			if canonicalType != nil {
				canon = canonicalType[field]
			}
			// #region agent log
			if field == "menu" && pathPrefix == "" {
				debugLog("H3", "emitInterfaceMethods:menu", "choosing menu return type", map[string]interface{}{"templateName": templateName, "canon": canon, "goIdent": goIdent})
			}
			// #endregion
			var ifaceName string
			if canon != "" {
				// Always use canonical type for shared partials (no primary-embedding types).
				ifaceName = canon
			} else if pathPrefix == "" && typeTrees != nil && n.fields != nil {
				// At root: if this field is a root-level field of a root-level shared partial (e.g. default has news/informer),
				// use that partial's ident so the root implements the layout's interface.
				for k := range n.fields {
					if canonicalType == nil || canonicalType[k] == "" {
						continue
					}
					otherTree := typeTrees[k]
					if otherTree == nil || otherTree.fields == nil || otherTree.fields[field] == nil {
						continue
					}
					ifaceName = contextInterfaceName(strings.TrimSuffix(canonicalType[k], "Context"), field)
					break
				}
			}
			if ifaceName == "" {
				if pathPrefix != "" && canonicalType != nil {
					// Nested path under a shared partial: use type from canonical partial's tree.
					firstSeg := pathPrefix
					if idx := strings.Index(firstSeg, "."); idx >= 0 {
						firstSeg = firstSeg[:idx]
					} else {
						firstSeg = strings.TrimSuffix(firstSeg, ".")
					}
					if canonRoot := canonicalType[firstSeg]; canonRoot != "" {
						canonGoName := strings.TrimSuffix(canonRoot, "Context")
						subPathUnderCanon := pathPrefix + field
						if strings.HasPrefix(subPathUnderCanon, firstSeg+".") {
							subPathUnderCanon = subPathUnderCanon[len(firstSeg)+1:]
						}
						ifaceName = contextInterfaceName(canonGoName, subPathUnderCanon)
					} else {
						ifaceName = contextInterfaceName(goIdent, pathPrefix+field)
					}
				} else {
					ifaceName = contextInterfaceName(goIdent, pathPrefix+field)
				}
			}
			w.line("%s() %s", methodName, ifaceName)
			continue
		}
		w.line("%s() any", methodName)
	}
}

// emitContextDataTypes emits all ...ContextData structs and FromMap for the given template.
func emitContextDataTypes(w *codeWriter, templateName, goName string, tree *typeNode, canonicalType, primaryCaller map[string]string, typeTrees map[string]*typeNode) {
	seen := make(map[string]bool)
	emitNodeContextDataTypes(w, templateName, goName, "", tree, seen, canonicalType)
	rootName := goName + "Context"
	rootDataName := contextDataStructName(rootName)
	layoutPartial := rootLayoutPartial(tree, canonicalType)
	if layoutPartial == templateName {
		layoutPartial = ""
	}
	var layoutTree *typeNode
	if layoutPartial != "" && typeTrees != nil {
		layoutTree = typeTrees[layoutPartial]
	}
	emitRootContextDataStruct(w, templateName, goName, tree, rootDataName, canonicalType, primaryCaller, layoutPartial, layoutTree, typeTrees)
	emitFromMap(w, goName, rootName, rootDataName)
}

func emitNodeContextDataTypes(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool, canonicalType map[string]string) {
	if n == nil {
		return
	}
	// Recurse first so nested/item types are emitted before types that depend on them.
	if n.fields != nil {
		var names []string
		for f := range n.fields {
			if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
				continue
			}
			names = append(names, f)
		}
		sort.Strings(names)
		for _, field := range names {
			child := n.fields[field]
			subPath := field
			if pathPrefix != "" {
				subPath = pathPrefix + "." + field
			}
			emitNodeContextDataTypes(w, templateName, goIdent, subPath, child, seen, canonicalType)
		}
	}
	if n.isSlice && n.sliceElem != nil && pathPrefix != "" {
		emitNodeContextDataTypes(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen, canonicalType)
	}
	// Emit this node's ContextData (skip root; root is emitted by emitRootContextDataStruct).
	if pathPrefix == "" {
		return
	}
	// Skip shared partial path: canonical partial already has its ContextData/FromMap; callers use it.
	if canonicalType != nil && canonicalType[pathPrefix] != "" {
		return
	}
	// Skip path under a shared partial (e.g. default.shared): canonical partial's tree emits it.
	if pathPrefix != "" && canonicalType != nil {
		firstSeg := pathPrefix
		if idx := strings.Index(firstSeg, "."); idx >= 0 {
			firstSeg = firstSeg[:idx]
		}
		if canonicalType[firstSeg] != "" {
			return
		}
	}
	if n.isSlice && n.sliceElem != nil {
		elemName := contextItemInterfaceName(goIdent, pathPrefix)
		dataName := contextDataStructName(elemName)
		if seen[dataName] {
			return
		}
		seen[dataName] = true
		emitItemContextDataStruct(w, goIdent, pathPrefix, n.sliceElem, dataName, elemName)
		return
	}
	// pathPrefix ending with "." means we're the slice element node (child of a slice); we already emitted the item type above.
	if pathPrefix != "" && strings.HasSuffix(pathPrefix, ".") {
		return
	}
	// Skip object context when we already emitted item context for this path (e.g. "orders" as slice -> MainOrdersItemContext; don't also emit MainOrdersContext).
	if pathPrefix != "" {
		itemDataName := contextDataStructName(contextItemInterfaceName(goIdent, pathPrefix))
		if seen[itemDataName] {
			return
		}
	}
	if len(n.fields) > 0 {
		ifaceName := contextInterfaceName(goIdent, pathPrefix)
		dataName := contextDataStructName(ifaceName)
		if seen[dataName] {
			return
		}
		seen[dataName] = true
		emitObjectContextDataStruct(w, templateName, goIdent, pathPrefix+".", n, dataName, ifaceName)
	}
}

func emitObjectContextDataStruct(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, dataName, ifaceName string) {
	w.line("")
	w.line("// %s is a map-backed implementation of %s.", dataName, ifaceName)
	w.line("type %s struct { m map[string]any }", dataName)
	w.line("")
	emitContextDataMethods(w, templateName, goIdent, pathPrefix, n, dataName, nil, nil, "", nil, nil)
	w.line("func (d %s) Raw() any { return d.m }", dataName)
	emitFromMapForIface(w, ifaceName, dataName)
}

func emitItemContextDataStruct(w *codeWriter, goIdent, collectionPath string, n *typeNode, dataName, ifaceName string) {
	w.line("")
	w.line("// %s is a map-backed implementation of %s.", dataName, ifaceName)
	w.line("type %s struct { m map[string]any }", dataName)
	w.line("")
	pathPrefix := collectionPath + "."
	emitContextDataMethods(w, "", goIdent, pathPrefix, n, dataName, nil, nil, "", nil, nil)
	w.line("func (d %s) Raw() any { return d.m }", dataName)
	emitFromMapForIface(w, ifaceName, dataName)
}

func emitContextDataMethods(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, dataName string, canonicalType, primaryCaller map[string]string, layoutPartial string, skipFieldsAtRoot map[string]bool, typeTrees map[string]*typeNode) {
	if n == nil || n.fields == nil {
		return
	}
	var names []string
	for f := range n.fields {
		if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
			continue
		}
		names = append(names, f)
	}
	sort.Strings(names)
	for _, field := range names {
		if pathPrefix == "" && layoutPartial != "" && field == layoutPartial {
			continue
		}
		if pathPrefix == "" && skipFieldsAtRoot != nil && skipFieldsAtRoot[field] {
			continue
		}
		child := n.fields[field]
		methodName := goFieldName(field)
		if methodName == "" {
			continue
		}
		mapKey := field
		if child.isSlice && child.sliceElem != nil {
			fullPath := pathPrefix + field
			elemIdent := goIdent
			elemPath := fullPath
			if pathPrefix != "" && canonicalType != nil {
				firstSeg := pathPrefix
				if idx := strings.Index(firstSeg, "."); idx >= 0 {
					firstSeg = firstSeg[:idx]
				} else {
					firstSeg = strings.TrimSuffix(firstSeg, ".")
				}
				if canonRoot := canonicalType[firstSeg]; canonRoot != "" {
					elemIdent = strings.TrimSuffix(canonRoot, "Context")
					if strings.HasPrefix(fullPath, firstSeg+".") {
						elemPath = fullPath[len(firstSeg)+1:]
					}
				}
			}
			elemName := contextItemInterfaceName(elemIdent, elemPath)
			elemDataName := contextDataStructName(elemName)
			w.line("func (d %s) %s() []%s {", dataName, methodName, elemName)
			w.indentInc()
			w.line("v := d.m[%q]", mapKey)
			w.line("if v == nil { return nil }")
			w.line("s, ok := v.([]any)")
			w.line("if !ok { return nil }")
			w.line("out := make([]%s, len(s))", elemName)
			w.line("for i := range s {")
			w.indentInc()
			w.line("if m, ok := s[i].(map[string]any); ok {")
			w.indentInc()
			w.line("out[i] = %s{m}", elemDataName)
			w.indentDec()
			w.line("}")
			w.indentDec()
			w.line("}")
			w.line("return out")
			w.indentDec()
			w.line("}")
			continue
		}
		if len(child.fields) > 0 {
			canon := ""
			if canonicalType != nil {
				canon = canonicalType[field]
			}
			var ifaceName, nestedDataName string
			if canon != "" {
				// Always use canonical type for shared partials (no primary-embedding types).
				ifaceName = canon
				nestedDataName = contextDataStructName(canon)
			} else if pathPrefix == "" && typeTrees != nil && n.fields != nil {
				// At root: if this field is a root-level field of a root-level shared partial, use that partial's ident.
				for k := range n.fields {
					if canonicalType == nil || canonicalType[k] == "" {
						continue
					}
					otherTree := typeTrees[k]
					if otherTree == nil || otherTree.fields == nil || otherTree.fields[field] == nil {
						continue
					}
					ifaceName = contextInterfaceName(strings.TrimSuffix(canonicalType[k], "Context"), field)
					nestedDataName = contextDataStructName(ifaceName)
					break
				}
			}
			if ifaceName == "" {
				ifaceName = contextInterfaceName(goIdent, pathPrefix+field)
				nestedDataName = contextDataStructName(ifaceName)
			}
			w.line("func (d %s) %s() %s {", dataName, methodName, ifaceName)
			w.indentInc()
			w.line("v := d.m[%q]", mapKey)
			w.line("if v == nil { return %s{map[string]any{}} }", nestedDataName)
			w.line("m, ok := v.(map[string]any)")
			w.line("if !ok { return %s{map[string]any{}} }", nestedDataName)
			w.line("return %s{m}", nestedDataName)
			w.indentDec()
			w.line("}")
			continue
		}
		w.line("func (d %s) %s() any { return d.m[%q] }", dataName, methodName, mapKey)
	}
}

func emitRootContextDataStruct(w *codeWriter, templateName, goIdent string, tree *typeNode, rootDataName string, canonicalType, primaryCaller map[string]string, layoutPartial string, layoutTree *typeNode, typeTrees map[string]*typeNode) {
	rootName := goIdent + "Context"
	w.line("")
	w.line("// %s is a map-backed implementation of %s.", rootDataName, rootName)
	w.line("type %s struct { m map[string]any }", rootDataName)
	w.line("")
	var skipFieldsAtRoot map[string]bool
	if layoutPartial != "" && layoutTree != nil {
		layoutCanon := canonicalType[layoutPartial]
		layoutGoName := strings.TrimSuffix(layoutCanon, "Context")
		emitContextDataMethods(w, layoutPartial, layoutGoName, "", layoutTree, rootDataName, canonicalType, primaryCaller, "", nil, nil)
		skipFieldsAtRoot = make(map[string]bool)
		skipFieldsAtRoot[layoutPartial] = true
		if layoutTree.fields != nil {
			for f := range layoutTree.fields {
				if f == "" || (len(f) > 0 && (f[0] == '@' || f[0] == '.')) {
					continue
				}
				skipFieldsAtRoot[f] = true
			}
		}
	}
	emitContextDataMethods(w, templateName, goIdent, "", tree, rootDataName, canonicalType, primaryCaller, layoutPartial, skipFieldsAtRoot, typeTrees)
	w.line("func (d %s) Raw() any { return d.m }", rootDataName)
}

func emitFromMap(w *codeWriter, goIdent, rootIfaceName, rootDataName string) {
	emitFromMapForIface(w, rootIfaceName, rootDataName)
}

// emitFromMapForIface emits a FromMap constructor for any context interface (root or nested).
func emitFromMapForIface(w *codeWriter, ifaceName, dataName string) {
	fromMapName := ifaceName + "FromMap"
	w.line("")
	w.line("// %s returns a %s backed by m. If m is nil, a new empty map is used.", fromMapName, ifaceName)
	w.line("func %s(m map[string]any) %s {", fromMapName, ifaceName)
	w.indentInc()
	w.line("if m == nil { m = make(map[string]any) }")
	w.line("return %s{m}", dataName)
	w.indentDec()
	w.line("}")
}
