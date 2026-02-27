# Built-in Helpers

go-hbars includes a comprehensive helpers library matching Handlebars.js core and handlebars-helpers 7.4. **Core helpers are automatically included by default** - no need to specify them unless you want to override or disable them.

## Using Helpers

**Using default core helpers (simplest):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates
// All core helpers are available automatically
```

**Selecting specific core helpers:**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -no-core-helpers \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers Upper,Lower,FormatDate
```

**Disabling core helpers and using custom ones:**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -no-core-helpers \
//  -import github.com/you/custom-helpers \
//  -helpers MyHelper,AnotherHelper
```

**Simple helper (local function):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=Upper

func Upper(args runtime.HelperArgs) (any, error) {
	if len(args.Args) == 0 {
		return "", nil
	}
	return strings.ToUpper(runtime.Stringify(args.Args[0])), nil
}
```

**Using the new shorthand syntax (recommended):**
```go
// Import a package and register multiple helpers
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers Upper,Lower,FormatDate

// With aliased imports
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -import extra:github.com/you/extra-helpers \
//  -helpers Upper,Lower \
//  -helpers extra:CustomHelper,extra:AnotherHelper

// Override helper names
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -import github.com/andriyg76/go-hbars/helpers/handlebars \
//  -helpers myUpper=Upper,myLower=Lower
```

**Legacy syntax (still supported):**
```go
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates -helper upper=github.com/you/helpers:Upper

// Multiple helpers
//go:generate hbc -in ./templates -out ./templates_gen.go -pkg templates \
//  -helper upper=Upper -helper lower=github.com/you/helpers:Lower
```

**Programmatic access (for advanced use cases):**
```go
import (
	"github.com/andriyg76/go-hbars/helpers"
	"github.com/andriyg76/go-hbars/internal/compiler"
)

registry := helpers.Registry()
compilerHelpers := make(map[string]compiler.HelperRef)
for name, ref := range registry {
	compilerHelpers[name] = compiler.HelperRef{
		ImportPath: ref.ImportPath,
		Ident:      ref.Ident,
	}
}
opts := compiler.Options{
	PackageName: "templates",
	Helpers:     compilerHelpers,
}
```

## Available Helpers

### String Helpers

- `upper`, `lower` - Convert case
- `capitalize`, `capitalizeAll` - Capitalize words
- `truncate` - Truncate strings with optional suffix
- `reverse` - Reverse a string
- `replace` - Replace substrings
- `stripTags`, `stripQuotes` - Remove HTML tags or quotes
- `join`, `split` - Join/split arrays with separator

### Comparison Helpers

- `eq`, `ne` - Equality checks
- `lt`, `lte`, `gt`, `gte` - Numeric comparisons
- `and`, `or`, `not` - Logical operations

### Date Helpers

- `formatDate` - Format dates with custom format (Go time format)
- `now` - Current time
- `ago` - Human-readable time ago

### Collection Helpers

- `lookup` - Look up values by key
- `default` - Fallback for empty values
- `length` - Get length of strings/arrays/objects
- `first`, `last` - Get first/last array element
- `inArray` - Check if value is in array

### Math Helpers

- `add`, `subtract`, `multiply`, `divide`, `modulo` - Arithmetic
- `floor`, `ceil`, `round`, `abs` - Rounding and absolute value
- `min`, `max` - Min/max of two numbers

### Number Helpers

- `formatNumber` - Format with precision and separator
- `toInt`, `toFloat`, `toNumber` - Type conversions
- `toFixed` - Fixed decimal places
- `toString` - Convert to string

### Object Helpers

- `has` - Check if object has property
- `keys`, `values` - Get object keys/values
- `size` - Get object/array size
- `isEmpty`, `isNotEmpty` - Empty checks

### URL Helpers

- `encodeURI`, `decodeURI` - URI encoding/decoding
- `stripProtocol`, `stripQuerystring` - URL manipulation

## Custom Helpers

You can implement custom helpers as regular Go functions and map them with `-helper name=Ident`. Helper functions must match this signature:

```go
func MyHelper(args runtime.HelperArgs) (any, error)
```

Arguments are resolved by the compiler before being passed; you receive `runtime.HelperArgs` with `Args` (positional) and `HashArgs` (named). For block invocations call `args.BlockFn()` and `args.InverseFn()` to render block content (writer captured in closure) (they return `error`).

```go
func FormatCurrency(args runtime.HelperArgs) (any, error) {
	if len(args.Args) == 0 {
		return "", nil
	}
	amount := runtime.Stringify(args.Args[0])
	symbol := "$"
	if args.HashArgs != nil {
		if s, ok := args.HashArgs["symbol"].(string); ok {
			symbol = s
		}
	}
	return fmt.Sprintf("%s%s", symbol, amount), nil
}
```

### Block Helpers

Block helpers use the same signature `func(args runtime.HelperArgs) (any, error)`. When used as a block, `args.IsBlock` is true and `args.Writer` is the template output writer. Call `args.BlockFn()` or `args.InverseFn()` with no arguments; each captures the template writer in a closure. They return `error`. The return value of the helper is ignored for block invocations.

```go
func MyBlockHelper(args runtime.HelperArgs) (any, error) {
	if !args.IsBlock {
		return nil, nil
	}
	if args.BlockFn != nil {
		if err := args.BlockFn(); err != nil {
			return nil, err
		}
	}
	if args.InverseFn != nil {
		if err := args.InverseFn(); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
```

Example: conditional block helper (see [Template API](api.md) for full details):

```go
func IfHelper(args runtime.HelperArgs) (any, error) {
	if !args.IsBlock || len(args.Args) == 0 {
		return nil, fmt.Errorf("if requires a condition and block")
	}
	condition := args.Args[0]
	if ok, _ := runtime.IsTruthy(condition); ok {
		if args.BlockFn != nil {
			return nil, args.BlockFn()
		}
	} else if args.InverseFn != nil {
		return nil, args.InverseFn()
	}
	return nil, nil
}
```

