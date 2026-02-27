package ast

import (
	"fmt"
	"strings"

	"github.com/andriyg76/hexerr"
)

// ParseExpr parses a Handlebars expression content string (the text inside {{ }})
// into a single Expr plus any top-level hash entries.
//
//   "title"            → PathRef{Path:"title"}, nil
//   "upper name"       → HelperCall{Callee:"upper", Params:[PathRef{"name"}]}, nil
//   "title class=foo"  → HelperCall{Callee:"title", Hash:[{class,PathRef{foo}}]}, nil
//   "(upper name)"     → HelperCall{Callee:"upper", Params:[PathRef{"name"}]}, nil
func ParseExpr(input string) (Expr, error) {
	tokens, err := tokenizeExpr(input)
	if err != nil {
		return nil, err
	}
	p := &exprParser{tokens: tokens}
	parts, hash, err := p.parseParts(false)
	if err != nil {
		return nil, err
	}
	if p.hasNext() {
		return nil, hexerr.New(fmt.Sprintf("unexpected token %q", p.peek().value))
	}
	return assembleExpr(parts, hash)
}

// ParseBlockHeader parses a block header expression like "count includeZero=true" or "users as |item|".
// It returns the first positional expression as the main expression, and any hash entries as block-level options.
// Unlike ParseExpr, it does NOT assemble positional+hash into a HelperCall; the hash is returned separately.
// Only one positional arg is accepted (the test/value/collection); additional positional args are an error.
func ParseBlockHeader(input string) (mainExpr Expr, hash []HashEntry, err error) {
	tokens, err := tokenizeExpr(input)
	if err != nil {
		return nil, nil, err
	}
	p := &exprParser{tokens: tokens}
	parts, h, err := p.parseParts(false)
	if err != nil {
		return nil, nil, err
	}
	if len(parts) == 0 {
		return nil, h, nil
	}
	if len(parts) > 1 {
		return nil, nil, hexerr.New(fmt.Sprintf("block header: unexpected extra expression %q", parts[1]))
	}
	return parts[0], h, nil
}

// ParsePartialArgs parses the content of a partial invocation (after ">" and trimming)
// into name, optional context expression, and hash entries.
//
//   "menu"                → name=PathRef{"menu"}, ctx=nil, hash=nil
//   "menu _shared.menu"   → name=PathRef{"menu"}, ctx=PathRef{"_shared.menu"}, hash=nil
//   "\"head\" user k=v"   → name=StringLit{"head"}, ctx=PathRef{"user"}, hash=[{k,PathRef{v}}]
func ParsePartialArgs(input string) (name Expr, ctx Expr, hash []HashEntry, err error) {
	tokens, err := tokenizeExpr(input)
	if err != nil {
		return nil, nil, nil, err
	}
	p := &exprParser{tokens: tokens}
	parts, h, err := p.parseParts(false)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(parts) == 0 {
		return nil, nil, nil, hexerr.New("partial: empty name")
	}
	name = parts[0]
	if len(parts) > 1 {
		ctx = parts[1]
	}
	if len(parts) > 2 {
		return nil, nil, nil, hexerr.New("partial: context must be a single expression")
	}
	return name, ctx, h, nil
}

// assembleExpr combines parsed parts and hash into a single Expr.
func assembleExpr(parts []Expr, hash []HashEntry) (Expr, error) {
	if len(parts) == 0 {
		if len(hash) > 0 {
			return nil, hexerr.New("hash arguments without a callee")
		}
		return nil, hexerr.New("empty expression")
	}
	if len(parts) == 1 && len(hash) == 0 {
		return parts[0], nil
	}
	// Multi-part or has hash: build a HelperCall.
	callee, ok := calleeOf(parts[0])
	if !ok {
		return nil, hexerr.New("helper callee must be a path")
	}
	return &HelperCall{Callee: callee, Params: parts[1:], Hash: hash}, nil
}

// calleeOf extracts the callee name from an Expr (must be a PathRef).
func calleeOf(e Expr) (string, bool) {
	if p, ok := e.(*PathRef); ok {
		return p.Path, true
	}
	return "", false
}

