package parser

import (
	"fmt"
	"strconv"
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

		if strings.HasPrefix(input[open:], "{{!--") {
			end := strings.Index(input[open+4:], "--}}")
			if end < 0 {
				return nil, 0, stopNone, fmt.Errorf("parser: unclosed comment")
			}
			i = open + 4 + end + len("--}}")
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
		end := strings.Index(input[open+startLen:], endDelim)
		if end < 0 {
			return nil, 0, stopNone, fmt.Errorf("parser: unclosed mustache")
		}
		content := strings.TrimSpace(input[open+startLen : open+startLen+end])
		i = open + startLen + end + len(endDelim)
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
			name, args := splitBlockStart(content[1:])
			if name == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: empty block name")
			}
			body, elseBody, next, err := parseBlock(input, i, name)
			if err != nil {
				return nil, 0, stopNone, err
			}
			nodes = append(nodes, &ast.Block{
				Name: name,
				Args: args,
				Body: body,
				Else: elseBody,
			})
			i = next
			continue
		}
		if strings.HasPrefix(content, ">") {
			rest := strings.TrimSpace(content[1:])
			if rest == "" {
				return nil, 0, stopNone, fmt.Errorf("parser: empty partial name")
			}
			name, ctxExpr := splitPartial(rest)
			nodes = append(nodes, &ast.Partial{Name: name, ContextExpr: ctxExpr})
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

func splitPartial(expr string) (string, string) {
	fields := strings.Fields(expr)
	if len(fields) == 0 {
		return "", ""
	}
	name := unquote(fields[0])
	if len(fields) == 1 {
		return name, ""
	}
	return name, strings.Join(fields[1:], " ")
}

func splitBlockStart(expr string) (string, string) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", ""
	}
	for i := 0; i < len(expr); i++ {
		if isSpace(expr[i]) {
			return expr[:i], strings.TrimSpace(expr[i:])
		}
	}
	return expr, ""
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}
	if value[0] != '"' && value[0] != '\'' {
		return value
	}
	unquoted, err := strconv.Unquote(value)
	if err != nil {
		return value
	}
	return unquoted
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
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
