package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
	"github.com/andriyg76/hexerr"
)

// lineOf returns the 1-based line number for the given offset in input.
func lineOf(input string, offset int) int {
	if offset <= 0 {
		return 1
	}
	if offset > len(input) {
		offset = len(input)
	}
	line := 1
	for i := 0; i < offset; i++ {
		if input[i] == '\n' {
			line++
		}
	}
	return line
}

// parserErr returns an error with " (line N)" appended when input and offset are available.
func parserErr(input string, offset int, msg string) error {
	return hexerr.New(msg + " (line " + strconv.Itoa(lineOf(input, offset)) + ")")
}

// parserErrWrap adds " (line N)" to an existing error's message.
func parserErrWrap(input string, offset int, err error) error {
	return hexerr.New(err.Error() + " (line " + strconv.Itoa(lineOf(input, offset)) + ")")
}

// Parse turns a template string into a list of nodes.
func Parse(input string) ([]ast.Node, error) {
	nodes, _, err := parseUntil(input, 0, "")
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

type stopKind int

const (
	stopNone stopKind = iota
	stopElse
	stopEnd
)

func parseUntil(input string, start int, endBlock string) ([]ast.Node, int, error) {
	nodes, next, stop, err := parseUntilStop(input, start, endBlock)
	if err != nil {
		return nil, 0, err
	}
	if stop != stopNone {
		return nil, 0, parserErr(input, next, fmt.Sprintf("parser: unexpected %s", stopLabel(stop, endBlock)))
	}
	return nodes, next, nil
}

func parseUntilStop(input string, start int, endBlock string) ([]ast.Node, int, stopKind, error) {
	var nodes []ast.Node
	i := start
	for i < len(input) {
		open := strings.Index(input[i:], "{{")
		if open < 0 {
			if i < len(input) {
				nodes = append(nodes, &ast.Text{Value: input[i:]})
			}
			if endBlock != "" {
				return nil, 0, stopNone, parserErr(input, i, fmt.Sprintf("parser: unclosed block %q", endBlock))
			}
			return nodes, len(input), stopNone, nil
		}
		open += i
		if open > i {
			nodes = append(nodes, &ast.Text{Value: input[i:open]})
		}

		if strings.HasPrefix(input[open:], "{{{{") {
			next, err := parseRawBlock(input, open, &nodes)
			if err != nil {
				return nil, 0, stopNone, err
			}
			i = next
			continue
		}

		if strings.HasPrefix(input[open:], "{{!--") || strings.HasPrefix(input[open:], "{{~!--") {
			trimLeft := strings.HasPrefix(input[open:], "{{~!--")
			start := open + 4
			if trimLeft {
				start++
				trimRightText(&nodes)
			}
			end := strings.Index(input[start:], "--}}")
			if end < 0 {
				return nil, 0, stopNone, parserErr(input, open, "parser: unclosed comment")
			}
			endPos := start + end
			trimRight := false
			if endPos > start && input[endPos-1] == '~' {
				trimRight = true
				endPos--
			}
			i = endPos + len("--}}")
			if trimRight {
				i = skipWhitespace(input, i)
			}
			continue
		}

		raw := false
		startLen := 2
		endDelim := "}}"
		if strings.HasPrefix(input[open:], "{{{") {
			raw = true
			startLen = 3
			endDelim = "}}}"
		}
		trimLeft := false
		if open+startLen < len(input) && input[open+startLen] == '~' {
			trimLeft = true
			startLen++
		}
		if trimLeft {
			trimRightText(&nodes)
		}
		end := strings.Index(input[open+startLen:], endDelim)
		if end < 0 {
			return nil, 0, stopNone, parserErr(input, open, "parser: unclosed mustache")
		}
		content := input[open+startLen : open+startLen+end]
		trimRight := false
		if strings.HasSuffix(content, "~") {
			trimRight = true
			content = strings.TrimSpace(strings.TrimSuffix(content, "~"))
		} else {
			content = strings.TrimSpace(content)
		}
		i = open + startLen + end + len(endDelim)
		if trimRight {
			i = skipWhitespace(input, i)
		}
		if content == "" {
			continue
		}
		if !raw && strings.HasPrefix(content, "&") {
			raw = true
			content = strings.TrimSpace(content[1:])
		}
		if strings.HasPrefix(content, "!") {
			continue
		}
		if content == "else" {
			if endBlock == "" {
				return nil, 0, stopNone, parserErr(input, open, "parser: unexpected else")
			}
			return nodes, i, stopElse, nil
		}
		if strings.HasPrefix(content, "/") {
			name := strings.TrimSpace(content[1:])
			if name == "" {
				return nil, 0, stopNone, parserErr(input, open, "parser: empty block name")
			}
			if endBlock == "" {
				return nil, 0, stopNone, parserErr(input, open, fmt.Sprintf("parser: unexpected closing block %q", name))
			}
			if name != endBlock {
				return nil, 0, stopNone, parserErr(input, open, fmt.Sprintf("parser: expected /%s, got /%s", endBlock, name))
			}
			return nodes, i, stopEnd, nil
		}
		if strings.HasPrefix(content, "#>") {
			return nil, 0, stopNone, parserErr(input, open, "parser: partial blocks ({{#>}}) are not supported")
		}
		if strings.HasPrefix(content, "#") {
			name, argsStr, blockParams, err := splitBlockStart(content[1:])
			if err != nil {
				return nil, 0, stopNone, parserErrWrap(input, open, err)
			}
			if name == "" {
				return nil, 0, stopNone, parserErr(input, open, "parser: empty block name")
			}
			body, elseBody, next, err := parseBlock(input, i, name)
			if err != nil {
				return nil, 0, stopNone, err
			}
			node, err := buildBlockNode(input, open, name, argsStr, blockParams, body, elseBody)
			if err != nil {
				return nil, 0, stopNone, parserErrWrap(input, open, err)
			}
			nodes = append(nodes, node)
			i = next
			continue
		}
		if strings.HasPrefix(content, ">") {
			rest := strings.TrimSpace(content[1:])
			if rest == "" {
				return nil, 0, stopNone, parserErr(input, open, "parser: empty partial name")
			}
			name, ctx, hash, err := ast.ParsePartialArgs(rest)
			if err != nil {
				return nil, 0, stopNone, parserErrWrap(input, open, err)
			}
			nodes = append(nodes, &ast.Partial{Name: name, Ctx: ctx, Hash: hash})
			continue
		}
		expr, err := ast.ParseExpr(content)
		if err != nil {
			return nil, 0, stopNone, parserErrWrap(input, open, err)
		}
		nodes = append(nodes, &ast.Mustache{Expr: expr, Raw: raw})
	}
	if endBlock != "" {
		return nil, 0, stopNone, parserErr(input, i, fmt.Sprintf("parser: unclosed block %q", endBlock))
	}
	return nodes, i, stopNone, nil
}

// buildBlockNode creates the appropriate typed block node from the parsed components.
func buildBlockNode(input string, offset int, name, argsStr string, blockParams []string, body, elseBody []ast.Node) (ast.Node, error) {
	switch name {
	case "if", "unless":
		if argsStr == "" {
			return nil, hexerr.New(fmt.Sprintf("parser: %s requires a condition", name))
		}
		test, blockHash, err := ast.ParseBlockHeader(argsStr)
		if err != nil {
			return nil, err
		}
		if test == nil {
			return nil, hexerr.New(fmt.Sprintf("parser: %s requires a condition expression", name))
		}
		return &ast.IfBlock{
			Unless: name == "unless",
			Test:   test,
			Hash:   blockHash,
			Body:   body,
			Else:   elseBody,
		}, nil

	case "with":
		if argsStr == "" {
			return nil, hexerr.New("parser: with requires a value expression")
		}
		val, err := ast.ParseExpr(argsStr)
		if err != nil {
			return nil, err
		}
		return &ast.WithBlock{
			Value:       val,
			BlockParams: blockParams,
			Body:        body,
			Else:        elseBody,
		}, nil

	case "each":
		col := argsStr
		// Normalise "each in collection" → collection = tail after "in "
		if strings.HasPrefix(col, "in ") {
			col = strings.TrimSpace(col[3:])
		}
		if col == "" {
			return nil, hexerr.New("parser: each requires a collection expression")
		}
		collection, err := ast.ParseExpr(col)
		if err != nil {
			return nil, err
		}
		return &ast.EachBlock{
			Collection:  collection,
			BlockParams: blockParams,
			Body:        body,
			Else:        elseBody,
		}, nil

	case "block":
		// {{#block "name"}} or {{#block name}}
		slotName, err := parseSlotName(argsStr)
		if err != nil {
			return nil, err
		}
		return &ast.LayoutBlock{Name: slotName, Body: body}, nil

	case "partial":
		// {{#partial "name"}}
		slotName, err := parseSlotName(argsStr)
		if err != nil {
			return nil, err
		}
		return &ast.LayoutPartial{Name: slotName, Body: body}, nil

	default:
		// Generic block: custom helper or universal section.
		var params []ast.Expr
		var hash []ast.HashEntry
		if argsStr != "" {
			e, err := ast.ParseExpr(argsStr)
			if err != nil {
				return nil, err
			}
			// If the full argsStr parsed into a HelperCall, that means there were multiple
			// tokens; spread them into Params and Hash of the Block.
			// If it's a single Expr, it's the first (and only) positional param.
			if hc, ok := e.(*ast.HelperCall); ok {
				// The HelperCall here has the block's first arg as Callee (misparse);
				// rebuild as params = [PathRef(callee)] + hc.Params, hash = hc.Hash.
				params = append([]ast.Expr{&ast.PathRef{Path: hc.Callee}}, hc.Params...)
				hash = hc.Hash
			} else {
				params = []ast.Expr{e}
			}
		}
		return &ast.Block{
			Name:        name,
			Params:      params,
			Hash:        hash,
			BlockParams: blockParams,
			Body:        body,
			Else:        elseBody,
		}, nil
	}
}

// parseSlotName extracts a slot/block name from a string like "content" or "\"main\"".
func parseSlotName(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", hexerr.New("parser: missing block/partial name")
	}
	// If quoted, unwrap the string literal.
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		expr, err := ast.ParseExpr(s)
		if err != nil {
			return "", err
		}
		if sl, ok := expr.(*ast.StringLit); ok {
			return sl.Value, nil
		}
		return "", hexerr.New("parser: block/partial name must be a string or identifier")
	}
	return s, nil
}

