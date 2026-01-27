package compiler

import (
	"fmt"
	"strings"
)

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
		return nil, nil, fmt.Errorf("unexpected token %q", p.peek().value)
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
			return nil, nil, fmt.Errorf("unexpected )")
		}
		if p.peek().typ == tokEquals {
			return nil, nil, fmt.Errorf("unexpected =")
		}
		if p.peek().typ == tokWord && p.peekNext().typ == tokEquals {
			key := p.next().value
			p.next()
			if key == "" {
				return nil, nil, fmt.Errorf("empty hash key")
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
		return nil, nil, fmt.Errorf("missing )")
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
		return expr{}, fmt.Errorf("unexpected )")
	case tokEquals:
		return expr{}, fmt.Errorf("unexpected =")
	case tokEOF:
		return expr{}, fmt.Errorf("unexpected end of expression")
	default:
		return expr{}, fmt.Errorf("unexpected token")
	}
}

func (p *exprParser) parseSubexpr() (expr, error) {
	parts, hash, err := p.parseParts(true)
	if err != nil {
		return expr{}, err
	}
	if !p.hasNext() || p.peek().typ != tokRParen {
		return expr{}, fmt.Errorf("missing )")
	}
	p.next()
	if len(parts) == 0 {
		return expr{}, fmt.Errorf("empty subexpression")
	}
	if len(parts) == 1 && len(hash) == 0 {
		return parts[0], nil
	}
	if parts[0].kind != exprPath {
		return expr{}, fmt.Errorf("subexpression must start with a helper name")
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
				return nil, fmt.Errorf("unclosed string literal")
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
