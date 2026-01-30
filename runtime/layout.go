package runtime

// Blocks holds content for layout block/partial: child templates register
// content with {{#partial "name"}}...{{/partial}}, layout outputs it with
// {{#block "name"}}default{{/block}}. When Blocks is nil, {{#block}} renders
// only its default body and {{#partial}} renders its body to the writer.
type Blocks struct {
	m map[string]string
}

// NewBlocks returns a new Blocks store for use with layout templates.
func NewBlocks() *Blocks {
	return &Blocks{m: make(map[string]string)}
}

// Get returns the content registered for name and true, or "" and false.
func (b *Blocks) Get(name string) (string, bool) {
	if b == nil || b.m == nil {
		return "", false
	}
	s, ok := b.m[name]
	return s, ok
}

// Set stores content for the given block name (used by {{#partial}}).
func (b *Blocks) Set(name, content string) {
	if b == nil {
		return
	}
	if b.m == nil {
		b.m = make(map[string]string)
	}
	b.m[name] = content
}
