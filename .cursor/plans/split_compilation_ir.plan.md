---
name: Split compilation IR and typed Bundle
overview: Розділити компіляцію на фазу IR (BuildBundle) і фазу кодгену (EmitGo); структурувати Bundle по шаблонах (Template) і використовувати типізовані мапи (TemplateName, GoIdent, ContextTypeName).
todos: []
isProject: false
---

# Розділення компіляції на IR і генерацію Go

## Зроблено

- **Фаза 1 (IR):** `BuildBundle(templates, BuildOptions) (*Bundle, error)` — парсинг і контекстне виведення без генерації Go.
- **Фаза 2 (Codegen):** `EmitGo(bundle *Bundle, opts CodegenOptions) ([]byte, error)` — генерація Go з готового Bundle.
- **CompileTemplates** = BuildBundle + EmitGo (оркестратор).
- **Bundle** структурований по шаблонах: `Templates map[TemplateName]*Template`, кожен `Template` містить Parsed, FuncName (GoIdent), TypeTree, SourcePath.
- **Типізовані мапи:** `PartialParamTypes map[TemplateName]ContextTypeName`, `TypeSet map[TemplateName]map[ContextTypeName]bool`, `CanonicalType map[TemplateName]ContextTypeName`, `PrimaryCaller map[TemplateName]TemplateName`.
- **TypeNode** експортований (Fields, SliceElem, IsSlice) для інспекції дерева типів зовні.
- Тести на BuildBundle (type trees, partials, canonical types) без генерації Go.

## Типи

- `TemplateName` — ім'я шаблону (ключ для Templates і в partial-мапах).
- `GoIdent` — Go-ідентифікатор (наприклад "Main").
- `ContextTypeName` — ім'я типу контексту (наприклад "MainContext").
