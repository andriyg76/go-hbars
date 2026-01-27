package runtime

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// LazySlot records a deferred {{#block "name"}} for layout-first (Direction B).
// The placeholder is written during layout render; content is resolved after content runs.
type LazySlot struct {
	Placeholder string
	Name        string
	DefaultFn   func(ctx *Context, w io.Writer) error
}

// LazySlotsRecorder collects lazy slots during layout render.
// When ctx.LazySlots is set, Block() writes a placeholder and records here instead of resolving now.
type LazySlotsRecorder struct {
	nextID int
	slots  []LazySlot
}

// NewLazySlotsRecorder returns a recorder for layout-first render.
func NewLazySlotsRecorder() *LazySlotsRecorder {
	return &LazySlotsRecorder{slots: make([]LazySlot, 0)}
}

// Record adds a lazy slot and returns the placeholder string to write to output.
func (r *LazySlotsRecorder) Record(name string, defaultFn func(ctx *Context, w io.Writer) error) string {
	r.nextID++
	ph := fmt.Sprintf("\x1eSLOT%d\x1e", r.nextID)
	r.slots = append(r.slots, LazySlot{Placeholder: ph, Name: name, DefaultFn: defaultFn})
	return ph
}

// Slots returns the recorded slots for resolution.
func (r *LazySlotsRecorder) Slots() []LazySlot {
	return r.slots
}

// ResolveLazySlots replaces placeholders in buf with content from blocks or default Fn.
// Call after layout (and content partial) have finished. ctx is used when running DefaultFn.
func ResolveLazySlots(buf *bytes.Buffer, slots []LazySlot, blocks map[string]string, ctx *Context) error {
	s := buf.String()
	for _, slot := range slots {
		var content string
		if blocks != nil {
			if c, ok := blocks[slot.Name]; ok && c != "" {
				content = c
			}
		}
		if content == "" && slot.DefaultFn != nil && ctx != nil {
			var b bytes.Buffer
			if err := slot.DefaultFn(ctx, &b); err != nil {
				return err
			}
			content = b.String()
		}
		s = strings.ReplaceAll(s, slot.Placeholder, content)
	}
	buf.Reset()
	_, err := buf.WriteString(s)
	return err
}
