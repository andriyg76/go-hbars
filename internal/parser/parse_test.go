package parser

import (
	"strings"
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
	assertPartialName(t, nodes[6], "head")
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
	block := assertIfBlock(t, nodes[0], "ok")
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
	each := assertEachBlock(t, nodes[0], "items", nil)
	if len(each.Body) != 1 {
		t.Fatalf("expected each body length 1, got %d", len(each.Body))
	}
	with := assertWithBlock(t, each.Body[0], "user", nil)
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
	block := assertEachBlock(t, nodes[0], "items", []string{"item", "idx"})
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

func TestParseErrorsIncludeLineNumber(t *testing.T) {
	// Parser errors should include " (line N)" for the position of the error.
	_, err := Parse("{{!--")
	if err == nil {
		t.Fatalf("expected unclosed comment error")
	}
	if !strings.Contains(err.Error(), " (line ") {
		t.Fatalf("expected error to include line number, got: %s", err.Error())
	}
	// Multi-line: error on line 3
	_, err = Parse("line1\nline2\n{{name")
	if err == nil {
		t.Fatalf("expected unclosed mustache error")
	}
	if !strings.Contains(err.Error(), " (line 3)") {
		t.Fatalf("expected error to include line 3, got: %s", err.Error())
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

func mustacheExprPath(m *ast.Mustache) string {
	if m == nil {
		return ""
	}
	if pr, ok := m.Expr.(*ast.PathRef); ok {
		return pr.Path
	}
	return ""
}

func assertMustache(t *testing.T, node ast.Node, path string, raw bool) {
	t.Helper()
	m, ok := node.(*ast.Mustache)
	if !ok {
		t.Fatalf("expected Mustache node, got %T", node)
	}
	if mustacheExprPath(m) != path || m.Raw != raw {
		t.Fatalf("Mustache = (expr=%v, raw=%v), want path=%q raw=%v", m.Expr, m.Raw, path, raw)
	}
}

func assertPartialName(t *testing.T, node ast.Node, wantName string) *ast.Partial {
	t.Helper()
	p, ok := node.(*ast.Partial)
	if !ok {
		t.Fatalf("expected Partial node, got %T", node)
	}
	var gotName string
	switch nm := p.Name.(type) {
	case *ast.StringLit:
		gotName = nm.Value
	case *ast.PathRef:
		gotName = nm.Path
	}
	if gotName != wantName {
		t.Fatalf("Partial name = %q, want %q", gotName, wantName)
	}
	return p
}

func assertIfBlock(t *testing.T, node ast.Node, testPath string) *ast.IfBlock {
	t.Helper()
	b, ok := node.(*ast.IfBlock)
	if !ok {
		t.Fatalf("expected IfBlock node, got %T", node)
	}
	pr, ok2 := b.Test.(*ast.PathRef)
	if !ok2 || pr.Path != testPath {
		t.Fatalf("IfBlock test = %v, want PathRef{%q}", b.Test, testPath)
	}
	return b
}

func assertWithBlock(t *testing.T, node ast.Node, valuePath string, blockParams []string) *ast.WithBlock {
	t.Helper()
	b, ok := node.(*ast.WithBlock)
	if !ok {
		t.Fatalf("expected WithBlock node, got %T", node)
	}
	pr, ok2 := b.Value.(*ast.PathRef)
	if !ok2 || pr.Path != valuePath {
		t.Fatalf("WithBlock value = %v, want PathRef{%q}", b.Value, valuePath)
	}
	if len(b.BlockParams) != len(blockParams) {
		t.Fatalf("WithBlock blockParams = %v, want %v", b.BlockParams, blockParams)
	}
	return b
}

func assertEachBlock(t *testing.T, node ast.Node, collectionPath string, blockParams []string) *ast.EachBlock {
	t.Helper()
	b, ok := node.(*ast.EachBlock)
	if !ok {
		t.Fatalf("expected EachBlock node, got %T", node)
	}
	pr, ok2 := b.Collection.(*ast.PathRef)
	if !ok2 || pr.Path != collectionPath {
		t.Fatalf("EachBlock collection = %v, want PathRef{%q}", b.Collection, collectionPath)
	}
	if len(b.BlockParams) != len(blockParams) {
		t.Fatalf("EachBlock blockParams = %v, want %v", b.BlockParams, blockParams)
	}
	for i, bp := range blockParams {
		if b.BlockParams[i] != bp {
			t.Fatalf("EachBlock blockParams[%d] = %q, want %q", i, b.BlockParams[i], bp)
		}
	}
	return b
}
