package parser

import (
	"testing"

	"github.com/andriyg76/go-hbars/internal/ast"
)

func TestParseMixed(t *testing.T) {
	input := "Hi {{name}}!{{{raw}}} {{& title}}{{!ignore}}{{!--block--}}{{> \"head\" user}}."
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 8 {
		t.Fatalf("expected 8 nodes, got %d", len(nodes))
	}

	assertText(t, nodes[0], "Hi ")
	assertMustache(t, nodes[1], "name", false)
	assertText(t, nodes[2], "!")
	assertMustache(t, nodes[3], "raw", true)
	assertText(t, nodes[4], " ")
	assertMustache(t, nodes[5], "title", true)
	assertPartial(t, nodes[6], "\"head\" user")
	assertText(t, nodes[7], ".")
}

func TestParseBlockIfElse(t *testing.T) {
	input := "{{#if ok}}Yes{{else}}No{{/if}}"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	block := assertBlock(t, nodes[0], "if", "ok", nil)
	if len(block.Body) != 1 || len(block.Else) != 1 {
		t.Fatalf("expected body/else length 1, got %d/%d", len(block.Body), len(block.Else))
	}
	assertText(t, block.Body[0], "Yes")
	assertText(t, block.Else[0], "No")
}

func TestParseNestedBlocks(t *testing.T) {
	input := "{{#each items}}{{#with user}}{{name}}{{/with}}{{/each}}"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	each := assertBlock(t, nodes[0], "each", "items", nil)
	if len(each.Body) != 1 {
		t.Fatalf("expected each body length 1, got %d", len(each.Body))
	}
	with := assertBlock(t, each.Body[0], "with", "user", nil)
	if len(with.Body) != 1 {
		t.Fatalf("expected with body length 1, got %d", len(with.Body))
	}
	assertMustache(t, with.Body[0], "name", false)
}

func TestParseBlockParams(t *testing.T) {
	input := "{{#each items as |item idx|}}{{item}}{{/each}}"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	block := assertBlock(t, nodes[0], "each", "items", []string{"item", "idx"})
	if len(block.Body) != 1 {
		t.Fatalf("expected body length 1, got %d", len(block.Body))
	}
	assertMustache(t, block.Body[0], "item", false)
}

func TestParseWhitespaceTrim(t *testing.T) {
	input := "a {{~name}} b {{name~}} c"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 5 {
		t.Fatalf("expected 5 nodes, got %d", len(nodes))
	}
	assertText(t, nodes[0], "a")
	assertMustache(t, nodes[1], "name", false)
	assertText(t, nodes[2], " b ")
	assertMustache(t, nodes[3], "name", false)
	assertText(t, nodes[4], "c")
}

func TestParseRawBlock(t *testing.T) {
	input := "Hi {{{{raw}}}} {{name}} {{{{/raw}}}}!"
	nodes, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	assertText(t, nodes[0], "Hi ")
	assertText(t, nodes[1], " {{name}} ")
	assertText(t, nodes[2], "!")
}

func TestParseErrors(t *testing.T) {
	if _, err := Parse("{{!--"); err == nil {
		t.Fatalf("expected unclosed comment error")
	}
	if _, err := Parse("{{name"); err == nil {
		t.Fatalf("expected unclosed mustache error")
	}
	if _, err := Parse("{{> }}"); err == nil {
		t.Fatalf("expected empty partial name error")
	}
	if _, err := Parse("{{else}}"); err == nil {
		t.Fatalf("expected unexpected else error")
	}
	if _, err := Parse("{{/if}}"); err == nil {
		t.Fatalf("expected unexpected closing block error")
	}
	if _, err := Parse("{{#if ok}}"); err == nil {
		t.Fatalf("expected unclosed block error")
	}
	if _, err := Parse("{{#if ok}}{{/each}}"); err == nil {
		t.Fatalf("expected mismatched block error")
	}
	if _, err := Parse("{{{{raw}}}}"); err == nil {
		t.Fatalf("expected unclosed raw block error")
	}
	if _, err := Parse("{{#each items as ||}}{{/each}}"); err == nil {
		t.Fatalf("expected invalid block params error")
	}
}

func assertText(t *testing.T, node ast.Node, value string) {
	t.Helper()
	text, ok := node.(*ast.Text)
	if !ok {
		t.Fatalf("expected Text node, got %T", node)
	}
	if text.Value != value {
		t.Fatalf("Text value = %q", text.Value)
	}
}

func assertMustache(t *testing.T, node ast.Node, expr string, raw bool) {
	t.Helper()
	m, ok := node.(*ast.Mustache)
	if !ok {
		t.Fatalf("expected Mustache node, got %T", node)
	}
	if m.Expr != expr || m.Raw != raw {
		t.Fatalf("Mustache = (%q, %v)", m.Expr, m.Raw)
	}
}

func assertPartial(t *testing.T, node ast.Node, expr string) {
	t.Helper()
	p, ok := node.(*ast.Partial)
	if !ok {
		t.Fatalf("expected Partial node, got %T", node)
	}
	if p.Expr != expr {
		t.Fatalf("Partial = %q", p.Expr)
	}
}

func assertBlock(t *testing.T, node ast.Node, name string, args string, params []string) *ast.Block {
	t.Helper()
	b, ok := node.(*ast.Block)
	if !ok {
		t.Fatalf("expected Block node, got %T", node)
	}
	if b.Name != name || b.Args != args {
		t.Fatalf("Block = (%q, %q)", b.Name, b.Args)
	}
	if len(b.Params) != len(params) {
		t.Fatalf("Block params = %v", b.Params)
	}
	for i, param := range params {
		if b.Params[i] != param {
			t.Fatalf("Block params = %v", b.Params)
		}
	}
	return b
}
