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
	_ Node = (*IfBlock)(nil)
	_ Node = (*WithBlock)(nil)
	_ Node = (*EachBlock)(nil)
	_ Node = (*LayoutBlock)(nil)
	_ Node = (*LayoutPartial)(nil)
)

// Ensure all Expr types implement Expr.
var (
	_ Expr = (*PathRef)(nil)
	_ Expr = (*StringLit)(nil)
	_ Expr = (*NumberLit)(nil)
	_ Expr = (*BoolLit)(nil)
	_ Expr = (*NullLit)(nil)
	_ Expr = (*HelperCall)(nil)
)

func TestText_Node(t *testing.T) {
	n := &Text{Value: "hello"}
	if n.Value != "hello" {
		t.Errorf("Text.Value = %q, want hello", n.Value)
	}
	_ = Node(n)
}

func TestMustache_Node(t *testing.T) {
	n := &Mustache{Expr: &PathRef{Path: "name"}, Raw: false}
	pr, ok := n.Expr.(*PathRef)
	if !ok || pr.Path != "name" || n.Raw {
		t.Errorf("Mustache = %+v", n)
	}
	n2 := &Mustache{Expr: &PathRef{Path: "raw"}, Raw: true}
	if !n2.Raw {
		t.Error("Mustache.Raw want true")
	}
	_ = Node(n)
	_ = Node(n2)
}

func TestPartial_Node(t *testing.T) {
	n := &Partial{Name: &PathRef{Path: "header"}}
	pr, ok := n.Name.(*PathRef)
	if !ok || pr.Path != "header" {
		t.Errorf("Partial.Name = %+v, want PathRef{header}", n.Name)
	}
	_ = Node(n)
}

func TestBlock_Node(t *testing.T) {
	n := &Block{
		Name:   "customHelper",
		Params: []Expr{&PathRef{Path: "ok"}},
		Body:   []Node{&Text{Value: "yes"}},
		Else:   []Node{&Text{Value: "no"}},
	}
	if n.Name != "customHelper" {
		t.Errorf("Block.Name = %q, want customHelper", n.Name)
	}
	if len(n.Body) != 1 || len(n.Else) != 1 {
		t.Errorf("Block body/else len = %d/%d", len(n.Body), len(n.Else))
	}
	n2 := &Block{Name: "myHelper", BlockParams: []string{"item", "idx"}}
	if len(n2.BlockParams) != 2 {
		t.Errorf("Block.BlockParams len = %d", len(n2.BlockParams))
	}
	_ = Node(n)
	_ = Node(n2)
}

func TestIfBlock_Node(t *testing.T) {
	n := &IfBlock{
		Test: &PathRef{Path: "ok"},
		Body: []Node{&Text{Value: "yes"}},
		Else: []Node{&Text{Value: "no"}},
	}
	pr, ok := n.Test.(*PathRef)
	if !ok || pr.Path != "ok" {
		t.Errorf("IfBlock.Test = %+v", n.Test)
	}
	_ = Node(n)
}

func TestEachBlock_Node(t *testing.T) {
	n := &EachBlock{
		Collection:  &PathRef{Path: "items"},
		BlockParams: []string{"item", "idx"},
		Body:        []Node{&Text{Value: "row"}},
	}
	pr, ok := n.Collection.(*PathRef)
	if !ok || pr.Path != "items" {
		t.Errorf("EachBlock.Collection = %+v", n.Collection)
	}
	if len(n.BlockParams) != 2 {
		t.Errorf("EachBlock.BlockParams len = %d", len(n.BlockParams))
	}
	_ = Node(n)
}