// ---- tokenizer ----

type tokType int

const (
	tokEOF    tokType = iota
	tokWord           // identifier, path, number, keyword
	tokString         // quoted string
	tokLParen         // (
	tokRParen         // )
	tokEquals         // =
)

type token struct {
	typ   tokType
	value string
}

func tokenizeExpr(input string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		for i < len(input) && isExprSpace(input[i]) {
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
			for i < len(input) && !isExprSpace(input[i]) &&
				input[i] != '(' && input[i] != ')' && input[i] != '=' {
				i++
			}
			tokens = append(tokens, token{typ: tokWord, value: input[start:i]})
		}
	}
	return tokens, nil
}

func isExprSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// ---- parser ----

type exprParser struct {
	tokens []token
	pos    int
}

func (p *exprParser) hasNext() bool { return p.pos < len(p.tokens) }

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

// parseParts parses a sequence of expressions and hash entries.
// If stopAtRParen is true, parsing stops just before the closing ')'.
func (p *exprParser) parseParts(stopAtRParen bool) ([]Expr, []HashEntry, error) {
	var parts []Expr
	var hash []HashEntry
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
		// hash entry: word = value
		if p.peek().typ == tokWord && p.peekNext().typ == tokEquals {
			key := p.next().value
			p.next() // consume '='
			if key == "" {
				return nil, nil, hexerr.New("empty hash key")
			}
			val, err := p.parseAtom()
			if err != nil {
				return nil, nil, err
			}
			hash = append(hash, HashEntry{Key: key, Value: val})
			continue
		}
		part, err := p.parseAtom()
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

// parseAtom parses a single expression atom (word, string, subexpr).
func (p *exprParser) parseAtom() (Expr, error) {
	tok := p.next()
	switch tok.typ {
	case tokWord:
		return classifyWord(tok.value), nil
	case tokString:
		return &StringLit{Value: tok.value}, nil
	case tokLParen:
		return p.parseSubexpr()
	case tokRParen:
		return nil, hexerr.New("unexpected )")
	case tokEquals:
		return nil, hexerr.New("unexpected =")
	case tokEOF:
		return nil, hexerr.New("unexpected end of expression")
	default:
		return nil, hexerr.New("unexpected token")
	}
}

// parseSubexpr parses (helper arg1 arg2 key=val) — a subexpression.
func (p *exprParser) parseSubexpr() (Expr, error) {
	parts, hash, err := p.parseParts(true)
	if err != nil {
		return nil, err
	}
	if !p.hasNext() || p.peek().typ != tokRParen {
		return nil, hexerr.New("missing )")
	}
	p.next() // consume ')'
	if len(parts) == 0 {
		return nil, hexerr.New("empty subexpression")
	}
	if len(parts) == 1 && len(hash) == 0 {
		return parts[0], nil
	}
	callee, ok := calleeOf(parts[0])
	if !ok {
		return nil, hexerr.New("subexpression must start with a helper name")
	}
	return &HelperCall{Callee: callee, Params: parts[1:], Hash: hash}, nil
}

// classifyWord turns a word token into the appropriate Expr.
func classifyWord(value string) Expr {
	lower := strings.ToLower(value)
	switch lower {
	case "true":
		return &BoolLit{Value: true}
	case "false":
		return &BoolLit{Value: false}
	case "null", "nil":
		return &NullLit{}
	default:
		if isNumberStr(value) {
			return &NumberLit{Value: value}
		}
		return &PathRef{Path: value}
	}
}

func isNumberStr(s string) bool {
	if s == "" {
		return false
	}
	hasDot := false
	hasDigit := false
	for i, ch := range s {
		if ch == '-' && i == 0 {
			continue
		}
		if ch == '.' {
			if hasDot {
				return false
			}
			hasDot = true
			continue
		}
		if ch >= '0' && ch <= '9' {
			hasDigit = true
			continue
		}
		return false
	}
	return hasDigit
}
