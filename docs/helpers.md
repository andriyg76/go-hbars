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

func Upper(args []any) (any, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToUpper(runtime.Stringify(args[0])), nil
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
func MyHelper(args []any) (any, error)
```

Arguments are resolved by the compiler before being passed; you receive evaluated values. No context is passed. Hash arguments are passed as the last element in `args`. Use `runtime.HashArg(args)` to retrieve them:

```go
func FormatCurrency(args []any) (any, error) {
	if len(args) == 0 {
		return "", nil
	}
	amount := runtime.Stringify(args[0])
	hash, _ := runtime.HashArg(args)
	symbol := "$"
	if hash != nil {
		if s, ok := hash["symbol"].(string); ok {
			symbol = s
		}
	}
	return fmt.Sprintf("%s%s", symbol, amount), nil
}
```

### Block Helpers

Block helpers use signature `func(args []any) error`. When used as a block, the helper receives `runtime.BlockOptions` as the last element of `args`. Use `runtime.GetBlockOptions(args)` to retrieve it. `BlockOptions.Fn` and `BlockOptions.Inverse` have type `func(io.Writer) error` (they receive only the writer):

```go
func MyBlockHelper(args []any) error {
	opts, ok := runtime.GetBlockOptions(args)
	if !ok {
		return fmt.Errorf("block helper did not receive BlockOptions")
	}
	if opts.Fn != nil {
		if err := opts.Fn(w); err != nil {
			return err
		}
	}
	return nil
}
```

Block helpers can conditionally render the main block (`opts.Fn`) or the inverse/else block (`opts.Inverse`). When invoked from generated code, only `args` is passed; the writer `w` is in scope in the generated render function (see [Template API](api.md) for details):

```go
func IfHelper(args []any) error {
	opts, ok := runtime.GetBlockOptions(args)
	if !ok {
		return fmt.Errorf("if helper must be used as a block")
	}
	if len(args) == 0 {
		return fmt.Errorf("if requires a condition")
	}
	condition := args[0]
	if runtime.IsTruthy(condition) {
		if opts.Fn != nil {
			return opts.Fn(w)
		}
	} else if opts.Inverse != nil {
		return opts.Inverse(w)
	}
	return nil
}
```