func parseBlock(input string, start int, name string) ([]ast.Node, []ast.Node, int, error) {
	body, next, stop, err := parseUntilStop(input, start, name)
	if err != nil {
		return nil, nil, 0, err
	}
	if stop == stopElse {
		elseBody, next, stop, err := parseElseBranch(input, next, name)
		if err != nil {
			return nil, nil, 0, err
		}
		if stop != stopEnd {
			return nil, nil, 0, parserErr(input, next, fmt.Sprintf("parser: unclosed block %q", name))
		}
		return body, elseBody, next, nil
	}
	if stop != stopEnd {
		return nil, nil, 0, parserErr(input, next, fmt.Sprintf("parser: unclosed block %q", name))
	}
	return body, nil, next, nil
}

func parseElseBranch(input string, start int, endBlock string) ([]ast.Node, int, stopKind, error) {
	var nodes []ast.Node
	i := start
	for i < len(input) {
		open := strings.Index(input[i:], "{{")
		if open < 0 {
			if i < len(input) {
				nodes = append(nodes, &ast.Text{Value: input[i:]})
			}
			return nodes, len(input), stopNone, parserErr(input, i, fmt.Sprintf("parser: unclosed else branch in block %q", endBlock))
		}
		open += i
		if open > i {
			nodes = append(nodes, &ast.Text{Value: input[i:open]})
		}

		contentStart := open + 2
		if contentStart >= len(input) {
			return nil, 0, stopNone, parserErr(input, open, "parser: unclosed mustache")
		}

		// Check for else if shorthand
		for _, prefix := range []string{"else if ", "elseif "} {
			if strings.HasPrefix(input[contentStart:], prefix) {
				condStart := contentStart + len(prefix)
				end := strings.Index(input[condStart:], "}}")
				if end < 0 {
					return nil, 0, stopNone, parserErr(input, condStart, "parser: unclosed else if")
				}
				cond := strings.TrimSpace(input[condStart : condStart+end])
				blockStart := condStart + end + 2
				test, err := ast.ParseExpr(cond)
				if err != nil {
					return nil, 0, stopNone, parserErrWrap(input, condStart, err)
				}
				ifBody, ifNext, ifStop, err := parseUntilStop(input, blockStart, "if")
				if err != nil {
					return nil, 0, stopNone, err
				}
				if ifStop == stopElse {
					ifElseBody, ifNext2, ifStop2, err := parseElseBranch(input, ifNext, "if")
					if err != nil {
						return nil, 0, stopNone, err
					}
					if ifStop2 != stopEnd {
						return nil, 0, stopNone, parserErr(input, ifNext2, "parser: unclosed else if block")
					}
					nodes = append(nodes, &ast.IfBlock{Test: test, Body: ifBody, Else: ifElseBody})
					i = ifNext2
				} else {
					if ifStop != stopEnd {
						return nil, 0, stopNone, parserErr(input, ifNext, "parser: unclosed else if block")
					}
					nodes = append(nodes, &ast.IfBlock{Test: test, Body: ifBody})
					i = ifNext
				}
				goto nextIter
			}
		}

		// Regular parsing
		{
			rest, next, stop, err := parseUntilStop(input, open, endBlock)
			if err != nil {
				return nil, 0, stopNone, err
			}
			nodes = append(nodes, rest...)
			if stop == stopEnd {
				return nodes, next, stopEnd, nil
			}
			if stop == stopElse {
				return nil, 0, stopNone, parserErr(input, open, "parser: unexpected else in else branch")
			}
			i = next
		}

	nextIter:
	}
	return nodes, i, stopNone, parserErr(input, i, fmt.Sprintf("parser: unclosed else branch in block %q", endBlock))
}

