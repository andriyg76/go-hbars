package ast

// Node is a template AST node.
type Node interface {
	node()
}

// Expr is a structured expression: a path reference, a literal, or a helper call.
type Expr interface {
	expr()
}

// PathRef is a path expression: ".", "this", "user.name", "@root.title".
type PathRef struct {
	Path string
}

func (*PathRef) expr() {}

// StringLit is a string literal value, e.g. "hello".
type StringLit struct {
	Value string
}

func (*StringLit) expr() {}

// NumberLit is a numeric literal, e.g. 42 or 3.14.
// Value is stored as a string to preserve the exact source form.
type NumberLit struct {
	Value string
}

func (*NumberLit) expr() {}

// BoolLit is true or false.
type BoolLit struct {
	Value bool
}

func (*BoolLit) expr() {}

// NullLit represents null / nil.
type NullLit struct{}

func (*NullLit) expr() {}

// HashEntry is one key=value pair in a call expression or partial invocation.
type HashEntry struct {
	Key   string
	Value Expr
}

// HelperCall is a helper invocation: callee arg1 arg2 key=val.
// Subexpressions like (upper name) are also represented as HelperCall.
type HelperCall struct {
	Callee string
	Params []Expr
	Hash   []HashEntry
}

func (*HelperCall) expr() {}

// ---- Node types ----

// Text is a raw text node.
type Text struct {
	Value string
}

func (*Text) node() {}

// Mustache is a {{expr}} or {{{expr}}} interpolation.
type Mustache struct {
	Expr Expr
	Raw  bool
}

func (*Mustache) node() {}

// Partial is a {{> name [ctx] [key=val ...]}} partial invocation.
// Name is a PathRef or StringLit (the partial name).
// Ctx is the explicit context expression; nil means same-scope (pass current context).
type Partial struct {
	Name Expr
	Ctx  Expr
	Hash []HashEntry
}

func (*Partial) node() {}

// IfBlock is {{#if test}} … {{else}} … {{/if}} or {{#unless test}} … {{/unless}}.
// Hash holds block-level options like includeZero=true.
type IfBlock struct {
	Unless bool
	Test   Expr
	Hash   []HashEntry
	Body   []Node
	Else   []Node
}

func (*IfBlock) node() {}

// WithBlock is {{#with value [as |param|]}} … {{/with}}.
type WithBlock struct {
	Value       Expr
	BlockParams []string
	Body        []Node
	Else        []Node
}

func (*WithBlock) node() {}

// EachBlock is {{#each collection [as |item [index]|]}} … {{/each}}.
// The "in" keyword form ({{#each in collection}}) is normalised to the same node.
type EachBlock struct {
	Collection  Expr
	BlockParams []string
	Body        []Node
	Else        []Node
}

func (*EachBlock) node() {}

// LayoutBlock is a {{#block "name"}} … {{/block}} slot placeholder.
type LayoutBlock struct {
	Name string
	Body []Node
}

func (*LayoutBlock) node() {}

// LayoutPartial is a {{#partial "name"}} … {{/partial}} content override.
type LayoutPartial struct {
	Name string
	Body []Node
}

func (*LayoutPartial) node() {}

// Block is a generic block helper invocation or a universal section.
// Used for custom registered helpers and for {{#anyPath}} … {{/anyPath}} sections.
type Block struct {
	Name        string
	Params      []Expr
	Hash        []HashEntry
	BlockParams []string
	Body        []Node
	Else        []Node
}

func (*Block) node() {}
