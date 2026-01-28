# Згенерований bootstrap (деталі реалізації)

Цей документ описує **блок bootstrap**, який генерує `hbc -bootstrap`: який код випромінюється та який **інтерфейс для розробника** він надає.

## Коли генерується bootstrap

Якщо запустити компілятор з `-bootstrap`:

```bash
hbc -in . -out ./templates_gen.go -pkg templates -bootstrap
```

у згенерованому файлі після функцій шаблонів з’являється додатковий блок коду. Пакет також отримує імпорти `github.com/andriyg76/go-hbars/pkg/renderer` та `github.com/andriyg76/go-hbars/pkg/sitegen`.

## Згенерований код (що містить файл)

### 1. `rendererFuncs`

Мапа від імені шаблону (рядок) до публічної функції рендеру:

```go
// rendererFuncs maps template names to render functions.
var rendererFuncs = map[string]func(io.Writer, any) error{
    "main":   RenderMain,
    "header": RenderHeader,
    "footer": RenderFooter,
}
```

Ключі — ті самі імена шаблонів, що й у ваших `.hbs` файлах (наприклад `main`, `header`, `blog/post`). Значення — згенеровані функції `RenderXxx`.

### 2. `NewRenderer()`

Повертає рендерер, який можна використовувати з `sitegen.NewProcessor` або `sitegen.NewServer`:

```go
// NewRenderer returns a ready-to-use template renderer.
// This renderer can be used with sitegen.NewProcessor or sitegen.NewServer.
func NewRenderer() renderer.TemplateRenderer {
    return sitegen.NewRendererFromFunctions(rendererFuncs)
}
```

### 3. `NewQuickProcessor()`

Створює процесор із конфігурацією за замовчуванням для генерації статичного сайту:

```go
// NewQuickProcessor creates a processor with default configuration.
// Use this for quick static site generation.
func NewQuickProcessor() (*sitegen.Processor, error) {
    config := sitegen.DefaultConfig()
    renderer := NewRenderer()
    return sitegen.NewProcessor(config, renderer)
}
```

### 4. `NewQuickServer()`

Створює сервер із конфігурацією за замовчуванням для розробки:

```go
// NewQuickServer creates a server with default configuration.
// Use this for quick development server setup.
func NewQuickServer() (*sitegen.Server, error) {
    config := sitegen.DefaultConfig()
    renderer := NewRenderer()
    return sitegen.NewServer(config, renderer)
}
```

## Інтерфейс для розробника

### `renderer.TemplateRenderer`

Означено в `github.com/andriyg76/go-hbars/pkg/renderer`:

```go
type TemplateRenderer interface {
    Render(templateName string, w io.Writer, data any) error
}
```

- **`NewRenderer()`** повертає цей інтерфейс.
- Процесор і сервер приймають `renderer.TemplateRenderer`; вони викликають `Render(templateName, w, data)` для рендеру сторінки за іменем шаблону (наприклад `"main"`, `"blog/post"`).

Ваш код не реалізує цей інтерфейс напряму; згенерований `NewRenderer()` повертає реалізацію на основі `rendererFuncs`.

### `*sitegen.Processor`

Повертається **`NewQuickProcessor()`**:

- **`Process() error`** — обробляє всі файли даних у `Config().DataPath`, записує HTML у `Config().OutputPath`.
- **`Config() *sitegen.Config`** — встановити `DataPath`, `OutputPath`, `SharedPath`, `RootPath` під свою структуру.

Використовуйте для генерації статичного сайту: файли даних з `_page.template` та `_page.output` рендеряться та записуються на диск.

### `*sitegen.Server`

Повертається **`NewQuickServer()`**:

- **`Start() error`** — запускає HTTP-сервер (блокуючий виклик).
- **`Config() *sitegen.Config`** — встановити `Addr`, `DataPath`, `SharedPath`, `RootPath` для сервера.

Використовуйте для розробки: URL зіставляються з файлами даних (наприклад `/` → `data/index.json`), сторінки рендеряться на вимогу.

### `*sitegen.Config`

Спільна конфігурація для процесора та сервера (через `Config()`):

| Поле         | Значення |
|--------------|----------|
| `RootPath`   | Корінь проекту (за замовчуванням — поточна робоча директорія). |
| `DataPath`   | Директорія з файлами даних JSON/YAML/TOML (наприклад `"data"`). |
| `SharedPath` | Директорія зі спільними даними, що підмешуються в кожну сторінку (наприклад `"shared"`). |
| `OutputPath` | (лише процесор) Директорія для згенерованого HTML (наприклад `"pages"`). |
| `Addr`       | (лише сервер) Адреса прослуховування (наприклад `":8080"`). |

## Типове використання

```go
import templates "myapp/processor/templates"

// Статичний сайт
proc, err := templates.NewQuickProcessor()
if err != nil { log.Fatal(err) }
proc.Config().DataPath = "data"
proc.Config().OutputPath = "pages"
if err := proc.Process(); err != nil { log.Fatal(err) }

// Або HTTP-сервер
srv, err := templates.NewQuickServer()
if err != nil { log.Fatal(err) }
srv.Config().DataPath = "data"
srv.Config().Addr = ":8080"
log.Fatal(srv.Start())
```

Bootstrap використовує лише публічні пакети (`pkg/renderer`, `pkg/sitegen`), тому ваш модуль може залежати від go-hbars з GitHub без імпорту internal-пакетів.