func splitBlockStart(expr string) (string, string, []string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", "", nil, nil
	}
	name, rest := splitNameArgs(expr)
	if name == "" {
		return "", "", nil, nil
	}
	args, params, err := extractBlockParams(rest)
	if err != nil {
		return "", "", nil, err
	}
	return name, args, params, nil
}

func splitNameArgs(expr string) (string, string) {
	for i := 0; i < len(expr); i++ {
		if isSpace(expr[i]) {
			return expr[:i], strings.TrimSpace(expr[i:])
		}
	}
	return expr, ""
}

func extractBlockParams(expr string) (string, []string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", nil, nil
	}
	pipeStart, pipeEnd := findPipePair(expr)
	if pipeStart < 0 || pipeEnd <= pipeStart {
		return expr, nil, nil
	}
	before := strings.TrimSpace(expr[:pipeStart])
	asIdx := lastAsTokenIndex(before)
	if asIdx < 0 {
		return expr, nil, nil
	}
	paramsPart := strings.TrimSpace(expr[pipeStart+1 : pipeEnd])
	if paramsPart == "" {
		return "", nil, hexerr.New("parser: empty block params")
	}
	params := strings.Fields(paramsPart)
	if len(params) == 0 {
		return "", nil, hexerr.New("parser: empty block params")
	}
	args := strings.TrimSpace(before[:asIdx])
	return args, params, nil
}

