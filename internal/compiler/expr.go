package compiler

import (
	"fmt"
	"strings"

	"github.com/andriyg76/go-hbars/internal/ast"
	"github.com/andriyg76/hexerr"
)

// astExprToExpr converts an ast.Expr (from the new typed AST) to the internal expr type.
func astExprToExpr(e ast.Expr) expr {
	if e == nil {
		return expr{kind: exprNull}
	}
	switch v := e.(type) {
	case *ast.PathRef:
		return expr{kind: exprPath, value: v.Path}
	case *ast.StringLit:
		return expr{kind: exprString, value: v.Value}
	case *ast.NumberLit:
		return expr{kind: exprNumber, value: v.Value}
	case *ast.BoolLit:
		if v.Value {
			return expr{kind: exprBool, value: "true"}
		}
		return expr{kind: exprBool, value: "false"}
	case *ast.NullLit:
		return expr{kind: exprNull}
	case *ast.HelperCall:
		args := make([]expr, len(v.Params))
		for i, p := range v.Params {
			args[i] = astExprToExpr(p)
		}
		hash := make([]hashArg, len(v.Hash))
		for i, h := range v.Hash {
			hash[i] = hashArg{key: h.Key, value: astExprToExpr(h.Value)}
		}
		return expr{kind: exprCall, name: v.Callee, args: args, hash: hash}
	default:
		return expr{kind: exprNull}
	}
}

// astHashToHashArgs converts []ast.HashEntry to []hashArg for the internal expr system.
func astHashToHashArgs(h []ast.HashEntry) []hashArg {
	if len(h) == 0 {
		return nil
	}
	out := make([]hashArg, len(h))
	for i, entry := range h {
		out[i] = hashArg{key: entry.Key, value: astExprToExpr(entry.Value)}
	}
	return out
}

// pathsFromAstExpr returns all path-like strings used in an ast.Expr.
func pathsFromAstExpr(e ast.Expr) []string {
	if e == nil {
		return nil
	}
	switch v := e.(type) {
	case *ast.PathRef:
		return []string{v.Path}
	case *ast.HelperCall:
		var out []string
		for _, p := range v.Params {
			out = append(out, pathsFromAstExpr(p)...)
		}
		for _, h := range v.Hash {
			out = append(out, pathsFromAstExpr(h.Value)...)
		}
		return out
	default:
		return nil
	}
}

type exprKind int

const (
	exprPath exprKind = iota
	exprString
	exprNumber
	exprBool
	exprNull
	exprCall
)

type expr struct {
	kind  exprKind
	value string
	name  string
	args  []expr
	hash  []hashArg
}

type hashArg struct {
	key   string
	value expr
}

func parseParts(input string) ([]expr, []hashArg, error) {
	tokens, err := tokenizeExpr(input)
	if err != nil {
		return nil, nil, err
	}
	p := exprParser{tokens: tokens}
	parts, hash, err := p.parseParts(false)
	if err != nil {
		return nil, nil, err
	}
	if p.hasNext() {
		return nil, nil, hexerr.New(fmt.Sprintf("unexpected token %q", p.peek().value))
	}
	return parts, hash, nil
}

type exprParser struct {
	tokens []token
	pos    int
}

func (p *exprParser) hasNext() bool {
	return p.pos < len(p.tokens)
}

func (p *exprParser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	return p.tokens[p.pos]
}

func (p *exprParser) peekNext() token {
	if p.pos+1 >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	return p.tokens[p.pos+1]
}

