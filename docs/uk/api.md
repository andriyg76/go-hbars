# Довідник API шаблонів

Цей документ описує рантайм API для роботи зі скомпільованими Handlebars-шаблонами.

## Базове використання

Після компіляції шаблонів через `hbc` ви отримуєте згенеровані функції для кожного шаблону. Компілятор випромінює **типізовані контексти** (наприклад `MainContext`), виведені з виразів у шаблоні:

```go
import "github.com/your/project/templates"

// Рендер у writer (data має задовольняти тип контексту шаблону, напр. MainContext)
var b strings.Builder
if err := templates.RenderMain(&b, data); err != nil {
    // обробити помилку
}
out := b.String()

// Або використати обгортку для рядка. Для даних-мапи використовуйте MainContextFromMap(data).
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

## Згенеровані функції

Для кожного файлу шаблону (наприклад `main.hbs`) компілятор генерує:

1. **Внутрішня функція рендеру**: `renderMain(data MainContext, w io.Writer, root any) error` (використовується парціалами; `root` — кореневий контекст викликача для `@root`)
2. **Публічна функція рендеру**: `RenderMain(w io.Writer, data MainContext) error`
3. **Обгортка для рядка**: `RenderMainString(data MainContext) (string, error)`

Тип контексту (наприклад `MainContext`) — це інтерфейс, виведений зі шляхів, що використовуються в шаблоні; можна передати структуру або `map[string]any` з потрібними полями.

## Пакет runtime

Пакет `runtime` надає типи та утиліти для згенерованого коду та власних хелперів.

### Вивід

```go
// WriteEscaped записує екрановане значення у writer
runtime.WriteEscaped(w, value)

// WriteRaw записує сире значення у writer
runtime.WriteRaw(w, value)

// Stringify перетворює значення на рядкове представлення
str := runtime.Stringify(value)
```

### Аргументи хелперів

Хелпери отримують один аргумент типу `runtime.HelperArgs`:

```go
type HelperArgs struct {
    HashArgs   map[string]any  // іменовані (hash) аргументи
    Args       []any           // позиційні аргументи
    BlockFn    func() error    // коли IsBlock: рендерить основний блок (writer у замиканні); інакше nil
    InverseFn  func() error    // коли IsBlock і є else: рендерить блок else; може бути nil
    IsBlock    bool            // true для виклику блокового хелпера
}
```

### Істинність

```go
// IsTruthy перевіряє, чи значення істинне
if runtime.IsTruthy(value) {
    // ...
}
```

### Безпечні рядки

```go
// SafeString позначає значення як попередньо екранований HTML
safe := runtime.SafeString("<b>bold</b>")
```

### Контекст і парціали

```go
// LookupPath повертає значення за крапковим шляхом від root (наприклад "title", "user.name").
// Root може бути map[string]any або реалізовувати Raw() any з мапою.
// Використовується згенерованим кодом для @root.xxx у парціалах, коли root приходить з іншого шаблону.
val := runtime.LookupPath(root, "title")
```

## Функції-хелпери

Усі хелпери (прості та блокові) мають однакову сигнатуру:

```go
func MyHelper(args runtime.HelperArgs) (any, error)
```

Аргументи **обчислюються компілятором** перед передачею; ви отримуєте `HelperArgs` з полями `Args` (позиційні) та `HashArgs` (іменовані). Для блокових викликів повернене значення ігнорується; викликайте `args.BlockFn()` та `args.InverseFn()` для рендеру контенту блоку (writer у замиканні) (вони повертають `error`).

### Доступ до аргументів

```go
func MyHelper(args runtime.HelperArgs) (any, error) {
    // Позиційні аргументи (вже обчислені)
    if len(args.Args) == 0 {
        return nil, fmt.Errorf("missing argument")
    }
    firstArg := args.Args[0]
    
    // Hash-аргументи (пари key=value)
    if args.HashArgs != nil {
        value := args.HashArgs["key"]
    }
    
    return result, nil
}
```

### Блокові хелпери

Коли хелпер використовується як блок (`{{#name}}...{{/name}}`), `args.IsBlock` дорівнює true і `args.Writer` — це writer виводу шаблону. Викликайте `args.BlockFn()` або `args.InverseFn()` без аргументів; кожен є замиканням, що вже захоплює writer шаблону. Повертають `error`. Блок рендериться лише при виклику.

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

## Партіали

Партіали автоматично реєструються в згенерованому коді:

```go
// partials map (внутрішня): ім'я шаблону -> func(data any, w io.Writer) error
partials["header"](data, w)
```

У шаблонах вони використовуються через `{{> header}}` або `{{> (lookup ...) }}`.

## Типи даних

### Дані контексту

Дані контексту для шаблону задовольняють згенерований інтерфейс контексту (наприклад `MainContext`). На практиці можна передавати:

- Мапи (`map[string]any`)
- Структури (з експортованими полями або JSON-тегами)
- Компілятор також генерує конструктори `XxxContextFromMap` для побудови контексту з `map[string]any`.

### Hash-аргументи

Hash-аргументи доступні в `args.HashArgs` (тип `map[string]any`) у структурі `HelperArgs`.

## Обробка помилок

Усі функції рендеру повертають помилки. Типові ситуації:

- Відсутній шаблон або парціал (помилка компіляції)
- Відсутній хелпер (помилка компіляції)
- Помилки рантайму в хелперах
- Невірні типи даних
- Помилки рантайму в хелперах

Завжди перевіряйте помилки. Коли дані у вигляді `map[string]any` (наприклад з JSON), використовуйте згенерований `XxxContextFromMap`, щоб дані задовольняли тип контексту:

```go
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
if err != nil {
    log.Fatal(err)
}
```

## Продуктивність

- Шаблони компілюються в Go-код, тому виконання швидке
- Немає парсингу шаблонів під час виконання
- Типи контексту визначаються під час компіляції; хелпери отримують уже обчислені аргументи

## Приклади

### Простий рендер шаблону

```go
data := map[string]any{
    "title": "Hello",
    "user": map[string]any{
        "name": "Alice",
    },
}
// Якщо шаблон використовує ці шляхи, згенерований MainContext дозволить мапу або структуру.
// Використовуйте MainContextFromMap(data), якщо компілятор його згенерував, або передайте структуру.
out, err := templates.RenderMainString(templates.MainContextFromMap(data))
```

### Власний хелпер

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

### Блоковий хелпер

```go
func IfHelper(args runtime.HelperArgs) (any, error) {
    if !args.IsBlock || len(args.Args) < 1 {
        return nil, fmt.Errorf("if потребує умову та блок")
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

Примітка: вбудовані `if`/`unless`/`each`/`with` реалізовані компілятором; приклад вище ілюструє рантайм API для власних блокових хелперів: `HelperArgs.BlockFn()` та `InverseFn()` (writer у замиканні).

## Див. також

- [Скомпільований файл шаблонів](compiled-templates.md) — що генерує компілятор (типи контексту, RenderXxx, FromMap)
- [Синтаксис Handlebars](syntax.md) — вирази та блоки
- [Вбудовані хелпери](helpers.md) — доступні хелпери та реєстрація власних
