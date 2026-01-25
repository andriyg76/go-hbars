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
	assertPartial(t, nodes[6], "head", "user")
	assertText(t, nodes[7], ".")
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

func assertPartial(t *testing.T, node ast.Node, name string, ctx string) {
	t.Helper()
	p, ok := node.(*ast.Partial)
	if !ok {
		t.Fatalf("expected Partial node, got %T", node)
	}
	if p.Name != name || p.ContextExpr != ctx {
		t.Fatalf("Partial = (%q, %q)", p.Name, p.ContextExpr)
	}
}