func (p *exprParser) next() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *exprParser) parseParts(stopAtRParen bool) ([]expr, []hashArg, error) {
	var parts []expr
	var hash []hashArg
	for p.hasNext() {
		if p.peek().typ == tokRParen {
			if stopAtRParen {
				return parts, hash, nil
			}
			return nil, nil, hexerr.New("unexpected )")
		}
		if p.peek().typ == tokEquals {
			return nil, nil, hexerr.New("unexpected =")
		}
		if p.peek().typ == tokWord && p.peekNext().typ == tokEquals {
			key := p.next().value
			p.next()
			if key == "" {
				return nil, nil, hexerr.New("empty hash key")
			}
			value, err := p.parseExpr()
			if err != nil {
				return nil, nil, err
			}
			hash = append(hash, hashArg{key: key, value: value})
			continue
		}
		part, err := p.parseExpr()
		if err != nil {
			return nil, nil, err
		}
		parts = append(parts, part)
	}
	if stopAtRParen {
		return nil, nil, hexerr.New("missing )")
	}
	return parts, hash, nil
}

func (p *exprParser) parseExpr() (expr, error) {
	tok := p.next()
	switch tok.typ {
	case tokWord:
		return classifyWord(tok.value), nil
	case tokString:
		return expr{kind: exprString, value: tok.value}, nil
	case tokLParen:
		return p.parseSubexpr()
	case tokRParen:
		return expr{}, hexerr.New("unexpected )")
	case tokEquals:
		return expr{}, hexerr.New("unexpected =")
	case tokEOF:
		return expr{}, hexerr.New("unexpected end of expression")
	default:
		return expr{}, hexerr.New("unexpected token")
	}
}

func (p *exprParser) parseSubexpr() (expr, error) {
	parts, hash, err := p.parseParts(true)
	if err != nil {
		return expr{}, err
	}
	if !p.hasNext() || p.peek().typ != tokRParen {
		return expr{}, hexerr.New("missing )")
	}
	p.next()
	if len(parts) == 0 {
		return expr{}, hexerr.New("empty subexpression")
	}
	if len(parts) == 1 && len(hash) == 0 {
		return parts[0], nil
	}
	if parts[0].kind != exprPath {
		return expr{}, hexerr.New("subexpression must start with a helper name")
	}
	return expr{
		kind: exprCall,
		name: parts[0].value,
		args: parts[1:],
		hash: hash,
	}, nil
}

func classifyWord(value string) expr {
	lower := strings.ToLower(value)
	switch lower {
	case "true", "false":
		return expr{kind: exprBool, value: lower}
	case "null", "nil":
		return expr{kind: exprNull}
	default:
		if isNumber(value) {
			return expr{kind: exprNumber, value: value}
		}
		return expr{kind: exprPath, value: value}
	}
}

type tokenType int

const (
	tokEOF tokenType = iota
	tokWord
	tokString
	tokLParen
	tokRParen
	tokEquals
)

type token struct {
	typ   tokenType
	value string
}

func tokenizeExpr(input string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		for i < len(input) && isSpace(input[i]) {
			i++
		}
		if i >= len(input) {
			break
		}
		switch input[i] {
		case '(':
			tokens = append(tokens, token{typ: tokLParen, value: "("})
			i++
		case ')':
			tokens = append(tokens, token{typ: tokRParen, value: ")"})
			i++
		case '=':
			tokens = append(tokens, token{typ: tokEquals, value: "="})
			i++
		case '"', '\'':
			quote := input[i]
			i++
			var sb strings.Builder
			closed := false
			for i < len(input) {
				ch := input[i]
				if ch == '\\' && i+1 < len(input) {
					next := input[i+1]
					if next == quote || next == '\\' {
						sb.WriteByte(next)
						i += 2
						continue
					}
				}
				if ch == quote {
					i++
					closed = true
					break
				}
				sb.WriteByte(ch)
				i++
			}
			if !closed {
				return nil, hexerr.New("unclosed string literal")
			}
			tokens = append(tokens, token{typ: tokString, value: sb.String()})
		default:
			start := i
			for i < len(input) && !isSpace(input[i]) && input[i] != '(' && input[i] != ')' && input[i] != '=' {
				i++
			}
			tokens = append(tokens, token{typ: tokWord, value: input[start:i]})
		}
	}
	return tokens, nil
}
