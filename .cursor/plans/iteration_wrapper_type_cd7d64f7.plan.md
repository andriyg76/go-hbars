---
name: Iteration Wrapper Type
overview: Introduce a public `Iter[T]` wrapper type in runtime that handles iteration complexity including @first/@last computation via one-element-forward-fetch. Supports typed slices/maps, Go iterators, and runtime type assertion for arbitrary collections.
todos:
  - id: iter-types
    content: Create runtime/iterate.go with Iter[T] generic type, IterItem, and forward-fetch logic
    status: pending
  - id: iter-constructors
    content: Implement WrapSlice, WrapMap, WrapIter, and Wrap (auto-detect) constructors
    status: pending
  - id: reflect-fallback
    content: Implement reflect-based fallback in Wrap() for arbitrary struct/slice/map types
    status: pending
  - id: refactor-emit
    content: Refactor emitEachBlock() in compile.go to generate cursor-based iteration code
    status: pending
  - id: tests
    content: Add tests for all iteration modes (typed, map-backed, iterators, edge cases)
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

## Goals

1. **Public runtime type** - users can explicitly wrap collections when forming contexts
2. **Support multiple input types**: typed slices, typed maps, Go iterators (`iter.Seq`), `[]any`, `map[string]any`
3. **Runtime type assertion** - for arbitrary structures when type is unknown at compile time
4. **Explicit user wrapping** - `runtime.WrapSlice(myItems)` at context creation time

## Design

### Public API in runtime package

```go
// Iter wraps any iterable collection with forward-fetch for @first/@last
type Iter[T any] struct {
    next    func() (key any, value T, ok bool)  // source iterator
    peeked  bool
    peekKey any
    peekVal T
    peekOk  bool
    index   int
    started bool
}

// IterItem represents the current iteration state
type IterItem[T any] struct {
    Index int    // 0-based index
    Key   any    // map key or index (string for maps, int for slices)
    Value T      // the element value
    First bool   // true if first element
    Last  bool   // true if last element (determined by peek)
}

// Public constructors
func WrapSlice[T any](s []T) *Iter[T]
func WrapMap[K comparable, V any](m map[K]V) *Iter[V]
func WrapIter[T any](seq iter.Seq[T]) *Iter[T]
func WrapIter2[K, V any](seq iter.Seq2[K, V]) *Iter[V]
func Wrap(v any) *Iter[any]  // auto-detect via type assertion + reflect fallback

// Iteration methods
func (it *Iter[T]) Next() bool
func (it *Iter[T]) Item() IterItem[T]
func (it *Iter[T]) Empty() bool  // true if no elements (for {{else}})
```

### How Forward-Fetch Works for @last

```
Collection: [A, B, C]

Initial:    peeked=false
Next() #1:  peek ahead -> peekVal=A, peekOk=true
            current=A, peek ahead -> peekVal=B, peekOk=true
            First=true (index==0), Last=false (peekOk)
Next() #2:  current=B, peek ahead -> peekVal=C, peekOk=true  
            First=false, Last=false
Next() #3:  current=C, peek ahead -> peekOk=false
            First=false, Last=true (no more elements)
Next() #4:  returns false
```

### User-Facing Usage

**Explicit wrapping at context creation:**

```go
type MyContext struct {
    items *runtime.Iter[Order]
}

func NewMyContext(orders []Order) MyContext {
    return MyContext{
        items: runtime.WrapSlice(orders),
    }
}

func (c MyContext) Items() *runtime.Iter[Order] {
    return c.items
}
```

**Auto-wrap for JSON/map-backed contexts:**

```go
// In generated FromMap code or runtime
cursor := runtime.Wrap(data["orders"])  // auto-detects []any or map[string]any
for cursor.Next() {
    item := cursor.Item()
    // item.Index, item.Key, item.Value, item.First, item.Last
}
```

### Generated Code Changes

**For typed contexts** (compile-time known type):

```go
// Generated code - typed slice
items := data.Items()  // returns *runtime.Iter[OrderContext]
for items.Next() {
    item := items.Item()
    // item.Value is OrderContext, item.First, item.Last available
}
if items.Empty() {
    // {{else}} block
}
```

**For map-backed contexts** (runtime type assertion):

```go
// Generated code - map-backed
cursor := runtime.Wrap(runtime.ResolvePathValue(ctx, "orders"))
for cursor.Next() {
    item := cursor.Item()
    // Unified interface regardless of underlying type
}
```

### Wrap() Auto-Detection Logic

```go
func Wrap(v any) *Iter[any] {
    switch x := v.(type) {
    case *Iter[any]:
        return x  // already wrapped
    case []any:
        return WrapSlice(x)
    case map[string]any:
        return WrapMap(x)
    case iter.Seq[any]:
        return WrapIter(x)
    default:
        // Reflect fallback for arbitrary slices/maps
        return wrapReflect(v)
    }
}

func wrapReflect(v any) *Iter[any] {
    rv := reflect.ValueOf(v)
    switch rv.Kind() {
    case reflect.Slice, reflect.Array:
        return newReflectSliceIter(rv)
    case reflect.Map:
        return newReflectMapIter(rv)
    default:
        return emptyIter[any]()
    }
}
```

## Key Files

- [runtime/iterate.go](runtime/iterate.go) (new) - `Iter[T]`, `IterItem[T]`, constructors
- [internal/compiler/compile.go](internal/compiler/compile.go) - Refactor `emitEachBlock()` 
- [helpers/handlebars/collection.go](helpers/handlebars/collection.go) - May use Iter for First/Last helpers

## Benefits

- **Unified API** for all iteration types (slices, maps, iterators)
- **@last via peek** - no upfront length computation needed, works with iterators
- **Public type** - users can pre-wrap collections for optimal performance
- **Extensible** - easy to add `@even`/`@odd`, `@depth` in future
- **Cleaner generated code** - single iteration pattern
- **Go 1.23+ iterators** - native support for `iter.Seq`/`iter.Seq2`