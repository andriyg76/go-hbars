package runtime

import "io"

// LazyBlockWriter is an io.Writer that can register lazy output blocks.
// A lazy block is written when Flush is called, preserving order with prior Write calls.
type LazyBlockWriter interface {
	io.Writer
	// WriteLazyBlock registers a block to be written at Flush. The block receives
	// the underlying writer and may write the "partial \"\" is not defined" message etc.
	WriteLazyBlock(block func(io.Writer))
	// Flush writes any buffered data and evaluates lazy blocks in order.
	Flush() error
}

// LazyWriter wraps w and implements LazyBlockWriter: Write and WriteLazyBlock
// are queued in order; Flush drains the queue to w.
type LazyWriter struct {
	w     io.Writer
	queue []lazyItem
}

type lazyItem struct {
	direct []byte
	block  func(io.Writer)
}

// NewLazyWriter returns a LazyBlockWriter that flushes to w.
func NewLazyWriter(w io.Writer) *LazyWriter {
	return &LazyWriter{w: w, queue: make([]lazyItem, 0)}
}

// Write implements io.Writer by queuing the bytes; they are written on Flush.
func (l *LazyWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	b := make([]byte, len(p))
	copy(b, p)
	l.queue = append(l.queue, lazyItem{direct: b})
	return len(p), nil
}

// WriteLazyBlock registers a block to run at Flush (writes to l.w).
func (l *LazyWriter) WriteLazyBlock(block func(io.Writer)) {
	if block == nil {
		return
	}
	l.queue = append(l.queue, lazyItem{block: block})
}

// Flush writes all queued direct bytes and runs lazy blocks in order.
func (l *LazyWriter) Flush() error {
	for _, item := range l.queue {
		if item.direct != nil {
			if _, err := l.w.Write(item.direct); err != nil {
				return err
			}
		} else if item.block != nil {
			item.block(l.w)
		}
	}
	l.queue = l.queue[:0]
	return nil
}
