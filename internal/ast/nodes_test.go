package ast

import (
	"testing"
)

// Ensure all node types implement Node.
var (
	_ Node = (*Text)(nil)
	_ Node = (*Mustache)(nil)
	_ Node = (*Partial)(nil)
	_ Node = (*Block)(nil)
)

func TestText_Node(t *testing.T) {
	n := &Text{Value: "hello"}
	if n.Value != "hello" {
		t.Errorf("Text.Value = %q, want hello", n.Value)
	}
	_ = Node(n)
}

func TestMustache_Node(t *testing.T) {
	n := &Mustache{Expr: "name", Raw: false}
	if n.Expr != "name" || n.Raw {
		t.Errorf("Mustache = %+v", n)
	}
	n2 := &Mustache{Expr: "raw", Raw: true}
	if !n2.Raw {
		t.Error("Mustache.Raw want true")
	}
	_ = Node(n)
	_ = Node(n2)
}

func TestPartial_Node(t *testing.T) {
	n := &Partial{Expr: "header"}
	if n.Expr != "header" {
		t.Errorf("Partial.Expr = %q, want header", n.Expr)
	}
	_ = Node(n)
}

func TestBlock_Node(t *testing.T) {
	n := &Block{
		Name:   "if",
		Args:   "ok",
		Params: nil,
		Body:   []Node{&Text{Value: "yes"}},
		Else:   []Node{&Text{Value: "no"}},
	}
	if n.Name != "if" || n.Args != "ok" {
		t.Errorf("Block = %+v", n)
	}
	if len(n.Body) != 1 || len(n.Else) != 1 {
		t.Errorf("Block body/else len = %d/%d", len(n.Body), len(n.Else))
	}
	n2 := &Block{Name: "each", Args: "items", Params: []string{"item", "idx"}}
	if len(n2.Params) != 2 {
		t.Errorf("Block.Params len = %d", len(n2.Params))
	}
	_ = Node(n)
	_ = Node(n2)
}
