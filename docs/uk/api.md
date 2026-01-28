# Довідник API шаблонів

Цей документ описує рантайм API для роботи зі скомпільованими Handlebars-шаблонами.

## Базове використання

Після компіляції шаблонів через `hbc` ви отримуєте згенеровані функції для кожного шаблону:

```go
import "github.com/your/project/templates"

// Рендер у writer
var b strings.Builder
if err := templates.RenderMain(&b, data); err != nil {
    // обробити помилку
}
out := b.String()

// Або використати обгортку для рядка
out, err := templates.RenderMainString(data)
```

## Згенеровані функції

Для кожного файлу шаблону (наприклад `main.hbs`) компілятор генерує:

1. **Внутрішня функція рендеру**: `renderMain(ctx *runtime.Context, w io.Writer) error`
2. **Публічна функція рендеру**: `RenderMain(w io.Writer, data any) error`
3. **Обгортка для рядка**: `RenderMainString(data any) (string, error)`

## Пакет runtime

Пакет `runtime` надає базову функціональність виконання шаблонів.

### Контекст

```go
// NewContext створює новий контекст рендеру
ctx := runtime.NewContext(data)

// WithData створює дочірній контекст з новими даними
childCtx := ctx.WithData(newData)

// WithScope створює дочірній контекст з новими даними та опційними locals/data vars
childCtx := ctx.WithScope(data, locals, dataVars)
```

### Розв’язання шляхів

```go
// ResolvePath шукає точковий шлях у поточному контексті
value, ok := runtime.ResolvePath(ctx, "user.name")

// ResolvePathParsed розв’язує попередньо розпарсений вираз шляху
value, ok := runtime.ResolvePathParsed(ctx, parsedPath)
```

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
// EvalArg обчислює вираз аргументу
value := runtime.EvalArg(ctx, runtime.ArgPath, "user.name")
value := runtime.EvalArg(ctx, runtime.ArgString, "literal")
value := runtime.EvalArg(ctx, runtime.ArgNumber, "42")

// HashArg витягує hash-аргументи з аргументів хелпера
hash, ok := runtime.HashArg(args)

// GetBlockOptions витягує опції блоку з аргументів хелпера
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

## Функції-хелпери

Хелпери повинні мати таку сигнатуру:

```go
func MyHelper(ctx *runtime.Context, args []any) (any, error)
```

### Доступ до аргументів

```go
func MyHelper(ctx *runtime.Context, args []any) (any, error) {
    // Позиційні аргументи
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

```go
func MyBlockHelper(ctx *runtime.Context, args []any) (any, error) {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        // Не використовується як блок
        return "default", nil
    }
    
    // Рендер основного блоку
    if opts.Fn != nil {
        var b strings.Builder
        if err := opts.Fn(ctx, &b); err != nil {
            return nil, err
        }
        return b.String(), nil
    }
    
    // Рендер блоку inverse/else
    if opts.Inverse != nil {
        var b strings.Builder
        if err := opts.Inverse(ctx, &b); err != nil {
            return nil, err
        }
        return b.String(), nil
    }
    
    return "", nil
}
```

## Партіали

Партіали автоматично реєструються в згенерованому коді:

```go
// Доступ до мапи partials (внутрішнє)
partials["header"](ctx, w)

// Партіали використовуються в шаблонах через {{> header}}
```

## Типи даних

### Дані контексту

Дані контексту можуть бути будь-яким Go-типом:
- Мапи (`map[string]any`)
- Структури (з експортованими полями або JSON-тегами)
- Зрізи/масиви
- Примітиви (string, int, float, bool тощо)

### Hash-аргументи

Hash-аргументи передаються як `runtime.Hash`:

```go
type Hash map[string]any
```

### Опції блоку

```go
type BlockOptions struct {
    Fn      func(*Context, io.Writer) error
    Inverse func(*Context, io.Writer) error
}
```

## Обробка помилок

Усі функції рендеру повертають помилки. Типові ситуації:

- Відсутній шаблон або партіал (помилка компіляції)
- Відсутній хелпер (помилка компіляції)
- Помилки рантайму в хелперах
- Невірні типи даних
- Помилки розв’язання шляху

Завжди перевіряйте помилки:

```go
out, err := templates.RenderMainString(data)
if err != nil {
    log.Fatal(err)
}
```

## Продуктивність

- Шаблони компілюються в Go-код, тому виконання швидке
- Немає парсингу шаблонів під час виконання
- Створення контексту легке
- Розв’язання шляхів використовує пошук рядків у рантаймі (`ResolvePath` / `ResolvePathValue`)

## Приклади

### Простий рендер шаблону

```go
data := map[string]any{
    "title": "Hello",
    "user": map[string]any{
        "name": "Alice",
    },
}

out, err := templates.RenderMainString(data)
```

### Власний хелпер

```go
func FormatCurrency(ctx *runtime.Context, args []any) (any, error) {
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
func IfHelper(ctx *runtime.Context, args []any) (any, error) {
    opts, ok := runtime.GetBlockOptions(args)
    if !ok {
        return nil, fmt.Errorf("if must be used as block")
    }
    
    condition := args[0]
    if runtime.IsTruthy(condition) {
        if opts.Fn != nil {
            var b strings.Builder
            if err := opts.Fn(ctx, &b); err != nil {
                return nil, err
            }
            return b.String(), nil
        }
    } else {
        if opts.Inverse != nil {
            var b strings.Builder
            if err := opts.Inverse(ctx, &b); err != nil {
                return nil, err
            }
            return b.String(), nil
        }
    }
    
    return "", nil
}
```
