# Problem: Template context interface invariance with shared partials

## What we are trying to fix

When multiple root templates include the **same partial** (e.g. `{{> menu}}`) with the same scope (no explicit context), the compiler used to generate **one nested context interface per template** for that partial:

- `DefaultContext.Menu() DefaultMenuContext`
- `FirstpageContext.Menu() FirstpageMenuContext`
- `TeatrAfishaContext.Menu() TeatrAfishaMenuContext`
- …

Each of these types (`DefaultMenuContext`, `FirstpageMenuContext`, …) is a **different Go interface**. They are not interchangeable: a value that implements `FirstpageMenuContext` does not implement `DefaultMenuContext`, even if both have the same methods, because Go interfaces are nominal by name.

## Why this is a problem

1. **Type mismatch in generic code**  
   If you have a single “page” type that must work with multiple templates (e.g. `FirstpageContext`, `DefaultContext`), you cannot express “this page has a menu” as one interface. You would need something like `interface{ Menu() DefaultMenuContext }` for default and `interface{ Menu() FirstpageMenuContext }` for firstpage, and they are incompatible.

2. **No shared contract for partial context**  
   Callers that only care about “something that has a menu” cannot use a single interface. Every template that includes `menu` gets its own `*MenuContext` type, so there is no single `MenuContext` that all of them satisfy.

3. **Hierarchical partials**  
   When partials include other partials (e.g. `menu` includes `menuItem` inside `{{#each items}}`), the same issue appears at each level: we want one canonical “menu item” context for the partial, not one per root template.

## Desired behaviour

- For each **partial** that is included by multiple templates (with the same scope), the compiler should emit a **single canonical context interface** (e.g. `MenuContext`, `MenuItemsItemContext` for items under `menu`).
- Only **one** “primary” caller per partial (e.g. the template named `default`, or the first alphabetically) keeps a **template-prefixed** nested interface (e.g. `DefaultMenuContext`) so that the default template’s context type stays rich and consistent.
- All **other** callers of that partial should expose the **canonical** interface in their context (e.g. `FirstpageContext.Menu() MenuContext`), not a template-specific one (`FirstpageMenuContext`).
- The same rules apply to **nested partials** (partial-in-partial) and to the generated `*ContextData` structs, so that `FromMap` and typed contexts align.

## Summary

**Problem:** Multiple templates including the same partial produced multiple, incompatible context interfaces, breaking shared “has a menu” abstractions and generic page handling.

**Fix:** Emit one canonical partial context per partial (e.g. `MenuContext`), use it for all non-primary callers, and keep a single primary-caller-specific interface (e.g. `DefaultMenuContext`) only for the chosen primary template. Apply the same logic for hierarchical partials and for generated data types.
