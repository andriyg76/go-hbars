# Compatibility fixtures

This folder contains a small template set that exercises:

- hash arguments
- subexpressions
- data variables (`@index`, `@key`, `@first`, `@last`, `@root`)
- parent paths (`../`)
- block params (`as |item idx|`)
- dynamic partials
- whitespace control (`~`)
- raw blocks

## Helper requirements

The templates reference the following helpers:

- `upper` (string -> string)
- `lower` (string -> string)
- `lookup` (returns a value by key; used for dynamic partial names)
- `formatDate` (accepts a value and hash `format`)
- `default` (returns a fallback from hash `value`)

Hash arguments are passed as the last `runtime.Hash` in the helper args.

## Compile example

```
hbc -in ./examples/compat/templates \
  -out ./examples/compat/templates_gen.go \
  -pkg compat \
  -helper upper=Upper \
  -helper lower=Lower \
  -helper lookup=Lookup \
  -helper formatDate=FormatDate \
  -helper default=Default
```

Use `data.json` as input data when rendering `main.hbs`.