func findPipePair(expr string) (int, int) {
	inQuote := byte(0)
	start := -1
	for i := 0; i < len(expr); i++ {
		ch := expr[i]
		if inQuote != 0 {
			if ch == '\\' && i+1 < len(expr) {
				i++
				continue
			}
			if ch == inQuote {
				inQuote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if ch == '|' {
			if start == -1 {
				start = i
			} else {
				return start, i
			}
		}
	}
	return -1, -1
}

func lastAsTokenIndex(expr string) int {
	i := len(expr)
	for i > 0 && isSpace(expr[i-1]) {
		i--
	}
	end := i
	for i > 0 && !isSpace(expr[i-1]) {
		i--
	}
	if end-i != 2 || expr[i:end] != "as" {
		return -1
	}
	return i
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func skipWhitespace(input string, start int) int {
	for start < len(input) && isSpace(input[start]) {
		start++
	}
	return start
}

func trimRightText(nodes *[]ast.Node) {
	if len(*nodes) == 0 {
		return
	}
	last := (*nodes)[len(*nodes)-1]
	text, ok := last.(*ast.Text)
	if !ok {
		return
	}
	text.Value = strings.TrimRightFunc(text.Value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	if text.Value == "" {
		*nodes = (*nodes)[:len(*nodes)-1]
	}
}

func parseRawBlock(input string, open int, nodes *[]ast.Node) (int, error) {
	start := open + len("{{{{")
	trimLeft := false
	if start < len(input) && input[start] == '~' {
		trimLeft = true
		start++
	}
	end := strings.Index(input[start:], "}}}}")
	if end < 0 {
		return 0, parserErr(input, open, "parser: unclosed raw block")
	}
	name := strings.TrimSpace(input[start : start+end])
	if name == "" {
		return 0, parserErr(input, start, "parser: empty raw block name")
	}
	if trimLeft {
		trimRightText(nodes)
	}
	bodyStart := start + end + len("}}}}")
	closeStart := strings.Index(input[bodyStart:], "{{{{/")
	if closeStart < 0 {
		return 0, parserErr(input, start, fmt.Sprintf("parser: unclosed raw block %q", name))
	}
	closeStart += bodyStart
	closeTagStart := closeStart + len("{{{{/")
	closeEnd := strings.Index(input[closeTagStart:], "}}}}")
	if closeEnd < 0 {
		return 0, parserErr(input, closeTagStart, fmt.Sprintf("parser: unclosed raw block %q", name))
	}
	closeContent := strings.TrimSpace(input[closeTagStart : closeTagStart+closeEnd])
	trimRight := false
	if strings.HasSuffix(closeContent, "~") {
		trimRight = true
		closeContent = strings.TrimSpace(strings.TrimSuffix(closeContent, "~"))
	}
	if closeContent != name {
		return 0, parserErr(input, closeTagStart, fmt.Sprintf("parser: expected /%s, got /%s", name, closeContent))
	}
	if closeStart > bodyStart {
		*nodes = append(*nodes, &ast.Text{Value: input[bodyStart:closeStart]})
	}
	next := closeTagStart + closeEnd + len("}}}}")
	if trimRight {
		next = skipWhitespace(input, next)
	}
	return next, nil
}

func stopLabel(stop stopKind, endBlock string) string {
	switch stop {
	case stopElse:
		return "else"
	case stopEnd:
		if endBlock == "" {
			return "block end"
		}
		return fmt.Sprintf("closing block %q", endBlock)
	default:
		return "delimiter"
	}
}
