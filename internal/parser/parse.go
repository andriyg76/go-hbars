package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
)

// Parse turns a template string into a list of nodes.
func Parse(input string) ([]ast.Node, error) {
	var nodes []ast.Node
	i := 0
	for i < len(input) {
		open := strings.Index(input[i:], "{{")
		if open < 0 {
			if i < len(input) {
				nodes = append(nodes, &ast.Text{Value: input[i:]})
			}
			break
		}
		open += i
		if open > i {
			nodes = append(nodes, &ast.Text{Value: input[i:open]})
		}

		if strings.HasPrefix(input[open:], "{{!--") {
			end := strings.Index(input[open+4:], "--}}")
			if end < 0 {
				return nil, fmt.Errorf("parser: unclosed comment")
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
			return nil, fmt.Errorf("parser: unclosed mustache")
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
		if strings.HasPrefix(content, ">") {
			rest := strings.TrimSpace(content[1:])
			if rest == "" {
				return nil, fmt.Errorf("parser: empty partial name")
			}
			name, ctxExpr := splitPartial(rest)
			nodes = append(nodes, &ast.Partial{Name: name, ContextExpr: ctxExpr})
			continue
		}
		nodes = append(nodes, &ast.Mustache{Expr: content, Raw: raw})
	}
	return nodes, nil
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
