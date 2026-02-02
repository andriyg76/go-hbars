---
name: Iteration Wrapper Type
overview: Introduce an `IterCursor` wrapper type in the compiler that handles iteration complexity including @first/@last computation via one-element-forward-fetch, simplifying `emitEachBlock()` and supporting both typed and map-backed contexts.
todos:
  - id: iter-types
    content: Create runtime/iterate.go with IterItem struct and IterCursor with forward-fetch logic
    status: pending
  - id: cursor-constructors
    content: Implement NewSliceCursor, NewMapCursor with reflect-based iteration
    status: pending
  - id: refactor-emit
    content: Refactor emitEachBlock() in compile.go to generate cursor-based iteration code
    status: pending
  - id: update-scope
    content: Update typedScope handling to integrate with cursor pattern
    status: pending
  - id: tests
    content: Add tests for cursor iteration (slices, maps, empty, single element edge cases)
    status: pending
isProject: false
---

# Iteration Wrapper Type for hbars Renderer

## Problem

The current `emitEachBlock()` in [internal/compiler/compile.go](internal/compiler/compile.go) (lines 1162-1306) has significant complexity:

- Dual code paths for typed slices vs map-backed contexts
- Runtime type assertions with slice/map fallback
- Manual scope management with `pushTypedScope`/`popTypedScope`
- No unified handling of `@first`/`@last`/`@index`/`@key`

## Solution

Introduce an `IterCursor` abstraction that:

1. Provides a unified iteration interface for both typed and map-backed contexts
2. Uses **one-element-forward-fetch** to compute `@last` lazily (peek ahead to know if current is last)
3. Encapsulates `@first`/`@last`/`@index`/`@key` in the cursor state

## Design

### Cursor-Based Iteration Pattern

```go
// IterItem represents a single iteration element
type IterItem struct {
    Index int    // 0-based index
    Key   string // map key (empty for slices)
    Value any    // the element value
    First bool   // true if first element
    Last  bool   // true if last element (determined by peek)
}

// IterCursor iterates over slices or maps with peek-ahead for @last
type IterCursor struct {
    // internal state: current, peeked, index, done
}

func NewSliceCursor(slice any) *IterCursor
func NewMapCursor(m map[string]any) *IterCursor

func (c *IterCursor) Next() bool      // advance; returns false when done
func (c *IterCursor) Item() IterItem  // current item with First/Last computed
```

### How Forward-Fetch Works

```
Collection: [A, B, C]

Step 1: peek=A, current=nil        -> Next() sets current=A, peek=B, First=true, Last=false
Step 2: current=A, peek=B          -> Next() sets current=B, peek=C, First=false, Last=false  
Step 3: current=B, peek=C          -> Next() sets current=C, peek=nil, First=false, Last=true
Step 4: current=C, peek=nil, done  -> Next() returns false
```

### Compiler Changes

Update `emitEachBlock()` to generate code using the cursor pattern:

**For typed contexts** (compile-time known slice type):

```go
// Generated code (typed)
cursor := runtime.NewTypedSliceCursor(data.Items())
for cursor.Next() {
    item := cursor.Item()
    // item.Index, item.First, item.Last available
    // item.Value is typed
}
```

**For map-backed contexts** (runtime type assertion):

```go
// Generated code (map-backed)
cursor := runtime.NewCursor(ctx.Get("items"))
for cursor.Next() {
    item := cursor.Item()
    // Unified interface regardless of []any or map[string]any
}
```

## Key Files to Modify

- [runtime/iterate.go](runtime/iterate.go) (new) - `IterItem`, `IterCursor`, cursor constructors
- [internal/compiler/compile.go](internal/compiler/compile.go) - Refactor `emitEachBlock()` to use cursor pattern
- [internal/compiler/compile.go](internal/compiler/compile.go) - Update `typedScope` to work with cursor

## Benefits

- Single iteration abstraction for all collection types
- `@last` computed lazily via peek (no upfront length needed)
- Cleaner generated code
- Easier to extend (e.g., add `@even`/`@odd` in future)
- Reduced duplication in `emitEachBlock()`