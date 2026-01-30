package compiler

import (
	"sort"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
)

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
	parts, _, err := parseParts(n.Expr)
	if err != nil {
		return nil
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
			if partialNodes, ok := c.parsed[partialName]; ok {
				if err := c.collectNodes(partialNodes); err != nil {
					return err
				}
			}
		}
	}
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

// CollectPartialParamTypes returns for each partial the context type to use for its render param.
// Only same-scope calls ({{> name}} with no expr) contribute: then we use the caller's type so partial and caller share one interface.
// When a partial is called with explicit context (e.g. {{> orderRow order}}), we do not set result[partialName], so the partial keeps its own context interface (e.g. OrderRowContext) and is called with the row data explicitly.
// Returns map[partialName]contextTypeName; empty string means use the partial's own context interface.
func CollectPartialParamTypes(parsed map[string][]ast.Node, names []string, funcNames map[string]string, helperExprs map[string]string) map[string]string {
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

// goFieldName returns a Go-style method name for a field (e.g. "user_name" -> "UserName").
func goFieldName(field string) string {
	return capitalize(strings.ReplaceAll(field, " ", ""))
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

func emitContextInterfaces(w *codeWriter, templateName string, tree *typeNode) {
	goName := goIdent(templateName)
	rootName := goName + "Context"
	seen := make(map[string]bool)
	seen[rootName] = true
	w.line("")
	w.line("// %s is the context interface inferred from template %q.", rootName, templateName)
	w.line("type %s interface {", rootName)
	w.indentInc()
	emitInterfaceMethods(w, templateName, goName, "", tree, seen)
	w.line("Raw() any")
	w.indentDec()
	w.line("}")
	emitNodeInterfaces(w, templateName, goName, "", tree, seen)
}

func emitNodeInterfaces(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool) {
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
		emitInterfaceMethods(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen)
		w.line("Raw() any")
		w.indentDec()
		w.line("}")
		emitNodeInterfaces(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen)
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
			emitNodeInterfaces(w, templateName, goIdent, subPath, child, seen)
			continue
		}
		if len(child.fields) > 0 {
			ifaceName := contextInterfaceName(goIdent, subPath)
			if seen[ifaceName] {
				continue
			}
			seen[ifaceName] = true
			w.line("")
			w.line("// %s is the context for path %q.", ifaceName, subPath)
			w.line("type %s interface {", ifaceName)
			w.indentInc()
			emitInterfaceMethods(w, templateName, goIdent, subPath+".", child, seen)
			w.line("Raw() any")
			w.indentDec()
			w.line("}")
			emitNodeInterfaces(w, templateName, goIdent, subPath, child, seen)
		}
	}
}

func emitInterfaceMethods(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool) {
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
		child := n.fields[field]
		methodName := goFieldName(field)
		if methodName == "" {
			continue
		}
		if child.isSlice && child.sliceElem != nil {
			elemName := contextItemInterfaceName(goIdent, pathPrefix+field)
			w.line("%s() []%s", methodName, elemName)
			continue
		}
		if len(child.fields) > 0 {
			ifaceName := contextInterfaceName(goIdent, pathPrefix+field)
			w.line("%s() %s", methodName, ifaceName)
			continue
		}
		w.line("%s() any", methodName)
	}
}

// emitContextDataTypes emits all ...ContextData structs and FromMap for the given template.
func emitContextDataTypes(w *codeWriter, templateName string, tree *typeNode) {
	goName := goIdent(templateName)
	seen := make(map[string]bool)
	emitNodeContextDataTypes(w, templateName, goName, "", tree, seen)
	rootName := goName + "Context"
	rootDataName := contextDataStructName(rootName)
	emitRootContextDataStruct(w, templateName, goName, tree, rootDataName)
	emitFromMap(w, goName, rootName, rootDataName)
}

func emitNodeContextDataTypes(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, seen map[string]bool) {
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
			emitNodeContextDataTypes(w, templateName, goIdent, subPath, child, seen)
		}
	}
	if n.isSlice && n.sliceElem != nil && pathPrefix != "" {
		emitNodeContextDataTypes(w, templateName, goIdent, pathPrefix+".", n.sliceElem, seen)
	}
	// Emit this node's ContextData (skip root; root is emitted by emitRootContextDataStruct).
	if pathPrefix == "" {
		return
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
	if n.fields != nil && len(n.fields) > 0 {
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
	emitContextDataMethods(w, templateName, goIdent, pathPrefix, n, dataName)
	w.line("func (d %s) Raw() any { return d.m }", dataName)
	emitFromMapForIface(w, ifaceName, dataName)
}

func emitItemContextDataStruct(w *codeWriter, goIdent, collectionPath string, n *typeNode, dataName, ifaceName string) {
	w.line("")
	w.line("// %s is a map-backed implementation of %s.", dataName, ifaceName)
	w.line("type %s struct { m map[string]any }", dataName)
	w.line("")
	pathPrefix := collectionPath + "."
	emitContextDataMethods(w, "", goIdent, pathPrefix, n, dataName)
	w.line("func (d %s) Raw() any { return d.m }", dataName)
	emitFromMapForIface(w, ifaceName, dataName)
}

func emitContextDataMethods(w *codeWriter, templateName, goIdent, pathPrefix string, n *typeNode, dataName string) {
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
		child := n.fields[field]
		methodName := goFieldName(field)
		if methodName == "" {
			continue
		}
		mapKey := field
		if child.isSlice && child.sliceElem != nil {
			elemName := contextItemInterfaceName(goIdent, pathPrefix+field)
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
			ifaceName := contextInterfaceName(goIdent, pathPrefix+field)
			nestedDataName := contextDataStructName(ifaceName)
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

func emitRootContextDataStruct(w *codeWriter, templateName, goIdent string, tree *typeNode, rootDataName string) {
	rootName := goIdent + "Context"
	w.line("")
	w.line("// %s is a map-backed implementation of %s.", rootDataName, rootName)
	w.line("type %s struct { m map[string]any }", rootDataName)
	w.line("")
	emitContextDataMethods(w, templateName, goIdent, "", tree, rootDataName)
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
