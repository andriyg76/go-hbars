---
name: Review reflect usage
overview: "Перегляд використання пакету reflect у go-hbars: 4 файли з reflect, аналіз необхідності, безпеки та продуктивності."
todos:
  - id: dedup-istruthy
    content: Узгодити/обєднати IsTruthy в runtime/blocks.go та helpers/util.go
    status: pending
  - id: review-renderer
    content: "Перевірити processor/renderer.go: потреба struct-шляху, recover для Call"
    status: completed
  - id: review-blocks
    content: "Перевірити runtime/blocks.go: чи можна звузити reflect"
    status: pending
  - id: review-helpers
    content: "Перевірити helpers: Length, IsEmpty — потреба екзотичних типів"
    status: pending
  - id: document
    content: Задокументувати рішення щодо reflect у docs/reflect-usage.md
    status: pending
isProject: false
---

# Перегляд використання reflect у go-hbars

У проєкті `reflect` використовується в **4 файлах**. Нижче — актуальний перелік і що переглянути в кожному.

---

## 1. [internal/processor/renderer.go](internal/processor/renderer.go)

**Функції:** `NewCompiledTemplateRenderer`, `Render`

**Що робить:**

- При передачі struct (не `map[string]func(...)`) — через `reflect.ValueOf` / `Type().NumMethod()` знаходяться методи з префіксом `Render`
- Зберігаються як `map[string]reflect.Value`
- При рендері викликаються через `reflect.Value.Call(args)`
- Для `map[string]func(io.Writer, any) error` — прямий виклик без reflect

**Що переглянути:**

- Чи потрібна підтримка struct з `Render*` методами, чи достатньо bootstrap-режиму з map
- Якщо залишаємо reflect: перевірити обробку панік при `Call` (невідповідність сигнатур), можливість `recover`
- Документувати коли використовується reflect-шлях, а коли прямий виклик

---

## 2. [runtime/blocks.go](runtime/blocks.go)

**Функції:** `IsNumericZero`, `IsTruthy`, `IncludeZeroTruthy`

**Що робить:**

- `IsNumericZero` — через reflect визначає числові типи і порівнює з нулем (після type switch на `json.Number`)
- `IsTruthy` — після великого type switch на конкретні типи (bool, string, int, float...) для "інших" типів використовує reflect
- `IncludeZeroTruthy` — комбінує попередні дві

**Що переглянути:**

- Чи достатньо поточного списку числових kind (complex не підтримується)
- Чи можна звузити reflect тільки до справді невідомих типів
- Поведінка для nil та pointer/interface (вже обробляються через unwrap loop)

---

## 3. [helpers/util.go](helpers/util.go)

**Функції:** `IsTruthy`, `IsEmpty`

**Що робить:**

- Після type switch по конкретних типах (bool, string, []any, map[string]any) для "default" гілки використовує reflect
- Перевіряє Kind (Slice, Map, Array, числові) для довжини/значення

**Що переглянути:**

- **Дублікат з runtime/blocks.go**: обидва мають `IsTruthy` — узгодити чи об'єднати
- Для `IsEmpty`: перевірити відповідність Handlebars-семантиці "empty"
- Чи потрібен fallback на `true` в IsTruthy для невідомих типів

---

## 4. [helpers/handlebars/collection.go](helpers/handlebars/collection.go)

**Функція:** `Length`

**Що робить:**

- Для типів відмінних від string, []any, []string, map[string]any, map[any]any використовує reflect
- Повертає `rv.Len()` для Slice, Map, Array, String

**Що переглянути:**

- Чи потрібна підтримка "екзотичних" типів ([]int, map[int]string) у шаблонах
- Якщо ні — можна обмежитись type switch без reflect
- Для невалідних типів повертає 0 — задокументувати

---

## Рекомендований порядок перегляду

1. **Узгодити дублікати** — `IsTruthy` в runtime vs helpers
2. **Processor renderer** — чи потрібен reflect-шлях для struct, безпека викликів
3. **Runtime hot path** — `IsNumericZero` та `IsTruthy`: можливість звузити reflect
4. **Helpers** — Length та IsEmpty: потреба в reflect vs фіксований набір типів

---

## Підсумок

| Файл | Функції | Reflect для |
|------|---------|-------------|
| processor/renderer.go | NewCompiledTemplateRenderer, Render | struct method discovery + Call |
| runtime/blocks.go | IsNumericZero, IsTruthy | числові типи, slice/map/array len |
| helpers/util.go | IsTruthy, IsEmpty | slice/map/array len, числа |
| helpers/handlebars/collection.go | Length | slice/map/array/string len |