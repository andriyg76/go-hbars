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

// PartialParam is one parameter to a partial: either a context path reference or hash key=value.
// Exactly one of Path or Hash is set. Params are merged in order into the partial's context map.
type PartialParam struct {
	Path string            // context reference (e.g. "_shared.menu", "user"); empty if Hash is set
	Hash map[string]string  // key=value pairs (value is expression string); nil if Path is set
}

// Partial is a partial invocation.
// NameOrExpr is the partial name or a subexpression (e.g. "(lookup . \"cardPartial\")") that evaluates at runtime.
// Params is the ordered list of context refs and/or hash; merged into a single context map for the partial.
type Partial struct {
	NameOrExpr string         // partial name or evaluable subexpression
	Params     []PartialParam // ordered: path refs and hash, merged into context
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
