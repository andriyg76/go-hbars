package ast

// Node is a template AST node.
type Node interface {
	node()
}

// Text is a raw text node.
type Text struct {
	Value string
}

func (*Text) node() {}

// Mustache is a simple mustache expression.
type Mustache struct {
	Expr string
	Raw  bool
}

func (*Mustache) node() {}

// Partial is a partial invocation.
type Partial struct {
	Expr string
}

func (*Partial) node() {}

// Block is a block helper invocation with an optional else branch.
type Block struct {
	Name   string
	Args   string
	Params []string
	Body   []Node
	Else   []Node
}

func (*Block) node() {}
