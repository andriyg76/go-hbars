package parser

import (
	"fmt"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
)

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
		return nil, 0, fmt.Errorf("parser: unexpected %s", stopLabel(stop, endBlock))
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
				return nil, 0, stopNone, fmt.Errorf("parser: unclosed block %q", endBlock)
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
				return nil, 0, stopNone, fmt.Errorf("parser: unclosed comment")
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
			return nil, 0, stopNone, fmt.Errorf("parser: unclosed mustache")
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
				return nil, 0, stopNone, fmt.Errorf("parser: unexpected else")
			}
			return nodes, i, stopElse, nil
		}
		if strings.HasPrefix(content, "/") {
			name := strings.TrimSpace(content[1:])
			if name == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: empty block name")
			}
			if endBlock == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: unexpected closing block %q", name)
			}
			if name != endBlock {
				return nil, 0, stopNone, fmt.Errorf("parser: expected /%s, got /%s", endBlock, name)
			}
			return nodes, i, stopEnd, nil
		}
		if strings.HasPrefix(content, "#") {
			name, args, params, err := splitBlockStart(content[1:])
			if err != nil {
				return nil, 0, stopNone, err
			}
			if name == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: empty block name")
			}
			body, elseBody, next, err := parseBlock(input, i, name)
			if err != nil {
				return nil, 0, stopNone, err
			}
			nodes = append(nodes, &ast.Block{
				Name:   name,
				Args:   args,
				Params: params,
				Body:   body,
				Else:   elseBody,
			})
			i = next
			continue
		}
		if strings.HasPrefix(content, ">") {
			rest := strings.TrimSpace(content[1:])
			if rest == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: empty partial name")
			}
			nodes = append(nodes, &ast.Partial{Expr: rest})
			continue
		}
		nodes = append(nodes, &ast.Mustache{Expr: content, Raw: raw})
	}
	if endBlock != "" {
		return nil, 0, stopNone, fmt.Errorf("parser: unclosed block %q", endBlock)
	}
	return nodes, i, stopNone, nil
}

func parseBlock(input string, start int, name string) ([]ast.Node, []ast.Node, int, error) {
	body, next, stop, err := parseUntilStop(input, start, name)
	if err != nil {
		return nil, nil, 0, err
	}
	if stop == stopElse {
		elseBody, next, stop, err := parseUntilStop(input, next, name)
		if err != nil {
			return nil, nil, 0, err
		}
		if stop != stopEnd {
			return nil, nil, 0, fmt.Errorf("parser: unclosed block %q", name)
		}
		return body, elseBody, next, nil
	}
	if stop != stopEnd {
		return nil, nil, 0, fmt.Errorf("parser: unclosed block %q", name)
	}
	return body, nil, next, nil
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
		return "", nil, fmt.Errorf("parser: empty block params")
	}
	params := strings.Fields(paramsPart)
	if len(params) == 0 {
		return "", nil, fmt.Errorf("parser: empty block params")
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
		return 0, fmt.Errorf("parser: unclosed raw block")
	}
	name := strings.TrimSpace(input[start : start+end])
	if name == "" {
		return 0, fmt.Errorf("parser: empty raw block name")
	}
	if trimLeft {
		trimRightText(nodes)
	}
	bodyStart := start + end + len("}}}}")
	closeStart := strings.Index(input[bodyStart:], "{{{{/")
	if closeStart < 0 {
		return 0, fmt.Errorf("parser: unclosed raw block %q", name)
	}
	closeStart += bodyStart
	closeTagStart := closeStart + len("{{{{/")
	closeEnd := strings.Index(input[closeTagStart:], "}}}}")
	if closeEnd < 0 {
		return 0, fmt.Errorf("parser: unclosed raw block %q", name)
	}
	closeContent := strings.TrimSpace(input[closeTagStart : closeTagStart+closeEnd])
	trimRight := false
	if strings.HasSuffix(closeContent, "~") {
		trimRight = true
		closeContent = strings.TrimSpace(strings.TrimSuffix(closeContent, "~"))
	}
	if closeContent != name {
		return 0, fmt.Errorf("parser: expected /%s, got /%s", name, closeContent)
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
