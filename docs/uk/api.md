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

1. **Внутрішня функція рендеру**: `renderMain(data MainContext, w io.Writer, root any) error` (використовується партіалами; `root` — кореневий контекст викликача для `@root`)
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

```go
// HashArg витягує hash-аргументи з аргументів хелпера
hash, ok := runtime.HashArg(args)

// GetBlockOptions витягує опції блоку з аргументів хелпера (для блокових хелперів)
opts, ok := runtime.GetBlockOptions(args)
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

### Контекст і партіали

```go
// LookupPath повертає значення за крапковим шляхом від root (наприклад "title", "user.name").
// Root може бути map[string]any або реалізовувати Raw() any з мапою.
// Використовується згенерованим кодом для @root.xxx у партіалах, коли root приходить з іншого шаблону.
val := runtime.LookupPath(root, "title")
```

## Функції-хелпери

Прості хелпери (не блокові) повинні мати таку сигнатуру:

```go
func MyHelper(args []any) (any, error)
```

Аргументи **обчислюються компілятором** перед передачею; ви отримуєте вже обчислені значення. Контекст не передається — компілятор підставляє потрібні звертання.

### Доступ до аргументів

```go
func MyHelper(args []any) (any, error) {
    // Позиційні аргументи (вже обчислені)
    if len(args) == 0 {
        return nil, fmt.Errorf("missing argument")
    }
    firstArg := args[0]
    
    // Hash-аргументи (пари key=value)
    hash, ok := runtime.HashArg(args)
    if ok {
        value := hash["key"]
    }
    
    return result, nil
}
```

### Блокові хелпери

Блокові хелпери викликаються компілятором з одним аргументом: повним зрізом `args`, останнім елементом якого є опції блоку. Використовуйте сигнатуру `func(args []any) error` та витягуйте опції через `runtime.GetBlockOptions(args)`:

```go
func MyBlockHelper(args []any) error {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return fmt.Errorf("block helper did not receive BlockOptions")
    }
    // Рендер основного блоку (opts.Fn(w) потребує w з контексту виклику)
    if opts.Fn != nil {
        if err := opts.Fn(w); err != nil {
            return err
        }
    }
    // Рендер блоку inverse/else
    if opts.Inverse != nil {
        if err := opts.Inverse(w); err != nil {
            return err
        }
    }
    return nil
}
```

`BlockOptions` містить:

```go
type BlockOptions struct {
    Fn      func(io.Writer) error  // тіло основного блоку
    Inverse func(io.Writer) error   // тіло блоку else
}
```

У runtime також визначено `BlockHelper` як `func(args []any, options BlockOptions) error` для ручного виклику блокового хелпера з двома аргументами. При виклику зі згенерованого коду передається лише `args` (опції — останній елемент).

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

Hash-аргументи передаються як `runtime.Hash`:

```go
type Hash map[string]any
```

### Опції блоку

```go
type BlockOptions struct {
    Fn      func(io.Writer) error
    Inverse func(io.Writer) error
}
```

## Обробка помилок

Усі функції рендеру повертають помилки. Типові ситуації:

- Відсутній шаблон або партіал (помилка компіляції)
- Відсутній хелпер (помилка компіляції)
- Помилки рантайму в хелперах
- Невірні типи даних
- Блоковий хелпер не отримав BlockOptions

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

### Блоковий хелпер

```go
func IfHelper(args []any) error {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return fmt.Errorf("if: no block options")
    }
    if len(args) < 1 {
        return fmt.Errorf("if requires a condition")
    }
    condition := args[0]
    if runtime.IsTruthy(condition) {
        if opts.Fn != nil {
            return opts.Fn(w) // w — writer виводу шаблону (у контексті згенерованого коду)
        }
    } else if opts.Inverse != nil {
        return opts.Inverse(w)
    }
    return nil
}
```

Примітка: вбудовані `if`/`unless`/`each`/`with` реалізовані компілятором; приклад вище ілюструє рантайм API для власних блокових хелперів. Коли компілятор викликає блоковий хелпер, він викликає `helper(args)`; writer `w` є у контексті згенерованої функції рендеру. Власні хелпери, що викликаються зі згенерованого коду та мають рендерити блок, повинні отримувати або захоплювати writer (наприклад через адаптер).

## Див. також

- [Скомпільований файл шаблонів](compiled-templates.md) — що генерує компілятор (типи контексту, RenderXxx, FromMap)
- [Синтаксис Handlebars](syntax.md) — вирази та блоки
- [Вбудовані хелпери](helpers.md) — доступні хелпери та реєстрація власних
