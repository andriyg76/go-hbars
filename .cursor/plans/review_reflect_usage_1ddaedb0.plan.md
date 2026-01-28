---
name: Review reflect usage
overview: "План послідовного перегляду кожного місця використання пакету `reflect` у go-hbars: мета використання, чи можна обійтись без нього, безпека та продуктивність."
todos: []
isProject: false
---

# План перегляду використання reflect у go-hbars

У проєкті `reflect` використовується в **6 файлах** у різних контекстах. Нижче — перелік випадків і що саме перевірити в кожному.

---

## 1. [internal/processor/renderer.go](internal/processor/renderer.go)

**Що робить:** при передачі в `NewCompiledTemplateRenderer` структури (не `map[string]func(...)`) — через `reflect.ValueOf` / `Type().NumMethod()` знаходяться методи з префіксом `Render`, зберігаються як `map[string]reflect.Value`, при рендері викликаються через `reflect.Value.Call(args)`.

**Що переглянути:**

- Чи потрібна підтримка "struct з Render* методами" далі, чи достатньо лише bootstrap-режиму з `map[string]func(io.Writer, any) error` (без reflect при виклику).
- Якщо залишаємо reflect: перевірити обробку панік при `Call` (наприклад, невідповідність сигнатур), можливість обгорнути в `recover`.
- Документувати в коді/README: коли використовується reflect-шлях, а коли — прямий виклик з map.

---

## 2. [runtime/context.go](runtime/context.go) — `lookupValue`

**Що робить:** резолвить один сегмент шляху (наприклад `user.name` → один крок по `val` і ключу `key`). Через `reflect` підтримує: `map` з string-ключами, `struct` (FieldByName + json-теги), slice/array по індексу, розгортання pointer/interface.

**Що переглянути:**

- Чи можна скоротити reflect: наприклад, спочатку type switch на `map[string]any`, `map[string]interface{}`, слайси — і лише для "інших" типів (struct, нестандартні мапи) використовувати reflect.
- Перевірити випадки nil (вже є на початку) та неекспортовані поля struct (PkgPath перевіряється) — чи немає небажаного доступу.
- Продуктивність: `lookupValue` викликається на кожен сегмент шляху; якщо hot path — варто подивитись кешування по типу або звуження reflect тільки до struct.

---

## 3. [runtime/blocks.go](runtime/blocks.go)

**Що робить:**

- **IsNumericZero** — через reflect визначає числові типи і порівнює з нулем (після type switch на `json.Number`).
- **IsTruthy** — після великого type switch використовує reflect для "решти" типів (bool, string, числа, slice/map за довжиною).
- **Iterate** — для `#each`: через reflect обходить slice/array (Index) і map (MapKeys, MapIndex).

**Що переглянути:**

- **IsNumericZero**: чи достатньо поточного списку числових kind-ів чи потрібні додаткові (наприклад, complex).
- **IsTruthy**: чи не краще для відомих типів (наприклад, з runtime) взагалі не заходити в reflect, а мати явні типи в switch — щоб зменшити залежність від reflect у hot path.
- **Iterate**: для slice/array/map reflect тут виглядає обґрунтовано (довільний тип елементів). Перевірити поведінку для nil, порожніх мап і мап з не-string ключами (вже повертається nil) — задокументувати.

---

## 4. [helpers/util.go](helpers/util.go) — `IsTruthy` та `IsEmpty`

**Що робить:** після type switch по конкретних типах (bool, string, []any, map[string]any тощо) для "default" гілки використовує `reflect.ValueOf` і перевірки Kind (Slice, Map, Array, числові) для довжини/значення.

**Що переглянути:**

- Узгодити з [runtime/blocks.go](runtime/blocks.go): там теж є `IsTruthy`. Чи мають бути одна реалізація (наприклад, в runtime) і хелпери лише її викликають, чи навмисно дві різні семантики.
- Для `IsEmpty`: чи повна відповідність до Handlebars-семантики "empty" (включно з нулем для чисел) — порівняти з документацією/іншими реалізаціями.

---

## 5. [helpers/handlebars/collection.go](helpers/handlebars/collection.go) — `Length`

**Що робить:** для типів, відмінних від string, []any, []string, map[string]any, map[any]any, використовує reflect (Slice, Map, Array, String) щоб повернути `Len()`.

**Що переглянути:**

- Чи потрібна підтримка "екзотичних" типів (наприклад, []int, map[int]string) у шаблонах — якщо ні, можна обмежитись type switch без reflect і для невідомого типу повертати 0 або помилку.
- Якщо залишаємо reflect — переконатись, що для невалідних/непідтримуваних типів повертається 0 (поточна поведінка) і це задокументовано.

---

## 6. [runtime/context_test.go](runtime/context_test.go)

**Що робить:** використовує `reflect.DeepEqual` для порівняння результату `ResolvePath`/`ResolvePathParsed` з очікуваним значенням (наприклад, порівняння map/struct).

**Що переглянути:**

- Чи достатньо `DeepEqual` для тестів (враховує вкладені map/slice) чи в окремих кейсах краще порівнювати поля вручну. Залишити reflect.DeepEqual тут — стандартна практика; перегляд швидше перевірка, що тести покривають граничні випадки (nil, порожні структури).

---

## Рекомендований порядок перегляду

1. **Узгодити дублікати** — `IsTruthy` в runtime vs helpers ([runtime/blocks.go](runtime/blocks.go), [helpers/util.go](helpers/util.go)).
2. **Processor renderer** — чи потрібен reflect-шлях для struct, безпека викликів ([internal/processor/renderer.go](internal/processor/renderer.go)).
3. **Runtime hot path** — `lookupValue` та блоки: можливість звузити reflect або додати type switch ([runtime/context.go](runtime/context.go), [runtime/blocks.go](runtime/blocks.go)).
4. **Helpers** — Length та IsEmpty/IsTruthy: потреба в reflect vs фіксований набір типів ([helpers/util.go](helpers/util.go), [helpers/handlebars/collection.go](helpers/handlebars/collection.go)).
5. **Тести** — лише переконатись, що DeepEqual використано коректно ([runtime/context_test.go](runtime/context_test.go)).

Після перегляду можна винести короткі висновки в один документ (наприклад, `docs/reflect-usage.md`): де reflect залишаємо і чому, де прибираємо або звужуємо, і які контракти (типи даних шаблонів) офіційно підтримуються.